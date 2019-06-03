FROM golang:1.12

RUN go get golang.org/x/net/html

RUN go get github.com/golang/dep
WORKDIR $GOPATH/src/github.com/golang/dep
RUN go install ./...
RUN dep ensure -add github.com/neo4j/neo4j-go-driver/neo4j
RUN wget https://github.com/neo4j-drivers/seabolt/releases/download/v1.7.4/seabolt-1.7.4-Linux-ubuntu-18.04.deb
RUN dpkg -i seabolt-1.7.4-Linux-ubuntu-18.04.deb
RUN go get github.com/neo4j/neo4j-go-driver/neo4j

RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app

COPY . .

CMD ["go", "run", "src/main.go"]