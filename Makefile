SHA = $(shell git rev-parse HEAD)
BINARY = $(notdir $(shell pwd))

run:
	go build -o ${BINARY}
	./${BINARY} -log-lvl DEBUG

build:
	go build -o ${BINARY}

build-linux:
	GOOS=linux GOARCH=amd64 go build -o ${BINARY}

clean:
	rm -rf ./${BINARY}

docker:
	docker build --build-arg BINARY=${BINARY}  -t skirsch10/${BINARY}:$(SHA) --platform=linux/amd64 .

push:
	docker push skirsch10/${BINARY}:$(SHA)%