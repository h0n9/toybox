all: build run

build:
	@docker-compose build

run:
	@docker-compose up; docker-compose down

push:
	@docker push ghcr.io/h0n9/kistio-agent
	@docker push ghcr.io/h0n9/sample-client

