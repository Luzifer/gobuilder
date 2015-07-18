package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func getEncryptCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "encrypt [SECRET]",
		Short:  "Encrypt a secret for use in .gobuilder.yml",
		Run:    cmdEncrypt,
		PreRun: checkRepoPresent,
	}

	return cmd
}

func cmdEncrypt(cmd *cobra.Command, args []string) {
	var secret string
	if len(args) == 1 {
		secret = args[0]
	} else {
		fmt.Printf("Go ahead and type your secret ...\n")
		s, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Printf("Unable to read from STDIN")
			os.Exit(1)
		}
		secret = string(s)
	}

	if len(strings.TrimSpace(secret)) == 0 {
		fmt.Printf("Please provide the secret to encrypt via argument or STDIN.")
		os.Exit(1)
	}

	resp, err := http.PostForm(fmt.Sprintf("https://gobuilder.me/api/v1/%s/encrypt", config.Repo), url.Values{
		"secret": []string{secret},
	})
	if err != nil {
		fmt.Printf("Was unable to communicate with GoBuilder, please try again.")
		os.Exit(1)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusInternalServerError:
		fmt.Printf("Something went wrong on server side. Please try again.")
		os.Exit(1)
	case http.StatusOK:
		io.Copy(os.Stdout, resp.Body)
		fmt.Printf("\n")
		os.Exit(0)
	}
}
