package main // import "github.com/Luzifer/gobuilder/cmd/gobuilder-cli"

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var config = struct {
	Repo  string
	Debug bool
}{
	Debug: false,
}

func init() {
	config.Repo = getRepo()
}

func main() {
	app := &cobra.Command{
		Use:   "gobuilder-cli",
		Short: "Remote-Control for GoBuilder.me",
	}

	app.PersistentFlags().StringVar(&config.Repo, "repo", config.Repo, "Repository to work with")

	app.AddCommand(
		getBuildCommand(),
		getEncryptCommand(),
		getGetCommand(),
		getGetAllCommand(),
	)

	app.Execute()
}

func getRepo() string {
	// Preferred: We have an import comment
	i, err := exec.Command("go", "list", "-e", "-f", "{{.ImportComment}}").Output()
	if err != nil && config.Debug {
		log.Printf("Error while reading ImportComment: %s\n", err)
	}
	if len(strings.TrimSpace(string(i))) > 0 {
		return strings.TrimSpace(string(i))
	}

	// Maybe other relieable methods?

	return ""
}

func checkRepoPresent(cmd *cobra.Command, args []string) {
	if config.Repo == "" {
		fmt.Printf("Could not determine your current project. Please set an import comment or specify --repo.\nSee https://golang.org/cmd/go/#hdr-Import_path_checking\n")
		os.Exit(1)
	}
}
