# About this project

[![Jenkins Build Status](http://badge.luzifer.io/v1/badge?title=Jenkins&text=Build%20Status)](http://jenkins.hub.luzifer.io/job/gobuilder-images/)
[![License: Apache 2.0](http://badge.luzifer.io/v1/badge?color=5d79b5&title=license&text=Apache%202.0)](http://www.apache.org/licenses/LICENSE-2.0)

This project originated when I was dissatisfied with how other automatic [Go][golang] building systems were working. I wanted a simple system building my Go code and after that packaging the results with some assets found in the same repository. Also I wanted to have a [webhook][gh-webhook] URL to trigger the builds automatically when my code is pushed to [GitHub][gh] or [BitBucket][bitbucket].

At the time I wrote this build system no existing system did satisfy my requirements or they were just not reliable enough for my needs. From that the project [Luzifer/gobuilder][gobuilder] was born.

# How To Use

The most simple usage is to take the package of your code (for example `github.com/Luzifer/gobuilder`) and put it into the build box on the start page of my [GoBuilder][gob]. After you submit your package to the queue it may need some time to build (also depends on how full the queue currently is) and after this you just visit `https://gobuilder.me/[package]` to view and download the build results. Following my example you can use [https://gobuilder.me/github.com/Luzifer/gobuilder](https://gobuilder.me/github.com/Luzifer/gobuilder) to download the build results of my GoBuilder itself.

## Using a webhook

There are currently two kinds of webhooks supported for automatically building your projects as soon as you push new code to the `master` branch. Currently only GitHub and BitBucket are supported:

- GitHub - `https://gobuilder.me/api/v1/webhook/github`
- BitBucket - `https://gobuilder.me/api/v1/webhook/bitbucket`

You just put the URL into the webhook section of your repository configuration and your project will be built automatically.

## Using the `.gobuilder.yml` file

To configure some aspects of your build you will need to create a `.gobuilder.yml` file in your repository root. This file currently has these options:

- `build_matrix`: A map of `OS/ARCH` combinations to build; if you don't specify the `build_matrix` we will build windows, osx and linux for you (For a list of valid platforms see [configreader](/cmd/configreader/main.go#L15-L25))
  - `build_tags`: A list of build tags to use while building
  - `ldflags`: A list of ldflags to use while building (Please note: This feature is experimental and you might get unexpected effects!)
- `readme_file`: The markdown file to display on the repository page in the web frontend. (Defaults to `README.md`)
- `triggers`: A list of repositories to build after a successful build of your repository. This could be used to generate some CLI utilities sitting in subdirs of your repository.
- `artifacts`: In this option you can list assets to include into the zip file created from the build. For example if you have a file called `LICENSE` in the root of your repository and want this to get included into the build result you just add a item with the content `LICENSE` to this array.
- `version_file`: If you provide a file name to this option the hash of the compiled commit will get written in this file and added to the result ZIP file.
- `notify`: You can ping some services after a successful / failed build. The notification can be filtered only to get sent on specific events by providing a `filter` value with `success` or `error`. Currently these services are supported:
    - `dockerhub`: Fill the whole URL you got as a "Build Trigger" as the target.
    - `pushover`: Put your "User Key" into the target to receive notifications.

An example configuration file:

```yaml
---
build_matrix:
  windows:
  osx:              # You can use "osx" as an alias to "darwin"
  linux/amd64:
    build_tags:
      - nofoo
      - bar
  linux/386:
    build_tags:
      - foo
      - nobar
  general:          # tags / ldflags for "general" are used as a fallback
    ldflags:
      - "-x main.version 1.0.0"
readme_file: frontend/help.md
triggers:
  - github.com/Luzifer/gobuilder/cmd/starter
artifacts:
  - frontend/*
version_file: VERSION
notify:
  - type: dockerhub
    target: https://registry.hub.docker.com[...]d59f8a5ab895/
    filter: success
  - type: pushover
    target: W2HNyg7sCkvNH[...]B
```

## Code verification and signatures

Starting with version 1.15.0 GoBuilder supports verification of code signatures. This can be used to give users of your projects an additional bit of security if you direct them to GoBuilder for downloads. If you have signed tags the repository view for your project will get an additional button in the top right corner as soon as a signed label is selected by your user. By clicking on that button a message will be displayed stating whether your tag was successfully verified. Passing this test means the code was not altered while transferred between your computer and the GoBuilder build system.

If you want to get those verifications for your master branch you need to sign your commits using an GPG key. As long as this signature is valid GoBuilder will show the verification result.


[golang]: http://golang.org/
[gh-webhook]: https://developer.github.com/webhooks/
[gh]: https://github.com/
[bitbucket]: https://bitbucket.org/
[gobuilder]: https://github.com/luzifer/gobuilder
[gob]: https://gobuilder.me/
