bindata:
		go-bindata --pkg frontend -o frontend/bindata.go --prefix frontend frontend/

godeps_save:
		godeps save ./...

build_test:
		docker run --rm -ti -v $(CURDIR):/go/src/github.com/Luzifer/gobuilder golang:alpine /bin/sh -c 'set -ex \
																					&& go build github.com/Luzifer/gobuilder \
																				  && go build github.com/Luzifer/gobuilder/cmd/asset-sync \
																				  && go build github.com/Luzifer/gobuilder/cmd/configreader \
																				  && go build github.com/Luzifer/gobuilder/cmd/gobuilder-cli \
																				  && go build github.com/Luzifer/gobuilder/cmd/starter'

travis:
		@docker login -e="." -u="$(DOCKER_USERNAME)" -p="$(DOCKER_PASSWORD)" quay.io
		docker build --pull --no-cache -t quay.io/gobuilder/build-image ./build-image
		docker push quay.io/gobuilder/build-image
