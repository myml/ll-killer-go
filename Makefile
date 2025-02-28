GO  ?= go
GOARCH ?= $(shell $(GO) env GOARCH)
LDFLAGS_STATIC := -extldflags -static
TRIMPATH := -trimpath

GO_BUILD := $(GO) build $(TRIMPATH) $(GO_BUILDMODE_STATIC) \
	$(EXTRA_FLAGS) -ldflags "$(LDFLAGS_STATIC) $(EXTRA_LDFLAGS)"

ll-killer: *.go build-aux/*
	$(GO_BUILD) -o $@ .