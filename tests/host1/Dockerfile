FROM golang:1.23.2-alpine

WORKDIR /usbgp
COPY ../src/. .
RUN go mod tidy
RUN go build -o ./usbgp ./cmd/usbgp/main.go
CMD ["./usbgp", "64512 10.200.100.2 65413 10.200.100.3 active"]