From golang:1.18 

WORKDIR /app

COPY . .

RUN go mod init msv-saas-message-subs
RUN go mod tidy

RUN mkdir -p  /app/logs/msv-saas-message-subs-logs
RUN go build -v -o msv-saas-message-subs

CMD ["/app/msv-saas-message-subs"]
