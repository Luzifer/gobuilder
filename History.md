# 1.37.0 / 2017-06-21

  * Remove all repository watches

# 1.36.2 / 2017-04-05

  * Enable /root mount
  * Allow mounting docker sock

# 1.36.1 / 2017-04-05

  * Change mount path

# 1.36.0 / 2017-04-04

  * Allow tmpdir to be configured
  * Dockerize starter

# 1.35.0 / 2016-10-23

  * Fix: HostConfig now needs to be passed at create
  * Update go-dockerclient to work with docker 1.12
  * Update vault2env to v0.5.0
  * Remove all but 100 log entries
  * Update to go1.7 builder

# 1.34.6 / 2016-05-31

  * Print go version into build log
  * Enabled travis builds for build-image

# 1.34.5 / 2016-05-28

  * Fix: Do not mess around with docker auth info

# 1.34.4 / 2016-05-27

  * Use better readable version info

# 1.34.3 / 2016-05-27

  * Updated dockerclient and dependent libraries
  * Removing captain as they broke their repo

# 1.34.2 / 2016-05-14

  * Fix: Vendored libs for commands

1.34.1 / 2016-04-16
==================

  * Fix: Do not include blocked Repos in sitemap.xml

1.34.0 / 2016-04-16
==================

  * Enhancement: Add functionality to block repositories
  * Enhancement: Do not depend on godeps tool
  * Enhancement: Build static binaries instead of dynamic linked ones
  * Fix: Set right content types
  * Fix: Do not use errors package with format strings
  * Fix: Be a bit more verbose with sending errors
  * Fix: Fixed ".built\_tags not found" error
  * Switched to MailGun
  * Removed prometheus repos; they are not longer buildable on GoBuilder

1.33.2 / 2015-12-23
==================

  * Remove code used during installation
  * Added Godeps for starter to fix build

1.33.1 / 2015-12-23
==================

  * Fix: Do not download code before setting GOPATH

1.33.0 / 2015-12-23
==================

  * Enhancement: Support more vendoring systems without having to install additional tools

1.32.0 / 2015-12-23
==================

  * Enhancement: Check whether an entry is in the queue before adding it
  * Enhancement: Added a config setting to disable `go fmt` runs
  * Fix: Fail unbuildable jobs way faster
  * Fix: Limit graphics in markdown files to 100% width
  * Fix: Do not crash if there is no previous build

1.31.0 / 2015-09-26
==================

  * Added RepoWatch feature
  * Change to go1.5
  * Added sitemap.xml, disallowed indexing of logs
  * Updated go-update & fixed checksum verification
  * Use own download without that magic from go-update
  * Verify download before patching
  * Ensure the hash is not empty

1.30.0 / 2015-08-07
==================

  * Fix: Save binaries for all tags
  * Added experimental autoupdate code
  * Store binary as an addition to the ZIPs
  * Followed linter advices

1.29.0 / 2015-08-06
==================

  * Search-Engines: Improved robot instructions
  * Added @commit notation for builds
  * Implemented already-built API method
  * Use built-commits set to determine last build
  * Remove old storage type from redis
  * Add build commit to set instead of storing last commit
  * Prepared build-image for dynamic branches / historic commits
  * Added http logging
  * Require at least one slash to prevent builds like "grafana"
  * Only download, do not try to install before checking for Godeps

1.28.0 / 2015-07-30
==================

  * Fix: Set modes
  * Added scroll to line and range mark in log

1.27.0 / 2015-07-28
==================

  * Implemented own asset-sync instead of rsync
  * Implemented `gobuilder-cli get(-all)`
  * Added yaml hash database generation
  * Fix: Try reconnecting redis instead of err-loop

1.26.0 / 2015-07-22
==================

  * Unified starter logs; log repo on all messages
  * Set number of concurrent builds to number of CPU cores
  * Fix: Starter crashed on failed build

1.25.0 / 2015-07-21
==================

  * Log hostname in BuildStarter logs
  * Fix: Use gits own --short option for rev-parse
  * Added killswitch on update detection
  * Abort non-buildable builds
  * Added metrics collector

1.24.0 / 2015-07-18
==================

  * Added README hint for encrypted values
  * Implemented gobuilder-cli

1.23.0 / 2015-07-18
==================

  * Documented email notification
  * Added email notification
  * Added notification target encryption

1.22.0 / 2015-07-12
==================

  * Display build logs next to "Not yet ready"
  * Fix: Logging did shut down build-starter, decreased level
  * Fix: Clean flashes after display
  * Fix: Redirects; Use redirects & flashes for alert display
  * Renamed Build-Logs panel
  * Fix: Positioning of log success indicator

1.21.0 / 2015-07-12
==================

  * Fix: Logging was not initialized
  * Fix: CSS error showing wrong colors
  * Let workers time out faster
  * Added metadata to log list

1.20.1 / 2015-07-12
==================

  * Fix: Removed remaining os.Getenv

1.20.0 / 2015-07-12
==================

  * Replaced undocumented Getenv with CLI parameters
  * Implemented GitHub auth / hook adding
  * Fix: Do not simply end without response on GitHub hook
  * Fix: Do not increase fail-count when job failed of gobuilders fault
  * Changed design

1.19.0 / 2015-06-21
==================

  * Write build started status with expiry
  * Allow evaluation of bash variables in ldflags
  * Added backoff not to quit too fast if pull fails at docker not yet ready
  * Fix: With only general build target set no build was done
  * Fix: Empty vars caused the build to crash

1.18.0 / 2015-06-14
==================

  * Store more than only one log and display last 10 on repo page

1.17.0 / 2015-06-14
==================

  * Full refactoring of build-starter
  * Added badges to README
  * Added auth configuration for docker pull; added periodical refresh

1.16.0 / 2015-06-14
==================

  * Enable "all" build target to build on all
  * Limit building gobuilder for linux only
  * Fix: Trigger webhook builds only on push to master
  * Added documentation for `build_matrix`
  * Added build matrix with tags / ldflags

1.15.3 / 2015-06-14
==================

  * Fix: Wrong naming of pushover token
  * Don't fail on notification failure but log them
  * Do not store metadata about builds in S3
  * Cleaned UI, added hint how to sign tags / commits

1.15.2 / 2015-06-14
==================

  * Added support for lightweight tags
  * Display signature button in different colors depending on Warning
  * Fix: Typo

1.15.1 / 2015-06-13
==================

  * Added ownertrust for own key
  * Added some documentation about code signing in GoBuilder

1.15.0 / 2015-06-13
==================

  * Added signing of file checksums to enable users to verify downloads
  * Added checking of code signatures in tags / commits
  * Give back some link love to the repo itself
  * Added license file
  * Write build artifact into /tmp to prevent file collisions
  * Fix: Build button in repo view was broken
  * Do not write git hash when upload was not triggered

1.14.0 / 2015-05-09
==================

  * Added Pushover support for notifications
  * Implemented APIv1
  * Replaced Martini framework with Gorilla mux

1.13.0 / 2015-05-01
==================

  * Added locking for build jobs
  * Fix: Do not write build-duration for duplicate builds

1.12.0 / 2015-04-21
==================

  * Rephrased working queue length
  * Rephrased build duration
  * Store build duration and display it
  * Fix: Errors in error templates
  * Don't crash if .gobuilder.yml was not passed out of the container

1.11.3 / 2015-04-14
==================

  * Fix: Build button gave no feedback as of template field errors

1.11.2 / 2015-04-14
==================

  * Fix: Don't show duplicates in last build list

1.11.1 / 2015-04-14
==================

  * Fixed wrong worker count when client crashed

1.11.0 / 2015-04-12
==================

  * Added some metadata to front page
  * Store build-status in Redis instead of S3
  * Replaced beanstalk queue with Redis list
  * Show the assets above the readme

1.10.0 / 2015-04-07
==================

  * Added FreshDesk widget
  * Make new builds visible even if they are not ready
  * Fix: Streamlined log lines
  * Added check to prevent spamming with broken repos

1.9.0 / 2015-03-29
==================

  * Added platform icons to file view
  * Removed newlines from log
  * Fix: When build is skipped .gobuilder.yml is not available
  * Use log level Info for build start

1.8.2 / 2015-03-27
==================

  * Streamlined logging

1.8.1 / 2015-03-27
==================

  * Fix: Upload .gobuilder.yml in any case
  * Fix: Installing a go utility needs its sources

1.8.0 / 2015-03-27
==================

  * Added .gobuilder.yml feature and removed old artifact handling
  * Use the same log format as the build starter for the frontend
  * Added favicon as static asset

1.7.0 / 2015-03-26
==================

  * Replaced loggly with papertrail

1.6.0 / 2015-03-26
==================

  * Fix: Permanent failing jobs should be removed from the queue
  * Moved commands to cmd namespace
  * Do not upload dotfiles
  * Moved BuildDB creation to build-starter
  * Hide assets not for main operating systems
  * Build gobuilder-frontend on top of golang:latest
  * Changed URL to gobuilder.me
  * Replaced curl as it lacks a return state

1.5.0 / 2015-03-02
==================

  * Added buildstatus to frontend

1.4.0 / 2015-02-28
==================

  * Hide all but newest 5 builds

1.3.1 / 2015-02-28
==================

  * Some more code cleanup

1.3.0 / 2015-02-28
==================

  * Removed progress display from rsync
  * Added buildlog display

1.2.2 / 2015-02-28
==================

  * Fix: Do not try to publish directories in asset dir

1.2.1 / 2015-02-28
==================

  * Fix: Exit status 256 is not possible

1.2.0 / 2015-02-28
==================

  * Added build-image Dockerfile / script
  * Added asset upload to build-starter
  * GIT_BRANCH is not longer supported
  * Published all parts of the builder on GitHub

1.1.1 / 2015-02-20
==================

  * Sorting by date seems to be a better idea

1.1.0 / 2015-02-20
==================

  * Added sorted displaying of branches
  * Submit reponame as string to get it displayed in loggly

1.0.1 / 2014-11-07
==================

  * [build-starter] Pull repo to ensure it is present
  * [build-starter] Removed debugging code which could cause runtime errors

1.0.0 / 2014-11-07
==================

  * Initial working version with support for all announced features