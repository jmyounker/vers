clean:
	rm vers

build:
	go build

run: build
	./vers

test:
	go test