package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Luzifer/gobuilder/buildconfig"
	"gopkg.in/codegangsta/cli.v1"
)

var (
	defaultArchs   = []string{"amd64", "386", "arm"}
	defaultOSs     = []string{"windows", "linux", "darwin"}
	validPlatForms = []string{
		"darwin/386", "darwin/amd64",
		"dragonfly/386", "dragonfly/amd64",
		"freebsd/386", "freebsd/amd64", "freebsd/arm",
		"linux/386", "linux/amd64", "linux/arm",
		"nacl/386", "nacl/amd64p32", "nacl/arm",
		"netbsd/386", "netbsd/amd64", "netbsd/arm",
		"openbsd/386", "openbsd/amd64",
		"plan9/386", "plan9/amd64",
		"solaris/amd64",
		"windows/386", "windows/amd64",
	}
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
		fmt.Printf("Unable to open / parse .gobuilder.yml file: \"%s\" - %s\n", context.GlobalString("config"), err)
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
	case "version_file":
		filter(cfg.VersionFile)
	case "arch_matrix":
		filter(strings.Join(buildArchList(cfg), " "))
	case "build_tags":
		filter(getBuildTags(cfg))
	case "ld_flags":
		filter(getLDFlags(cfg))
	}
}

func getBuildTags(cfg *buildconfig.BuildConfig) string {
	selectors := []string{
		fmt.Sprintf("%s/%s", os.Getenv("GOOS"), os.Getenv("GOARCH")),
		os.Getenv("GOOS"),
		"general",
	}

	for _, s := range selectors {
		if conf, ok := cfg.BuildMatrix[s]; ok {
			if len(conf.Tags) > 0 {
				return strings.Join(conf.Tags, " ")
			}
		}
	}

	return ""
}

func getLDFlags(cfg *buildconfig.BuildConfig) string {
	selectors := []string{
		fmt.Sprintf("%s/%s", os.Getenv("GOOS"), os.Getenv("GOARCH")),
		os.Getenv("GOOS"),
		"general",
	}

	for _, s := range selectors {
		if conf, ok := cfg.BuildMatrix[s]; ok {
			if len(conf.LDFlags) > 0 {
				return strings.Join(conf.LDFlags, " ")
			}
		}
	}

	return ""
}

func buildArchList(cfg *buildconfig.BuildConfig) []string {
	archs := []string{}
	for k := range cfg.BuildMatrix {

		// Some special handlings
		switch k {
		case "osx":
			k = "darwin"
		case "all":
			archs = validPlatForms
		case "general":
			continue
		}

		if strings.Contains(k, "/") {
			archs = append(archs, k)
		} else {
			for _, a := range defaultArchs {
				archs = append(archs, fmt.Sprintf("%s/%s", k, a))
			}
		}
	}

	if len(archs) == 0 {
		for _, o := range defaultOSs {
			for _, a := range defaultArchs {
				archs = append(archs, fmt.Sprintf("%s/%s", o, a))
			}
		}
	}

	out := []string{}
	for _, v := range archs {
		if sliceContains(validPlatForms, v) && !sliceContains(out, v) {
			out = append(out, v)
		}
	}

	return out
}

func sliceContains(slice []string, match string) bool {
	for _, v := range slice {
		if v == match {
			return true
		}
	}
	return false
}
