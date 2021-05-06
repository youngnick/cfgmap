
export GO111MODULE=on
.DEFAULT_GOAL := bin

.PHONY: test
test:
	go test ./pkg/... ./cmd/... -coverprofile cover.out

.PHONY: bin
bin: fmt vet
	go build -o bin/kubectl-dump github.com/youngnick/kubectl-directory-output/cmd/kubectl-dump
	go build -o bin/kubectl-load github.com/youngnick/kubectl-directory-output/cmd/kubectl-load

.PHONY: install
install: fmt vet
	go install github.com/youngnick/kubectl-directory-output/cmd/kubectl-dump
	go install github.com/youngnick/kubectl-directory-output/cmd/kubectl-load

.PHONY: fmt
fmt:
	go fmt ./pkg/... ./cmd/...

.PHONY: vet
vet:
	go vet ./pkg/... ./cmd/...

.PHONY: kubernetes-deps
kubernetes-deps:
	go get k8s.io/client-go@v11.0.0
	go get k8s.io/api@kubernetes-1.14.0
	go get k8s.io/apimachinery@kubernetes-1.14.0
	go get k8s.io/cli-runtime@kubernetes-1.14.0

.PHONY: setup
setup:
	make -C setup