From golang:1.18 

WORKDIR /app

COPY . .

RUN go mod init msv-saas-phr-subs
RUN go mod tidy

RUN mkdir -p  /app/logs/msv-saas-phr-subs-logs
RUN go build -v -o msv-saas-phr-subs

CMD ["/app/msv-saas-phr-subs"]
