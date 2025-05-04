FROM golang:1.24.2-bookworm

EXPOSE 3000

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY services/ ./services/
COPY storage/ ./storage/

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /main

# Run
CMD ["/main"]