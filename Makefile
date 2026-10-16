BINARY := gen-android
TEMPLATES := template-main template-single template-kmm

.PHONY: build clean-templates test

build: clean-templates
	go build -o $(BINARY) .

test: clean-templates
	go test ./...

# go:embed cannot exclude paths, so local build output must be removed before building.
clean-templates:
	find $(TEMPLATES) \( -name build -o -name .gradle -o -name .kotlin -o -name .idea \
		-o -name xcuserdata -o -name DerivedData \) -type d -prune -exec rm -rf {} +
	find $(TEMPLATES) \( -name local.properties -o -name .DS_Store \) -type f -delete
