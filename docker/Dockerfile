FROM golang:latest

WORKDIR /app

COPY ../src /app/

RUN go mod download

RUN go build -o main .

EXPOSE 8080

ENV DB_USER=go_user \
    DB_PASSWORD=go_pwd \
    DB_HOST=mysql \
    DB_PORT=3306 \
    DB_NAME=tododb \
    SERVER_PORT=8080

CMD ["./main"]
