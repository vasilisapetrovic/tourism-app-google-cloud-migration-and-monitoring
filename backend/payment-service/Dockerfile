FROM golang:alpine

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN GONOSUMDB=* go mod download

COPY . .
RUN go build -o main .

EXPOSE 8086

CMD ["./main"]
