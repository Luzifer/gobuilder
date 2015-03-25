package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"

	"github.com/Luzifer/gobuilder/builddbCreator"
	"github.com/Luzifer/gobuilder/buildjob"
	"github.com/fsouza/go-dockerclient"
	"github.com/kr/beanstalk"
	"github.com/segmentio/go-loggly"
)

var dockerClient *docker.Client
var log *loggly.Client
var s3Bucket *s3.Bucket

const maxJobRetries int = 5

func orFail(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	log = loggly.New(os.Getenv("LOGGLY_TOKEN"))
	log.Tag("GoBuild-Starter")

	awsAuth, err := aws.EnvAuth()
	orFail(err)
	s3Bucket = s3.New(awsAuth, aws.EUWest).Bucket("gobuild.luzifer.io")

	dockerClientTmp, err := docker.NewClient("unix:///var/run/docker.sock")
	orFail(err)
	dockerClient = dockerClientTmp

	err = dockerClient.PullImage(docker.PullImageOptions{
		Repository: os.Getenv("BUILD_IMAGE"),
		Tag:        "latest",
	}, docker.AuthConfiguration{})
	orFail(err)

	conn, err := beanstalk.Dial("tcp", os.Getenv("BEANSTALK_ADDR"))
	if err != nil {
		_ = log.Error("Beanstalk-Connect", loggly.Message{
			"error": fmt.Sprintf("%v", err),
		})
		panic(err)
	}

	defer func() {
		_ = conn.Close()
	}()

	waitForBuild(conn)
}

func waitForBuild(conn *beanstalk.Conn) {
	ts := beanstalk.NewTubeSet(conn, "gobuild.luzifer.io")
	for {
		fmt.Printf("Starting waitcycle...\n")
		id, body, err := ts.Reserve(10 * time.Second)
		if cerr, ok := err.(beanstalk.ConnError); ok && cerr.Err == beanstalk.ErrTimeout {
			fmt.Println("timed out")
			continue
		} else if err != nil {
			_ = log.Error("Tube-Reserve", loggly.Message{
				"error": fmt.Sprintf("%v", err),
			})
			panic(err)
		}

		job, err := buildjob.FromBytes(body)
		if err != nil {
			// there was a job we could not parse throw it away and continue
			_ = conn.Delete(id)
			continue
		}
		repo := job.Repository

		tmpDir, err := ioutil.TempDir("", "gobuild")
		orFail(err)
		buildResult, triggerUpload := build(repo, tmpDir)
		_ = conn.Delete(id)

		if triggerUpload {
			orFail(builddbCreator.GenerateBuildDB(tmpDir))
			uploadAssets(repo, tmpDir)
		}

		if buildResult {
			orFail(s3Bucket.Put(fmt.Sprintf("%s/build.status", repo), []byte("finished"), "text/plain", s3.PublicRead))
			_ = os.RemoveAll(tmpDir)

			_ = log.Info("Finished build", loggly.Message{
				"repository": repo,
			})
		} else {
			orFail(s3Bucket.Put(fmt.Sprintf("%s/build.status", repo), []byte("queued"), "text/plain", s3.PublicRead))
			_ = log.Error("Failed build", loggly.Message{
				"repository": repo,
			})

			// Try to build the job only 5 times not to clutter the queue
			if job.NumberOfExecutions < maxJobRetries-1 {
				job.NumberOfExecutions = job.NumberOfExecutions + 1
				queueEntry, err := job.ToByte()
				orFail(err)

				_, err = conn.Put([]byte(queueEntry), 1, 0, 900*time.Second)
				orFail(err)
			} else {
				_ = log.Error("Finally failed build", loggly.Message{
					"repository":         repo,
					"numberOfBuildTries": maxJobRetries,
				})
				orFail(s3Bucket.Put(fmt.Sprintf("%s/build.status", repo), []byte("failed"), "text/plain", s3.PublicRead))
			}
		}

	}
}

func build(repo, tmpDir string) (bool, bool) {
	fmt.Printf("Beginning to process %s\n", repo)

	cfg := &docker.Config{
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Image:        os.Getenv("BUILD_IMAGE"),
		Env: []string{
			fmt.Sprintf("REPO=%s", repo),
		},
	}
	container, err := dockerClient.CreateContainer(docker.CreateContainerOptions{
		Config: cfg,
	})
	orFail(err)
	err = dockerClient.StartContainer(container.ID, &docker.HostConfig{
		Binds:        []string{fmt.Sprintf("%s:/artifacts", tmpDir)},
		Privileged:   false,
		PortBindings: make(map[docker.Port][]docker.PortBinding),
	})
	orFail(err)

	orFail(s3Bucket.Put(fmt.Sprintf("%s/build.status", repo), []byte("building"), "text/plain", s3.PublicRead))
	status, err := dockerClient.WaitContainer(container.ID)
	orFail(err)

	logFile, err := os.Create(fmt.Sprintf("%s/build.log", tmpDir))
	orFail(err)
	defer func() {
		_ = logFile.Close()
	}()

	err = dockerClient.Logs(docker.LogsOptions{
		Container:    container.ID,
		Stdout:       true,
		OutputStream: logFile,
	})
	orFail(err)

	if status == 0 {
		return true, true
	} else if status == 130 {
		// Special case: Build was aborted due to redundant build request
		return true, false
	}
	return false, false
}

func uploadAssets(repo, tmpDir string) {
	fmt.Printf("ASSETS\n")

	assets, err := ioutil.ReadDir(tmpDir)
	orFail(err)
	fmt.Printf("%+v", assets)
	for _, f := range assets {
		if f.IsDir() {
			// Some repos are creating directories. Don't know why. Ignore them.
			continue
		}
		if strings.HasPrefix(f.Name(), ".") {
			// Dotfiles are used to transport metadata from the container
			continue
		}
		fmt.Printf("Uploading asset %s...\n", f.Name())
		originalPath := fmt.Sprintf("%s/%s", tmpDir, f.Name())
		path := fmt.Sprintf("%s/%s", repo, f.Name())
		fileContent, err := ioutil.ReadFile(originalPath)
		orFail(err)

		err = s3Bucket.Put(path, fileContent, "", s3.PublicRead)
		orFail(err)
	}
}
