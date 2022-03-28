all: dep build run

dep:
	@go mod download

build: build-agent

build-agent:
	@go build ./cmd/kistio-agent

run: run-agent

run-agent:
	@./kistio-agent --seed test

docker: docker-build

docker-push:
	@docker push h0n9/kistio-agent

docker-build:
	@docker build -t h0n9/kistio-agent .

