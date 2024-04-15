From golang:1.18 

WORKDIR /app

COPY . .

RUN go mod init careplanner-mobile-api-receiver
RUN go mod tidy

RUN mkdir -p  /app/logs/careplanner-mobile-api-logs
RUN go build -v -o careplanner-mobile-api-receiver

CMD ["/app/careplanner-mobile-api-receiver"]
