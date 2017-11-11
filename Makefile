all: clean update build test show

clean:
	rm -f vers

update:
	git status --porcelain --branch
	env | sort
	go get

buildp1:
	go build -ldflags "-X main.version=P1"

build: buildp1
	go build -ldflags "-X main.version=$(shell ./vers -f version.json show)"

show: build
	./vers --version

run: build
	./vers -f version.json show

test: build
	go test
