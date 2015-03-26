
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
