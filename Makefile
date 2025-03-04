GO  ?= go
GOARCH ?= $(shell $(GO) env GOARCH)
LDFLAGS_STATIC := -extldflags -static
TRIMPATH := -trimpath

VERSION := $(shell git describe --tags --always)
BUILDTIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X "main.Version=$(VERSION)" -X "main.BuildTime=$(BUILDTIME)"

GO_BUILD := $(GO) build $(TRIMPATH) $(GO_BUILDMODE_STATIC) \
	$(EXTRA_FLAGS) -ldflags "$(LDFLAGS) $(LDFLAGS_STATIC) $(EXTRA_LDFLAGS)"

ll-killer: *.go build-aux/* ptrace/*.go apt.conf.d/*
	$(GO_BUILD) -o $@ .