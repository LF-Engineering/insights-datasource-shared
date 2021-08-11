#for race CGO_ENABLED=1
#GO_ENV=CGO_ENABLED=1
GO_ENV=CGO_ENABLED=0
#GO_BUILD=go build -ldflags '-s -w' -race
GO_BUILD=go build -ldflags '-s -w'
GO_FMT=gofmt -s -w
GO_LINT=golint -set_exit_status
GO_VET=go vet
GO_IMPORTS=goimports -w
GO_ERRCHECK=errcheck -asserts -ignore '[FS]?[Pp]rint*'
GO_FILES=error.go log.go redacted.go time.go utils.go
all: check build
check: fmt lint imports vet errcheck
lint: ${GO_FILES}
	${GO_LINT}
fmt: ${GO_FILES}
	${GO_FMT} ${GO_FILES}
vet: ${GO_FILES}
	${GO_VET} ${GO_FILES}
imports: ${GO_FILES}
	${GO_IMPORTS} ${GO_FILES}
errcheck: ${GO_FILES}
	${GO_ERRCHECK} ${GO_FILES}
build: ${GO_FILES}
	${GO_BUILD}
