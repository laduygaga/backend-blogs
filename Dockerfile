# Use an official Go runtime as a parent image
FROM golang:1.21 as build

WORKDIR /golang
COPY . .

RUN CGO_ENABLED=0 go build -o ota-blog-admin cmd/web/main.go

EXPOSE 9090
CMD ["./ota-blog-admin"]

