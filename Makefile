GO ?= go
PROTOC ?= protoc

bin_name := google-font-info
protobuf_go := internal/google_fonts/fonts_public.pb.go

# Hook into every command and dump the target variables
ifeq ($(MAKE_DEBUG), 1)
OLD_SHELL := $(SHELL)
SHELL = $(warning [$@ ($^) ($?)])$(OLD_SHELL)
endif

.PHONY: source
source: $(protobuf_go)
	$(GO) build -o $(bin_name)

.PHONY: clean
clean:
	$(RM) $(bin_name) $(protobuf_go)

$(protobuf_go) : %.pb.go : %.proto
	$(PROTOC) --go_out=paths=source_relative:. $<
