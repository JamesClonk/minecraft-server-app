.PHONY: run server binary deploy setup glide test update push docker-build docker-run docker-publish docker-push
SHELL := /bin/bash

all: run

run: binary
	scripts/run.sh

deploy: binary
	scripts/setup.sh
	scripts/deploy.sh

server:
	java -Xmx1024M -Xms1024M -jar minecraft.jar nogui

binary:
	GOARCH=amd64 GOOS=linux go build -i -o minecraft-server-app

setup:
	go get -v -u github.com/codegangsta/gin
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

push:
	cf push

docker-build:
	docker build -t jamesclonk:minecraft-server-app .

docker-run:
	docker run jamesclonk:minecraft-server-app

docker-publish:
	docker push jamesclonk:minecraft-server-app

docker-push:
	cf push -o jamesclonk/minecraft-server-app -i 1 -m 1536M -k 1G
