FROM golang:alpine

WORKDIR /app

COPY . .

RUN go build -o proxy .

EXPOSE 8080

ENTRYPOINT ["./proxy"]
