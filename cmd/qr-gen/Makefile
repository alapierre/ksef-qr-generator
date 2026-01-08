VERSION ?= $(shell \
  TAG=$$(git tag -l "v*" | sort -V | tail -n 1); \
  [ -n "$$TAG" ] && echo $$TAG | sed 's/^v//' || echo 1.0.0 \
)

ROOT_DIR=../..

LDFLAGS = -w -s -X 'github.com/alapierre/ksef-qr-generator.Version=$(VERSION)'

build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -a -installsuffix cgo .

windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)"

push:
	itrust-updater push --artifact-path qr-gen --repo-id 'ksef' --app-id 'qr-gen' --version $(VERSION) --use-keyring
	@if [ -f qr-gen.exe ]; then itrust-updater push --artifact-path qr-gen.exe --repo-id 'ksef' --app-id 'qr-gen' --version $(VERSION) --use-keyring

clean:
	rm -f itrust-updater

.PHONY: gen build docker push clean
