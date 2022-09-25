all: build

build:
	go build ./cmd/free-my-lists

install:
	cp free-my-lists /usr/bin/free-my-lists
	rm -rf /usr/lib/free-my-lists
	mkdir -p /usr/lib/free-my-lists/
	cp -dr ./templates /usr/lib/free-my-lists/templates
