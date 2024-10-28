FROM golang:1.23.2-alpine

WORKDIR /usbgp
COPY ../src/. .
RUN go mod tidy
RUN go build -o ./usbgp ./cmd/usbgp
CMD ["./usbgp", "65413 10.200.100.3 64512 10.200.100.2 passive"]