FROM golang:1.24

EXPOSE 3000

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY services/ ./services/
COPY engines/ ./engines/
COPY storage/ ./storage/
COPY dto/ ./dto/

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main

RUN mkdir /app/logs/

# Run
CMD ["/app/main"]