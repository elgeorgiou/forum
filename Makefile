ifneq (,$(wildcard .env))
include .env
export
endif

IMAGE_NAME := gameforum
CONTAINER_NAME := gameforum

.PHONY: run test fmt vet check docker-build docker-run docker-stop docker-clean

run:
	go run ./cmd

test:
	go test ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

check: fmt test vet

docker-build:
	docker build -t $(IMAGE_NAME) .

docker-run:
	@docker run --rm \
		--name $(CONTAINER_NAME) \
		-p 8080:8080 \
		-e GITHUB_CLIENT_ID \
		-e GITHUB_CLIENT_SECRET \
		-e GITHUB_REDIRECT_URL \
		-e GOOGLE_CLIENT_ID \
		-e GOOGLE_CLIENT_SECRET \
		-e GOOGLE_REDIRECT_URL \
		-v $(CURDIR)/data:/app/data \
		-v $(CURDIR)/static/uploads:/app/static/uploads \
		$(IMAGE_NAME)

docker-stop:
	-@docker stop $(CONTAINER_NAME)

docker-clean:
	-@docker rm -f $(CONTAINER_NAME)
	-@docker image rm $(IMAGE_NAME)