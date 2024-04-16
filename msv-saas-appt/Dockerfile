From golang:1.18 

WORKDIR /app

COPY . .

RUN go mod init msv-saas-appt-subs
RUN go mod tidy

RUN mkdir -p  /app/logs/msv-saas-appt-subs-logs
RUN go build -v -o msv-saas-appt-subs

CMD ["/app/msv-saas-appt-subs"]
