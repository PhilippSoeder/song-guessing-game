# builder stage
FROM golang:1.27.2-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/song-guessing-game ./cmd/song-guessing-game

# runtime stage
FROM gcr.io/distroless/static-debian11

WORKDIR /app
COPY --from=builder /app/song-guessing-game /app/song-guessing-game
COPY --from=builder /app/static /app/static

EXPOSE 8888

ENV PORT=8888
CMD ["/app/song-guessing-game"]
