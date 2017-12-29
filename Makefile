all: clean update build test show

CMD := vers
PKG_NAME := vers

clean:
	rm -f $(CMD)
	rm -rf target

update:
	go get

buildp1:
	go build -ldflags "-X main.version=P1"

build: buildp1
	go build -ldflags "-X main.version=$(shell ./$(CMD) -f version.json show)"

set-prefix:
ifndef PREFIX
ifeq ($(shell uname),Darwin)
	$(eval PREFIX := /usr/local)
	$(eval INSTALL_USER := $(shell id -u))
	$(eval INSTALL_GROUP := $(shell id -g))
else
	$(eval PREFIX := /usr)
	$(eval INSTALL_USER := root)
	$(eval INTALL_GROUP := root)
endif
endif

install: set-prefix test
	install -m 755 -o $(INSTALL_USER) -g $(INSTALL_GROUP) $(CMD) $(PREFIX)/bin/$(CMD)

show: build
	./$(CMD) --version

run: build
	./$(CMD) -f version.json show

test: build
	go test

version:
	$(eval VERSION := $(shell ./$(CMD) -f version.json show))

package-base: test
	mkdir target
	mkdir target/model
	mkdir target/package

package-osx: package-base version
	mkdir target/model/osx
	mkdir target/model/osx/usr
	mkdir target/model/osx/usr/local
	mkdir target/model/osx/usr/local/bin
	install -m 755 $(CMD) target/model/osx/usr/local/bin/$(CMD)
	fpm -s dir -t osxpkg -n $(PKG_NAME) -v $(VERSION) -p target/package -C target/model/osx .

package-rpm: package-base
	mkdir target/model/linux-x86-rpm
	mkdir target/model/linux-x86-rpm/usr
	mkdir target/model/linux-x86-rpm/usr/bin
	install -m 755 $(CMD) target/model/linux-x86-rpm/usr/bin/vers
	fpm -s dir -t rpm -n $(PKG_NAME) -v $(VERSION) -p target/package -C target/model/linux-x86-rpm .

package-deb: package-base
	mkdir target/model/linux-x86-deb
	mkdir target/model/linux-x86-deb/usr
	mkdir target/model/linux-x86-deb/usr/bin
	install -m 755 $(CMD) target/model/linux-x86-deb/usr/bin/vers
	fpm -s dir -t deb -n $(PKG_NAME) -v $(VERSION) -p target/package -C target/model/linux-x86-deb .
