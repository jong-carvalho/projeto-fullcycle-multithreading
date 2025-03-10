
FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod ./

RUN go mod tidy

COPY . .

RUN go build -o server .

EXPOSE 8080

CMD ["./server"]

#projeto-full-cycle-labs-01/projeto-fullcycle-labs-01