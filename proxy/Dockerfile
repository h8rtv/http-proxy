FROM golang:alpine

WORKDIR /app

COPY . .

RUN sha256sum -c out.sha256sum

RUN go build -o proxy .

EXPOSE 8080

ENTRYPOINT ["./proxy"]
