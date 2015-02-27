#!/bin/bash -e

function log {
  echo "[$(date +%H:%M:%S.%N)] $@"
}

export GO_VERSION=$(go version)

branch=${GIT_BRANCH}
product=${REPO##*/}; product=${product%\.*}

log "Fetching GO repository ${REPO}"
gopath=${REPO}
go get -v -u ${REPO}

cd /go/src/${gopath}

short_commit=$(git rev-parse HEAD | head -c6)
tags=$(git tag -l --contains HEAD)

# GoDeps support
if [ -f Godeps/Godeps.json ]; then
  log "Found Godeps. Restoring them"
  go get github.com/tools/godep
  godep restore
fi

go fmt ./...

mkdir -p /tmp/go-build
curl https://s3-eu-west-1.amazonaws.com/gobuild.luzifer.io/${gopath}/build_${branch} -o /tmp/go-build/build_${branch} || touch /tmp/go-build/build_${branch}
curl https://s3-eu-west-1.amazonaws.com/gobuild.luzifer.io/${gopath}/build.db -o /tmp/go-build/build.db || bash -c 'echo "{}" > /tmp/go-build/build.db'

if [ "$(cat /tmp/go-build/build_${branch})" == "${short_commit}" ]; then
  log "Commit ${short_commit} was already built. Skipping."
  exit 0
fi

echo ${short_commit} > /tmp/go-build/build_${branch}

for platform in ${GOLANG_CROSSPLATFORMS}; do
  export GOOS=${platform%/*}
  export GOARCH=${platform##*/}
  log "Building ${product} for ${GOOS}-${GOARCH}..."

  mkdir -p /tmp/go-build/${product}/
  go build -o ./${product} ./ || { log "Build for ${GOOS}-${GOARCH} failed."; continue; }

  if [ "${GOOS}" == "windows" ]; then
    mv ./${product} /tmp/go-build/${product}/${product}.exe
  else
    mv ./${product} /tmp/go-build/${product}/${product}
  fi

  if [ -e .artifact_files ]; then
    log "Collecting artifacts..."
    rsync -ar --progress --files-from=.artifact_files ./ /tmp/go-build/${product}/
  fi

  log "Compressing artifacts..."
  cd /tmp/go-build/
  zip -r ${product}_${branch}_${GOOS}-${GOARCH}.zip ${product}
  for tag in ${tags}; do
    ln ${product}_${branch}_${GOOS}-${GOARCH}.zip ${product}_${tag/\//_}_${GOOS}-${GOARCH}.zip
  done
  cd -

  rm -rf /tmp/go-build/${product}/
done

log "Creating builddb..."
cd /tmp/go-build
/go/bin/builddb_creator

log "Uploading assets..."
rsync -ar --progress /tmp/go-build/ /artifacts/

log "Cleaning up..."
rm -rf /tmp/go-build

log "Build finished."
