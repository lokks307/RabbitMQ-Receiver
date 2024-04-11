From golang:1.18 

WORKDIR /app

COPY . .

RUN go mod init careplanner-admin-api
Run go mod tidy

RUN mkdir -p  /app/logs/careplanner-admin-api-logs
RUN go build -v -o careplanner-admin-api-receiver

CMD ["/app/careplanner-admin-api-receiver"]