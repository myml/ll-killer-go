GO  ?= go
GOARCH ?= $(shell $(GO) env GOARCH)
TARGET ?= $(shell echo $$(uname -m)-$$(uname -s | tr '[:upper:]' '[:lower:]')-gnu)
LDFLAGS_STATIC := -extldflags -static
TRIMPATH := -trimpath

VERSION := $(shell git describe --tags --always)
BUILDTIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X "main.Version=$(VERSION)" -X "main.BuildTime=$(BUILDTIME)"

FUSE_LIBS := libfuse-overlayfs.a libgnu.a
FUSE_DIR := fuse-overlayfs
FUSE_PROJECT_DEPS := configure.ac Makefile.am
FUSE_PROJECT_SRC=Makefile *.c *h
FUSE_PROJECT := $(foreach file, $(FUSE_PROJECT_DEPS), $(FUSE_DIR)/$(file))
FUSE_SRCS := $(foreach file, $(FUSE_PROJECT_SRC), $(FUSE_DIR)/$(file))

SRC_DIRS := apps layer pty utils types
GO_SOURCES:= $(wildcard *.go) $(shell find $(SRC_DIRS) -name '*.go')
GO_BUILD := $(GO) build $(TRIMPATH) $(GO_BUILDMODE_STATIC) \
	$(EXTRA_FLAGS) -ldflags "$(LDFLAGS) $(LDFLAGS_STATIC) $(EXTRA_LDFLAGS)"

ll-killer: $(GO_SOURCES) apps/build-aux/build-aux/* apps/build-aux/apt.conf.d/* $(FUSE_LIBS)
	$(GO_BUILD) -o $@ .

$(FUSE_DIR)/Makefile: $(FUSE_PROJECT)
	cd $(FUSE_DIR);\
	git apply --check ../patches/fuse-overlayfs.patch -q && git apply ../patches/fuse-overlayfs.patch;\
	./autogen.sh;\
	LIBS="-ldl" LDFLAGS="-static" ./configure --host=$(TARGET);

$(FUSE_LIBS): $(FUSE_SRCS)
	make -C $(FUSE_DIR)
	cp $(FUSE_DIR)/lib/libgnu.a \
	   $(FUSE_DIR)/libfuse-overlayfs.a .
