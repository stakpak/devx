.PHONY: build
build:
	@go build -o ./bin/ ./cmd/devx

.PHONY: test
test:
	@go test ./... --race --cover

.PHONY: testcue
testcue:
	@cd test && cue eval
