UNAME_S = $(shell uname -s)


.PHONY: all
all: test indexer

# Start the minimum requirements for the service, i.e. db
.PHONY: up
up:
	docker-compose up -d

# Stop all services
.PHONY: down
down:
	docker-compose down

# Explicitly install dependencies. In most cases this is not required as go will automatically download missing deps.
.PHONY: deps
deps:
	go mod download

.PHONY: build-all indexer
build-all: indexer

indexer:
	go install -mod=readonly ./cmd/indexer

# This is a specialized build for running the executable inside a minimal scratch container
.PHONY: build-app
build-app:
ifeq (,$(APP_TYPE))
	@echo "provide APP_TYPE"
else
	go build -ldflags="-w -s" -a -o ./main ./cmd/${APP_TYPE}
endif

# Watch for source code changes to recompile + test
.PHONY: watch
watch:
	GO111MODULE=off go get github.com/cortesi/modd/cmd/modd
	modd

# Run all unit tests
.PHONY: test
test:
	go test -short ./...

# Run all benchmarks
.PHONY: bench
bench:
	go test -short -bench=. ./...

# Same as test but with coverage turned on
.PHONY: cover
cover:
	go test -short -cover -covermode=atomic ./...


# Apply https://golang.org/cmd/gofmt/ to all packages
.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: fmt-check
fmt-check:
ifneq ($(shell gofmt -l .),)
	$(error gofmt fail in $(shell gofmt -l .))
endif

# Apply https://github.com/golangci/golangci-lint to changes since forked from master branch
.PHONY: lint
lint:
	golangci-lint run --timeout=5m --enable=unparam --enable=misspell --enable=prealloc --tests=false

# Remove all compiled binaries from the directory
.PHONY: clean
clean:
	go clean

# Analyze the code for any unused dependencies
.PHONY: prune-deps
prune-deps:
	go mod tidy

# Create the service docker image
.PHONY: image
image:
	docker build --force-rm -t dezswap/dezswap-api .

# Migrate database.
.PHONY: indexer-migrate-test indexer-migrate-up indexer-migrate-down indexer-generate-migration

indexer-migrate-test:
	go test -count=1 -tags=mig ./db/migration/indexer

indexer-migrate-up:
	go run -tags=mig db/migration/indexer/*

 indexer-migrate-down:
	go run -tags=mig db/migration/indexer/* down

# Create a new empty migration file.
indexer-generate-migration:
	$(eval VERSION := $(shell date +"%Y%m%d_%H%M%S"))
	$(eval PATH := db/migration/indexer)
	mkdir -p $(PATH)
	$(shell sed 's/DATE_TIME/$(VERSION)/g' $(PATH)/template.txt > $(PATH)/$(VERSION)_SUMMARY.go)

api-prepare-swagger:
	go install github.com/swaggo/swag/cmd/swag@latest

api-generate-swagger:
	swag init -g api/app.go --output api/docs
