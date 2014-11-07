package main

import (
	"fmt"
	"os"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/kr/beanstalk"
	"github.com/segmentio/go-loggly"
)

func orFail(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	log := loggly.New(os.Getenv("LOGGLY_TOKEN"))
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
		log.Error("Beanstalk-Connect", loggly.Message{
			"error": fmt.Sprintf("%v", err),
		})
		panic(err)
	}

	defer conn.Close()

	ts := beanstalk.NewTubeSet(conn, "gobuild.luzifer.io")
	for {
		fmt.Printf("Starting waitcycle...\n")
		id, body, err := ts.Reserve(10 * time.Second)
		if cerr, ok := err.(beanstalk.ConnError); ok && cerr.Err == beanstalk.ErrTimeout {
			fmt.Println("timed out")
			continue
		} else if err != nil {
			log.Error("Tube-Reserve", loggly.Message{
				"error": fmt.Sprintf("%v", err),
			})
			panic(err)
		}

		fmt.Printf("Beginning to process %s\n", body)

		cfg := &docker.Config{
			AttachStdin:  false,
			AttachStdout: true,
			AttachStderr: true,
			Image:        os.Getenv("BUILD_IMAGE"),
			Env: []string{
				fmt.Sprintf("GIT_URL=%s", body),
				"GIT_BRANCH=master",
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
			conn.Delete(id)
			log.Info("Finished build", loggly.Message{
				"repository": string(body),
			})
		} else {
			conn.Release(id, 1, 120*time.Second)
			log.Error("Failed build", loggly.Message{
				"repository":  body,
				"exit_status": status,
			})
		}

	}
}
