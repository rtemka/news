# syntax=docker/dockerfile:1

FROM golang:1.18-buster As build
LABEL version="1.2.3" maintainer="Artem Rybakov<rybakov333@gmail.com>" 

WORKDIR /go/src/github.com/rtemka/news

# copy source code to WORKDIR
COPY go.* .
COPY ./pkg ./pkg/
COPY ./cmd ./cmd/

# install dependencies

RUN go mod tidy 

# build binary; CGO_ENABLED=0 needed to compile binary with no external dependencies
# -ldflags "-s -w" strips out debugging information from binary

RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o ./cmd/news/news ./cmd/news/news.go

# Second stage

FROM alpine:latest

WORKDIR /app

COPY --from=build go/src/github.com/rtemka/news/cmd/news/ .

# 8080 - API listen port; 5432 - postgres port; 27017 - mongodb port

EXPOSE 8080 5432 27017

ENTRYPOINT [ "./news", "./config.json" ]