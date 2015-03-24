package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Luzifer/gobuilder/buildconfig"
	"gopkg.in/codegangsta/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "configreader"
	app.Usage = "Parses .gobuilder.yml and enables access through bash build script"

	app.Commands = []cli.Command{
		{
			Name:   "read",
			Usage:  "reads a value from the .gobuilder.yml file and prints it",
			Action: readConfig,
		},
		{
			Name:   "checkEmpty",
			Usage:  "checks whether the value is empty",
			Action: checkEmptyVar,
		},
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "./.gobuilder.yml",
			Usage: "Path of the configuration yaml file",
		},
	}
	app.Run(os.Args)
}

func readConfig(context *cli.Context) {
	handleCommand(context, filterPrint)
}

func checkEmptyVar(context *cli.Context) {
	handleCommand(context, filterCheckEmpty)
}

func filterPrint(varContent string) {
	fmt.Println(varContent)
	os.Exit(0)
}

func filterCheckEmpty(varContent string) {
	if len(strings.TrimSpace(varContent)) > 0 {
		fmt.Println("false")
		os.Exit(1)
	}
	fmt.Println("true")
	os.Exit(0)
}

func handleCommand(context *cli.Context, filter func(string)) {
	cfg, err := buildconfig.LoadFromFile(context.GlobalString("config"))
	if err != nil {
		fmt.Printf("Unable to open / parse .gobuilder.yml file: \"%s\"\n", context.GlobalString("config"))
		os.Exit(1)
	}

	if len(context.Args()) != 1 {
		fmt.Println("Please provide a key to retrieve")
		os.Exit(1)
	}

	key := context.Args()[0]

	switch key {
	case "readme_file":
		filter(cfg.ReadmeFile)
	case "triggers":
		filter(strings.Join(cfg.Triggers, "\n"))
	case "artifacts":
		filter(strings.Join(cfg.Artifacts, "\n"))
	case "version_file":
		filter(cfg.VersionFile)
	}
}
