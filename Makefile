.PHONY: run binary setup glide test update push
SHELL := /bin/bash

all: run

run: binary
	scripts/run.sh

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
	cf push
