package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

func getBuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "build",
		Short:  "Trigger a build for this repository",
		Run:    cmdBuild,
		PreRun: checkRepoPresent,
	}

	return cmd
}

func cmdBuild(cmd *cobra.Command, args []string) {
	resp, err := http.PostForm("https://gobuilder.me/api/v1/webhook/cli", url.Values{
		"repository": []string{config.Repo},
	})
	if err != nil {
		fmt.Printf("Was unable to communicate with GoBuilder, please try again.")
		os.Exit(1)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNoContent:
		fmt.Printf("You did not specify any respository to build.")
		os.Exit(1)
	case http.StatusNotAcceptable:
		fmt.Printf("GoBuilder rejected your repository. Please contact support at help@gobuilder.me")
		os.Exit(1)
	case http.StatusInternalServerError:
		fmt.Printf("Something went wrong on server side. Please try again.")
		os.Exit(1)
	case http.StatusOK:
		fmt.Printf("Your build has been queued.")
		os.Exit(0)
	}
}
