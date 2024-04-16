From golang:1.18 

WORKDIR /app

COPY . .

RUN go mod init msv-saas-checkup-subs
RUN go mod tidy

RUN mkdir -p  /app/logs/msv-saas-checkup-subs-logs
RUN go build -v -o msv-saas-checkup-subs

CMD ["/app/msv-saas-checkup-subs"]
