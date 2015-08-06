#!/bin/bash -e

exec 2>&1

function log {
  echo "[$(date +%H:%M:%S.%N)] $@"
}

product=${REPO##*/}; product=${product%\.*}

SIGNING=1
cat /root/gpgkey.asc.enc | openssl enc -aes-256-cbc -a -d -k ${GPG_DECRYPT_KEY} | gpg --import 2>&1 1>/dev/null || SIGNING=0
if [ ${SIGNING} -eq 1 ]; then
  echo "E2FF3D20865D6F9B6AE74ECB7D5420F913246261:6:" | gpg --import-ownertrust
fi
unset GPG_DECRYPT_KEY

log "Fetching GO repository ${REPO}"
gopath=${REPO}
go get -d -v -u ${REPO}

cd /go/src/${gopath}

if [ ! -z ${COMMIT} ]; then
  git checkout ${COMMIT}
fi

# Fetch all refs from origin for tag / branch detection
git fetch origin

short_commit=$(git rev-parse --short HEAD)
tags=$(git show-ref --tags -d | grep "^${short_commit}" | sed -e 's,.* refs/tags/,,' -e 's/\^{}//')
branches=$(git show-ref -d --heads | grep "^${short_commit}" | sed -e 's,.* refs/heads/,,')

# GoDeps support
if [ -f Godeps/Godeps.json ]; then
  log "Found Godeps. Restoring them"
  go get github.com/tools/godep
  godep restore
fi

go fmt ./...

mkdir -p /tmp/go-build
wget -qO /tmp/go-build/.build_commit "https://gobuilder.me/api/v1/${gopath}/already-built?commit=${short_commit}" || touch /tmp/go-build/.build_commit
wget -qO /tmp/go-build/.build.db https://gobuilder.me/api/v1/${gopath}/build.db || bash -c 'echo "{}" > /tmp/go-build/.build.db'

if [ ! -f .gobuilder.yml ]; then
  # Ensure .gobuilder.yml is present to prevent tools failing later
  echo "---" > .gobuilder.yml
fi
# Upload .gobuilder.yml to enable notifications even when script fails while build
cp .gobuilder.yml /artifacts/
sync

if ! ( test "${FORCE_BUILD}" == "true" ); then
  if [ "$(cat /tmp/go-build/.build_commit)" == "${short_commit}" ]; then
    log "Commit ${short_commit} was already built. Skipping."
    exit 130
  fi
fi

log "Verifying tag signatures..."
for tag in ${tags}; do
  if ( test $(LANG=C git cat-file -t ${tag}) == "tag" ); then
    # Identified as an annotated (real) tag
    if ( LANG=C git tag --verify ${tag} 2>&1 | grep "Good signature" ); then
      LANG=C git tag --verify ${tag} 2>&1 | grep "gpg:" > /tmp/go-build/.signature_${tag}
    fi
  else
    # Identified as a commit (lightweight tag)
    if ( LANG=C git show --show-signature ${tag} | grep "Good signature" ); then
      LANG=C git show --show-signature ${tag} | grep "gpg:" > /tmp/go-build/.signature_${tag}
    fi
  fi

  if ! [ -e /tmp/go-build/.signature_${tag} ]; then
    echo "No valid signature for ${tag}"
  fi
done

log "Verifying commit signature..."
if ( LANG=C git show --show-signature HEAD | grep "Good signature" ); then
  LANG=C git show --show-signature HEAD | grep "gpg:" > /tmp/go-build/.signature_${short_commit}
  for branch in ${branches}; do
    ln /tmp/go-build/.signature_${short_commit} /tmp/go-build/.signature_${branch}
  done
else
  echo "No valid signature for ${short_commit}"
fi

log "Collecting build matrix..."
platforms=$(configreader read arch_matrix)
echo ${platforms}

for platform in ${platforms}; do
  export GOOS=${platform%/*}
  export GOARCH=${platform##*/}
  log "Building ${product} for ${GOOS}-${GOARCH}..."

  mkdir -p /tmp/go-build/${product}/
  echo "go build " \
    "-tags \"$(configreader read build_tags)\"" \
    "-ldflags \"$(configreader read ld_flags)\"" \
    "-o /tmp/go-build/${product}/${product}" \
    "./" | bash -x || { log "Build for ${GOOS}-${GOARCH} failed."; continue; }

  if [ "${GOOS}" == "windows" ]; then
    mv /tmp/go-build/${product}/${product} /tmp/go-build/${product}/${product}.exe
  fi

  log "Collecting artifacts..."
  asset-sync $(pwd) /tmp/go-build/${product}/

  if ! ( configreader checkEmpty version_file ); then
    version_file="/tmp/go-build/${product}/$(configreader read version_file)"
    mkdir -p $(dirname $version_file)
    git rev-parse HEAD >> ${version_file}
  fi

  log "Compressing artifacts..."
  cd /tmp/go-build/
  zip -r ${product}_${short_commit}_${GOOS}-${GOARCH}.zip ${product}
  for tag in ${branches} ${tags}; do
    ln ${product}_${short_commit}_${GOOS}-${GOARCH}.zip ${product}_${tag/\//_}_${GOOS}-${GOARCH}.zip
  done
  cd -

  rm -rf /tmp/go-build/${product}/
done

log "Checking README-File..."
if ! ( configreader checkEmpty readme_file ) && [ -f "$(configreader read readme_file)" ]; then
  cp "$(configreader read readme_file)" /tmp/go-build/${short_commit}_README.md
else
  if [ -f README.md ]; then
    cp README.md /tmp/go-build/${short_commit}_README.md
  fi
fi
if [ -f /tmp/go-build/${short_commit}_README.md ]; then
  cd /tmp/go-build/
  for tag in ${branches} ${tags}; do
    ln ${short_commit}_README.md ${tag/\//_}_README.md
  done
  cd -
fi

log "Building file hashes..."
cd /tmp/go-build/
for tag in ${branches} ${tags}; do
  for artifact in ${product}_${tag}_*.zip; do
    echo "[${artifact}]" >> .hashes_${tag}.txt
    for hasher in md5sum sha1sum sha256sum sha384sum; do
      echo "${hasher} = $(${hasher} ${artifact} | awk {'print $1'})" >> .hashes_${tag}.txt
    done
    echo >> .hashes_${tag}.txt
  done

  echo "---" >> .hashes_${tag}.yaml
  for artifact in ${product}_${tag}_*.zip; do
    echo "${artifact}:" >> .hashes_${tag}.yaml
    for hasher in md5sum sha1sum sha256sum sha384sum; do
      echo "  ${hasher}: $(${hasher} ${artifact} | awk {'print $1'})" >> .hashes_${tag}.yaml
    done
    echo >> .hashes_${tag}.yaml
  done

  if [ $SIGNING -eq 1 ]; then
    gpg --clearsign --output sig .hashes_${tag}.txt
    mv sig .hashes_${tag}.txt
  fi

  echo "${tag}" >> /tmp/go-build/.built_tags
done
cd -

log "Preparing metadata..."
echo ${short_commit} > /tmp/go-build/.build_commit
go version > /tmp/go-build/.goversion

log "Uploading assets..."
rsync -arv /tmp/go-build/ /artifacts/

log "Cleaning up..."
rm -rf /tmp/go-build

log "Build finished."
exit 0
