all: build run

build:
	@docker-compose build agent-0 client-0

run:
	@docker-compose up && docker-compose down

push:
	@docker-compose push agent-0 client-0

