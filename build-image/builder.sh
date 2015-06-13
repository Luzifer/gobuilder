#!/bin/bash -e

exec 2>&1

function log {
  echo "[$(date +%H:%M:%S.%N)] $@"
}

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
wget -qO /tmp/go-build/build_master https://gobuilder.me/api/v1/${gopath}/last-build || touch /tmp/go-build/build_master
wget -qO /tmp/go-build/build.db https://s3-eu-west-1.amazonaws.com/gobuild.luzifer.io/${gopath}/build.db || bash -c 'echo "{}" > /tmp/go-build/build.db'

if [ ! -f .gobuilder.yml ]; then
  # Ensure .gobuilder.yml is present to prevent tools failing later
  echo "---" > .gobuilder.yml
fi
# Upload .gobuilder.yml to enable notifications even when script fails while build
cp .gobuilder.yml /artifacts/
sync

if [ "$(cat /tmp/go-build/build_master)" == "${short_commit}" ]; then
  log "Commit ${short_commit} was already built. Skipping."
  exit 130
fi

log "Verifying tag signatures..."
for tag in ${tags}; do
  if ( LANG=C git tag --verify ${tag} 2>&1 | grep "Good signature" ); then
    LANG=C git tag --verify ${tag} 2>&1 | grep "gpg:" > /tmp/go-build/.signature_${tag}
  else
    echo "No valid signature for ${tag}"
  fi
done

log "Verifying commit signature..."
if ( LANG=C git show --show-signature HEAD | grep "Good signature" ); then
  LANG=C git show --show-signature HEAD | grep "gpg:" > /tmp/go-build/.signature_master
else
  echo "No valid signature for master"
fi

if ! ( configreader checkEmpty artifacts ); then
  configreader read artifacts > /tmp/go-build/.artifact_files
fi

for platform in ${GOLANG_CROSSPLATFORMS}; do
  export GOOS=${platform%/*}
  export GOARCH=${platform##*/}
  log "Building ${product} for ${GOOS}-${GOARCH}..."

  mkdir -p /tmp/go-build/${product}/
  go build -o /tmp/go-build/${product}/${product} ./ || { log "Build for ${GOOS}-${GOARCH} failed."; continue; }

  if [ "${GOOS}" == "windows" ]; then
    mv /tmp/go-build/${product}/${product} /tmp/go-build/${product}/${product}.exe
  fi

  if [ -e /tmp/go-build/.artifact_files ]; then
    log "Collecting artifacts..."
    rsync -arv --files-from=/tmp/go-build/.artifact_files ./ /tmp/go-build/${product}/
  fi

  if ! ( configreader checkEmpty version_file ); then
    version_file="/tmp/go-build/${product}/$(configreader read version_file)"
    mkdir -p $(dirname $version_file)
    git rev-parse HEAD >> ${version_file}
  fi

  log "Compressing artifacts..."
  cd /tmp/go-build/
  zip -r ${product}_master_${GOOS}-${GOARCH}.zip ${product}
  for tag in ${tags}; do
    ln ${product}_master_${GOOS}-${GOARCH}.zip ${product}_${tag/\//_}_${GOOS}-${GOARCH}.zip
  done
  cd -

  rm -rf /tmp/go-build/${product}/
done

log "Checking README-File..."
if ! ( configreader checkEmpty readme_file ) && [ -f "$(configreader read readme_file)" ]; then
  cp "$(configreader read readme_file)" /tmp/go-build/master_README.md
else
  if [ -f README.md ]; then
    cp README.md /tmp/go-build/master_README.md
  fi
fi
if [ -f /tmp/go-build/master_README.md ]; then
  cd /tmp/go-build/
  for tag in ${tags}; do
    ln master_README.md ${tag/\//_}_README.md
  done
  cd -
fi

log "Building file hashes..."
cd /tmp/go-build/
for tag in master ${tags}; do
  echo "---" >> .hashes_${tag}.yaml
  for artifact in ${product}_${tag}_*.zip; do
    echo "${artifact}:" >> .hashes_${tag}.yaml
    for hasher in md5sum sha1sum sha256sum sha384sum; do
      echo "  ${hasher}: $(${hasher} ${artifact} | awk {'print $1'})" >> .hashes_${tag}.yaml
    done
  done

  echo "${tag}" >> /tmp/go-build/.built_tags
done
cd -

log "Preparing metadata..."
echo ${short_commit} > /tmp/go-build/build_master
go version > /tmp/go-build/.goversion

log "Uploading assets..."
rsync -arv /tmp/go-build/ /artifacts/

log "Cleaning up..."
rm -rf /tmp/go-build

log "Build finished."
exit 0
