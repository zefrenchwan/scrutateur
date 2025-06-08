FROM golang:1.24

EXPOSE 3000

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# server
COPY *.go ./
COPY services/ ./services/
COPY engines/ ./engines/
COPY storage/ ./storage/
COPY dto/ ./dto/
# extra content 
COPY static/ /app/static/

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main

RUN mkdir /app/logs/
RUN chmod -R 444 /app/static/

# Run
CMD ["/app/main"]