# About this project

This project originated when I was dissatisfied with how other automatic [Go][golang] building systems were working. I wanted a simple system building my Go code and after that packaging the results with some assets found in the same repository. Also I wanted to have a [webhook][gh-webhook] URL to trigger the builds automatically when my code is pushed to [GitHub][gh] or [BitBucket][bitbucket].

At the time I wrote this build system no existing system did satisfy my requirements or they were just not reliable enough for my needs. From that the project [Luzifer/gobuilder][gobuilder] was born.

# How To Use

The most simple usage is to take the package of your code (for example `github.com/Luzifer/gobuilder`) and put it into the build box on the start page of my [GoBuilder][gob]. After you submit your package to the queue it may need some time to build (also depends on how full the queue currently is) and after this you just visit `http://gobuild.luzifer.io/[package]` to view and download the build results. Following my example you can use [http://gobuild.luzifer.io/github.com/Luzifer/gobuilder](http://gobuild.luzifer.io/github.com/Luzifer/gobuilder) to download the build results of my GoBuilder itself.

## Using a webhook

There are currently two kinds of webhooks supported for automatically building your projects as soon as you push new code to the `master` branch. Currently only GitHub and BitBucket are supported:

- GitHub - `http://gobuild.luzifer.io/webhook/github`
- BitBucket - `http://gobuild.luzifer.io/webhook/bitbucket`

You just put the URL into the webhook section of your repository configuration and your project will be built automatically.

## Packaging assets into the zip-file

If you want to add files from your repository to the zip-file you just have to create a file named `.artifact_files` and add each asset as an extra line to that file. For example if you have a file called `LICENSE` in the root of your repository and want this to get included into the build result you just add a line with the content `LICENSE` to your `.artifact_files`.


[golang]: http://golang.org/
[gh-webhook]: https://developer.github.com/webhooks/
[gh]: https://github.com/
[bitbucket]: https://bitbucket.org/
[gobuilder]: https://github.com/luzifer/gobuilder
[gob]: http://gobuild.luzifer.io/
