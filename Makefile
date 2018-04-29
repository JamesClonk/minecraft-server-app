.PHONY: run server build deploy setup glide test update push
SHELL := /bin/bash

all: run

run: build
	scripts/run.sh

deploy: build
	scripts/setup.sh
	scripts/deploy.sh

server:
	java -Xmx1024M -Xms1024M -jar minecraft.jar nogui

build:
	GOARCH=amd64 GOOS=linux go build -i -o minecraft-server-app
	javac launcher/*.java
	jar -cfe minecraft-server-app.jar launcher.Main launcher/*.class minecraft-server-app minecraft.jar server.properties rcon-cli mc eula.txt ops.json banned-ips.json banned-players.json whitelist.json server-icon.png scripts/backup.sh

setup:
	# go get -v -u github.com/codegangsta/gin
	go get -v -u github.com/Masterminds/glide
	scripts/setup.sh

glide:
	glide install --force

test:
	GOARCH=amd64 GOOS=linux go test $$(go list ./... | grep -v /vendor/)

update:
	git checkout master
	git fetch --all
	git merge upstream/master
	git push

push: build
	cf push
