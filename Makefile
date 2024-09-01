build:
	go build -o bin/weather

run: build
	./bin/weather

test:
	go test -v -race ./...