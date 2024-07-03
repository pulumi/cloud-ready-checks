PROJECT      := github.com/pulumi/cloud-ready-checks

GO              ?= go

all:: ensure test

ensure::
	$(GO) mod download

test::
	$(GO) test -short -v -coverprofile="coverage.txt" -coverpkg=./... ./...
