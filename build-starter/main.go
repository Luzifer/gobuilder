package main

import (
	"fmt"
	"os"
	"time"

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

	dockerClient, err := docker.NewClient("unix:///var/run/docker.sock")
	orFail(err)

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

		if build(string(body)) {
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

func build(repo string) bool {
	fmt.Printf("Beginning to process %s\n", repo)

	cfg := &docker.Config{
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Image:        os.Getenv("BUILD_IMAGE"),
		Env: []string{
			fmt.Sprintf("GIT_URL=%s", repo),
		},
	}
	container, err := dockerClient.CreateContainer(docker.CreateContainerOptions{
		Config: cfg,
	})
	orFail(err)
	err = dockerClient.StartContainer(container.ID, &docker.HostConfig{
		Binds:        []string{},
		Privileged:   false,
		PortBindings: make(map[docker.Port][]docker.PortBinding),
	})
	orFail(err)
	status, err := dockerClient.WaitContainer(container.ID)
	orFail(err)

	if status == 0 {
		return true
	}
	return false
}
