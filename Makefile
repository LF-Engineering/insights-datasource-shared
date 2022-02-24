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
GO_FILES=context.go email.go error.go es.go exec.go json.go log.go mbox.go redacted.go request.go threads.go time.go utils.go uuid.go
ALL_GO_FILES=context.go email.go error.go es.go exec.go json.go log.go mbox.go redacted.go request.go threads.go time.go utils.go uuid.go firehose/firehose.go
all: check build
check: fmt lint imports vet errcheck
lint: ${ALL_GO_FILES}
	${GO_LINT}
fmt: ${ALL_GO_FILES}
	${GO_FMT} ${GO_FILES}
	${GO_FMT} firehose/firehose.go
vet: ${ALL_GO_FILES}
	${GO_VET} ${GO_FILES}
	${GO_VET} firehose/firehose.go
imports: ${ALL_GO_FILES}
	${GO_IMPORTS} ${GO_FILES}
	${GO_IMPORTS} firehose/firehose.go
errcheck: ${ALL_GO_FILES}
	${GO_ERRCHECK} ${GO_FILES}
	${GO_ERRCHECK} firehose/firehose.go
build: ${ALL_GO_FILES}
	${GO_BUILD}
.PHONY: all
