package main // import "github.com/Luzifer/gobuilder/cmd/asset-sync"

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/Luzifer/gobuilder/buildconfig"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Missing arguments.")
		os.Exit(1)
	}

	fromPathPrefix := os.Args[1]
	toPathPrefix := os.Args[2]

	cfg, err := buildconfig.LoadFromFile(".gobuilder.yml")
	if err != nil {
		fmt.Printf("Unable to open / parse .gobuilder.yml file.\n")
		os.Exit(1)
	}

	i := 0
	for to, from := range cfg.Artifacts {
		i++
		fromPath := path.Join(fromPathPrefix, from)
		toPath := path.Join(toPathPrefix, to)
		fmt.Printf("(%d/%d) %s => %s...\n", i, len(cfg.Artifacts), fromPath, toPath)

		if exists(toPath) {
			fmt.Printf("  ERR: Target path '%s' already exists.\n", toPath)
			continue
		}

		if !exists(fromPath) {
			fmt.Printf("  ERR: Source path '%s' does not exist.\n", fromPath)
		}

		src, err := os.Stat(fromPath)
		if err != nil {
			fmt.Printf("  ERR: Unable to read file stats\n")
			continue
		}

		if src.IsDir() {
			if err := copyDir(fromPath, toPath); err != nil {
				fmt.Printf("  ERR: %s\n", err)
				continue
			}
		} else {
			if err := copyFile(fromPath, toPath); err != nil {
				fmt.Printf("  ERR: %s\n", err)
				continue
			}
		}

		fmt.Printf("  OK\n")
	}
}

func exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func copyFile(source string, dest string) (err error) {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourcefile.Close()

	err = os.MkdirAll(path.Dir(dest), 0755)
	if err != nil {
		return err
	}

	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer destfile.Close()

	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err == nil {
			err = os.Chmod(dest, sourceinfo.Mode())
		}
	}

	return nil
}

func copyDir(source string, dest string) (err error) {
	// get properties of source dir
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	// create dest dir
	if err := os.MkdirAll(dest, sourceinfo.Mode()); err != nil {
		return err
	}

	directory, _ := os.Open(source)
	objects, err := directory.Readdir(-1)

	for _, obj := range objects {
		sourcefilepointer := path.Join(source, obj.Name())
		destinationfilepointer := path.Join(dest, obj.Name())

		if obj.IsDir() {
			// create sub-directories - recursively
			if err := copyDir(sourcefilepointer, destinationfilepointer); err != nil {
				return err
			}
		} else {
			// perform copy
			if err := copyFile(sourcefilepointer, destinationfilepointer); err != nil {
				return err
			}
		}
	}
	return nil
}
