.PHONY:
test:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY:
build:
	mkdir -p bin
	go build -v -o ./bin ./cmd/example/...

.PHONY:
install:
	go install -v ./cmd/...

clean:
	rm -rf ./bin
	rm -r coverage.txt

docker-image:
	docker build -t example .