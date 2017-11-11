
all: clean update build test

clean:
	rm -f vers

update:
	git status --porcelain --branch
	go get

buildp1:
	go build -ldflags "-X main.version=P1"

build: buildp1
	go build -ldflags "-X main.version=$(shell ./vers -f version.json show)"

run: build
	./vers

test:
	go test
