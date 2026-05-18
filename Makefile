.PHONY: fmt vet lint build test prepare-release tag-release

fmt:
	gofmt -w .

vet:
	go vet ./...
	@for dir in $$(./scripts/moduledirs.sh); do (cd "$$dir" && go vet ./...) || exit $$?; done

lint:
	golangci-lint run ./...
	cd ./example && golangci-lint run ./...
	@for dir in $$(./scripts/moduledirs.sh); do (cd "$$dir" && golangci-lint run ./...) || exit $$?; done

build:
	go build ./...
	@for dir in $$(./scripts/moduledirs.sh); do (cd "$$dir" && go build ./...) || exit $$?; done

test:
	go test ./... -v -failfast
	@for dir in $$(./scripts/moduledirs.sh); do (cd "$$dir" && go test ./... -v -failfast) || exit $$?; done

prepare-release:
ifndef VERSION
	$(error VERSION is required. Usage: make prepare-release VERSION=1.0.0)
endif
	./scripts/prepare-release.sh $(VERSION)

tag-release:
ifndef VERSION
	$(error VERSION is required. Usage: make tag-release VERSION=1.0.0)
endif
	./scripts/tag-release.sh $(VERSION)
