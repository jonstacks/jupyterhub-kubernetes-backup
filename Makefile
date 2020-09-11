DOCKER_IMAGE_NAME="jupyterhub-kubernetes-backup"

.PHONY:
test:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY:
build:
	mkdir -p bin
	go build -v -o ./bin ./cmd/...

.PHONY:
install:
	go install -v ./cmd/...

clean:
	-rm -rf ./bin
	-rm -r coverage.txt

.PHONY:
docker-image:
	docker build -t $(DOCKER_IMAGE_NAME) .

.PHONY:
run-docker-image: docker-image
	docker run --rm -it --entrypoint=/bin/ash $(DOCKER_IMAGE_NAME) 