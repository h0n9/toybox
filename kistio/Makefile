all: build run

build:
	@docker-compose build agent-0 client-0

run:
	@docker-compose up && docker-compose down

push:
	@docker-compose push agent-0 client-0

deploy:
	@k3d image load -c h0n9-cluster ghcr.io/h0n9/kistio-agent:latest ghcr.io/h0n9/sample-client:latest
	@kubectl -n kistio rollout restart deployment/kistio-center
	@sleep 3
	@kubectl -n sample-client rollout restart deployment/account
	@kubectl -n sample-client rollout restart deployment/balance

