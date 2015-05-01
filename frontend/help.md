# About this project

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

- `readme_file`: The markdown file to display on the repository page in the web frontend. (Defaults to `README.md`)
- `triggers`: A list of repositories to build after a successful build of your repository. This could be used to generate some CLI utilities sitting in subdirs of your repository.
- `artifacts`: In this option you can list assets to include into the zip file created from the build. For example if you have a file called `LICENSE` in the root of your repository and want this to get included into the build result you just add a item with the content `LICENSE` to this array.
- `version_file`: If you provide a file name to this option the hash of the compiled commit will get written in this file and added to the result ZIP file.
- `notify`: You can ping some services after a successful / failed build. The notification can be filtered only to get sent on specific events by providing a `filter` value with `success` or `error`. Currently these services are supported:
    - `dockerhub`: Fill the whole URL you got as a "Build Trigger" as the target.

An example configuration file:

```
---
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
```


[golang]: http://golang.org/
[gh-webhook]: https://developer.github.com/webhooks/
[gh]: https://github.com/
[bitbucket]: https://bitbucket.org/
[gobuilder]: https://github.com/luzifer/gobuilder
[gob]: https://gobuilder.me/
