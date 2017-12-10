all: clean update build test show

clean:
	rm -f vers
	rm -rf target

update:
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

package-base: test
	mkdir target
	mkdir target/model
	mkdir target/package

package-osx: package-base
	mkdir target/model/osx
	mkdir target/model/osx/usr
	mkdir target/model/osx/usr/local
	mkdir target/model/osx/usr/local/bin
	install -m 755 vers target/model/osx/usr/local/bin/vers
	fpm -s dir -t osxpkg -n vers -v $(shell ./vers -f version.json show) -p target/package -C target/model/osx .

package-rpm: package-base
	mkdir target/model/linux-x86-rpm
	mkdir target/model/linux-x86-rpm/usr
	mkdir target/model/linux-x86-rpm/usr/bin
	install -m 755 vers target/model/linux-x86-rpm/usr/bin/vers
	fpm -s dir -t rpm -n vers -v $(shell ./vers -f version.json show) -p target/package -C target/model/linux-x86-rpm .

package-deb: package-base
	mkdir target/model/linux-x86-deb
	mkdir target/model/linux-x86-deb/usr
	mkdir target/model/linux-x86-deb/usr/bin
	install -m 755 vers target/model/linux-x86-deb/usr/bin/vers
	fpm -s dir -t deb -n vers -v $(shell ./vers -f version.json show) -p target/package -C target/model/linux-x86-deb .
