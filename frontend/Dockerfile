FROM golang:1.24 AS builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app .

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app /app/app
EXPOSE 8080
CMD ["/app/app"]
