
install:
	glide install

test:
	go test

build-ci: install test
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-X main.__VERSION__=${TRAVIS_TAG} -s -w -extldflags '-static'" -o ./dist/kube-deploy-linux
	GOOS=darwin GOARCH=amd64 go build -ldflags="-X main.__VERSION__=${TRAVIS_TAG}" -o ./dist/kube-deploy-mac
