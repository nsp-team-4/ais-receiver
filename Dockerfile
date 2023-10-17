FROM golang:alpine

WORKDIR /app

COPY ./app/go.mod ./app/go.sum ./

RUN go mod download

COPY ./app/ ./

RUN go build -o /main

CMD ["/main"]
