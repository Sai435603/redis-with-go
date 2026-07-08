FROM golang:tip-alpine3.24

WORKDIR /app

COPY . .

RUN go mod download
# build the go pkg in the curr dir "."
RUN go build -o server .

EXPOSE 8000

CMD ["./server"]