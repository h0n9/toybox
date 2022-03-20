all: build run

build: build-agent

build-agent:
	@go build ./cmd/kistio-agent

run: run-agent

run-agent:
	@./kistio-agent --seed test
