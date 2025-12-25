FROM golang:1.24-alpine AS build
WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /app ./cmd/orders

FROM alpine:3.20
WORKDIR /
COPY --from=build /app /app
EXPOSE 8081
ENTRYPOINT ["/app"]