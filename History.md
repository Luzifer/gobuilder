
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
