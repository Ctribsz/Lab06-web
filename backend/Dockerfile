FROM golang:1.21

WORKDIR /app

COPY go.mod ./
COPY main.go ./

RUN go mod tidy
RUN go get github.com/lib/pq
RUN go build -o main

CMD ["./main"]