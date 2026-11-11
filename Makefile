BINARY := gen-android
TEMPLATES := templates/main templates/single templates/kmm
VERSION := $(shell git describe --tags --dirty --always 2>/dev/null)
LDFLAGS := -X github.com/paulcoding810/android-template/internal/version.Version=$(VERSION)

.PHONY: build clean-templates test upstream-diff

build: clean-templates
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

test: clean-templates
	go test ./...

# go:embed cannot exclude paths, so local build output must be removed before building.
clean-templates:
	find $(TEMPLATES) \( -name build -o -name .gradle -o -name .kotlin -o -name .idea \
		-o -name xcuserdata -o -name DerivedData \) -type d -prune -exec rm -rf {} +
	find $(TEMPLATES) \( -name local.properties -o -name .DS_Store \) -type f -delete

# Compare a template with a fresh `android create` project, e.g. make upstream-diff TEMPLATE=templates/main
upstream-diff:
	./scripts/upstream-diff.sh $(or $(TEMPLATE),templates/single)
