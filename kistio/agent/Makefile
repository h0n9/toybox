all: dep build run

dep:
	@go mod download

build: build-agent

build-agent:
	@go build ./cmd/kistio-agent

run: run-agent

run-agent:
	@./kistio-agent --seed test

docker: docker-build docker-run

docker-build:
	@docker build -t kistio-agent .

docker-run:
	@docker run --rm -it --env SEED=test kistio-agent
