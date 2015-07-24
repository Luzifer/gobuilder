package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

func getGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get IMPORT_PATH [VERSION]",
		Short: "Get the current ZIP file of that package and version for your OS and ARCH",
		Run:   cmdGet,
	}

	return cmd
}

func cmdGet(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Printf("Please provide an import path (like 'github.com/Luzifer/gobuilder') to download\n")
		os.Exit(1)
	}

	if len(args) == 1 {
		args = append(args, "master")
	}

	search := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	if err := downloadBuildResult(args[0], args[1], search, "./"); err != nil {
		switch err.(type) {
		case noDownloadFoundError:
			fmt.Printf("No download has been found for your request.\n")
		default:
			fmt.Printf("%s\n", err)
		}

		os.Exit(1)
	}
}
