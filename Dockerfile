FROM golang:1.14-alpine as builder

RUN mkdir /app
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go

FROM alpine
WORKDIR /root
COPY --from=builder /app/main .
CMD ["./main"]