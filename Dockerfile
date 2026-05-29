FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/main .

FROM alpine:3.21
WORKDIR /forumserver
COPY --from=builder /app/main .
EXPOSE 2333
ENTRYPOINT ["./main"]