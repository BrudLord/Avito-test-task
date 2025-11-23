FROM golang:1.25

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/api

EXPOSE 8080

CMD ["./server"]