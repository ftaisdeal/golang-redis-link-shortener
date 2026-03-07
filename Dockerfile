FROM golang:1.22-alpine AS build
WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/linkshortener ./cmd/server

FROM alpine:3.20
RUN adduser -D appuser
USER appuser
WORKDIR /app
COPY --from=build /bin/linkshortener /app/linkshortener

EXPOSE 8080
CMD ["/app/linkshortener"]
