package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"

	"github.com/fsouza/go-dockerclient"
	"github.com/kr/beanstalk"
	"github.com/segmentio/go-loggly"
)

var dockerClient *docker.Client
var log *loggly.Client

func orFail(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	log = loggly.New(os.Getenv("LOGGLY_TOKEN"))
	log.Tag("GoBuild-Starter")

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

		tmpDir, err := ioutil.TempDir("", "gobuild")
		orFail(err)
		buildResult, triggerUpload := build(string(body), tmpDir)

		if triggerUpload {
			uploadAssets(string(body), tmpDir)
		}

		if buildResult {
			_ = os.RemoveAll(tmpDir)

			_ = conn.Delete(id)
			_ = log.Info("Finished build", loggly.Message{
				"repository": string(body),
			})
		} else {
			_ = conn.Release(id, 1, 120*time.Second)
			_ = log.Error("Failed build", loggly.Message{
				"repository": string(body),
			})
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
	} else if status == 256 {
		// Special case: Build was aborted due to redundant build request
		return true, false
	}
	return false, false
}

func uploadAssets(repo, tmpDir string) {
	fmt.Printf("ASSETS\n")
	awsAuth, err := aws.EnvAuth()
	orFail(err)
	s3Bucket := s3.New(awsAuth, aws.EUWest).Bucket("gobuild.luzifer.io")

	assets, err := ioutil.ReadDir(tmpDir)
	orFail(err)
	fmt.Printf("%+v", assets)
	for _, f := range assets {
		fmt.Printf("Uploading asset %s...\n", f.Name())
		originalPath := fmt.Sprintf("%s/%s", tmpDir, f.Name())
		path := fmt.Sprintf("%s/%s", repo, f.Name())
		fileContent, err := ioutil.ReadFile(originalPath)
		orFail(err)

		err = s3Bucket.Put(path, fileContent, "", s3.PublicRead)
		orFail(err)
	}
}
