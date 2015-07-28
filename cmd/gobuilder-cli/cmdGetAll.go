package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func getGetAllCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-all IMPORT_PATH PATH [VERSION]",
		Short: "Get the current ZIP files of that package and version and store them to PATH",
		Run:   cmdGetAll,
	}

	return cmd
}

func cmdGetAll(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		fmt.Printf("Please provide an import path (like 'github.com/Luzifer/gobuilder') to download and a storage path\n")
		os.Exit(1)
	}

	if len(args) == 2 {
		args = append(args, "master")
	}

	os.MkdirAll(args[1], 0755)

	if err := downloadBuildResult(args[0], args[2], "", args[1]); err != nil {
		switch err.(type) {
		case noDownloadFoundError:
			fmt.Printf("No downloads has been found for your request.\n")
		default:
			fmt.Printf("%s\n", err)
		}
		os.Exit(1)
	}
}
