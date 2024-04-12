From golang:1.18 

WORKDIR /app

COPY . .

RUN go mod init careplanner-api
Run go mod tidy

RUN mkdir -p  /app/logs/careplanner-api-logs
RUN go build -v -o careplanner-api-receiver

CMD ["/app/careplanner-api-receiver"]