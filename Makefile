.PHONY: run server binary setup glide test update push
SHELL := /bin/bash

all: run

run: binary
	scripts/run.sh

server:
	java -Xmx1024M -Xms1024M -jar minecraft.jar nogui

binary:
	GOARCH=amd64 GOOS=linux go build -i -o minecraft-server-app

setup:
	go get -v -u github.com/codegangsta/gin
	go get -v -u github.com/Masterminds/glide

glide:
	glide install --force

test:
	GOARCH=amd64 GOOS=linux go test $$(go list ./... | grep -v /vendor/)

update:
	git checkout master
	git fetch --all
	git merge upstream/master
	git push

push:
	cf push -o JamesClonk/minecraft-server-app -i 1 -m 1536M -k 1G
