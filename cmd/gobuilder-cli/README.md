# GoBuilder / gobuilder-cli

This CLI tool emerged in GoBuilder v1.24.0 after we started to support encrypted values in `.gobuilder.yml` in v1.23.0.

## Usage

- For best results please provide an [import comment](https://golang.org/cmd/go/#hdr-Import_path_checking) in your package line of the current project you are executing gobuilder-cli for.
- Remember to push your changes before running `gobuilder-cli build`

```bash
# gobuilder-cli help
Remote-Control for GoBuilder.me

Usage:
  gobuilder-cli [command]

Available Commands:
  build       Trigger a build for this repository
  encrypt     Encrypt a secret for use in .gobuilder.yml
  help        Help about any command

Flags:
  -h, --help=false: help for gobuilder-cli
      --repo="": Repository to work with


Use "gobuilder-cli [command] --help" for more information about a command.
```
