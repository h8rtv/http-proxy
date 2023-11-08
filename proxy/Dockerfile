FROM golang:alpine

WORKDIR /app

COPY . .

# Build the Go app
RUN go build -o proxy .

EXPOSE 8080

ENTRYPOINT ["./proxy"]
