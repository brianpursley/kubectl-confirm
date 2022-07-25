# Copyright 2022 Brian Pursley.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

VERSION=`git tag --sort=committerdate | grep -E 'v[0-9].*' | tail -1 | cut -b 2-`
GIT_COMMIT=`git rev-parse HEAD`
LDFLAGS=-ldflags="-X 'github.com/brianpursley/kubectl-confirm/internal/version.Version=${VERSION}' -X 'github.com/brianpursley/kubectl-confirm/internal/version.GitCommit=${GIT_COMMIT}'"

.PHONY: build
build: checks
	@echo "Building"
	@mkdir -p _output && cd _output && go build $(LDFLAGS) ../cmd/kubectl-confirm.go

checks: staticcheck lint vet verify test

.PHONY: install
install: build
	@echo "Installing"
	@mv _output/kubectl-confirm ~/go/bin/kubectl-confirm

.PHONY: clean
clean:
	@echo "Deleting _output"
	@rm -rf _output

.PHONY: test
test:
	@echo "Running go test"
	@go test $(LDFLAGS) ./...

.PHONY: staticcheck
staticcheck:
ifeq (, $(shell which staticcheck))
	$(error staticcheck not found (go install honnef.co/go/tools/cmd/staticcheck@latest))
endif
	@echo "Running staticcheck"
	@staticcheck ./...

.PHONY: lint
lint:
ifeq (, $(shell which golint))
	$(error golint not found (go install golang.org/x/lint/golint@latest))
endif
	@echo "Running golint"
	@golint -set_exit_status ./...

.PHONY: vet
vet:
	@echo "Running go vet"
	@go vet ./...

.PHONY: verify
verify:
	@echo "Running go verify"
	@go mod verify

.PHONY: release
release: clean darwin-amd64 darwin-arm64 linux-amd64 linux-arm64 windows-amd64 sha256
	@cat _output/checksum.txt

.PHONY: darwin-amd64
darwin-amd64:
	@echo "Building darwin amd64"
	@mkdir -p _output && cd _output && \
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) ../cmd/kubectl-confirm.go && \
	cp ../LICENSE . && \
	tar -czf kubectl-confirm-darwin-amd64.tar.gz kubectl-confirm LICENSE && \
	rm kubectl-confirm LICENSE

.PHONY: darwin-arm64
darwin-arm64:
	@echo "Building darwin arm64"
	@mkdir -p _output && cd _output && \
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) ../cmd/kubectl-confirm.go && \
	cp ../LICENSE . && \
	tar -czf kubectl-confirm-darwin-arm64.tar.gz kubectl-confirm LICENSE && \
	rm kubectl-confirm LICENSE

.PHONY: linux-amd64
linux-amd64:
	@echo "Building linux amd64"
	@mkdir -p _output && cd _output && \
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) ../cmd/kubectl-confirm.go && \
	cp ../LICENSE . && \
	tar -czf kubectl-confirm-linux-amd64.tar.gz kubectl-confirm LICENSE && \
	rm kubectl-confirm LICENSE

.PHONY: linux-arm64
linux-arm64:
	@echo "Building linux arm64"
	@mkdir -p _output && cd _output && \
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) ../cmd/kubectl-confirm.go && \
	cp ../LICENSE . && \
	tar -czf kubectl-confirm-linux-arm64.tar.gz kubectl-confirm LICENSE && \
	rm kubectl-confirm LICENSE

.PHONY: windows-amd64
windows-amd64:
	@echo "Building windows amd64"
	@mkdir -p _output && cd _output && \
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) ../cmd/kubectl-confirm.go && \
	cp ../LICENSE . && \
	tar -czf kubectl-confirm-windows-amd64.tar.gz kubectl-confirm.exe LICENSE && \
	rm kubectl-confirm.exe LICENSE

sha256:
	@echo "Generating checksum.txt"
	@cd _output && sha256sum *.gz > checksum.txt