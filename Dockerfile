FROM golang:1.21

ARG SUBS="empty"
ARG RUNMODE="prod"
ARG RABBITSERVER="empty"

RUN echo "subs and runmode"
RUN echo $SUBS
RUN echo $RUNMODE
RUN echo "server"
RUN echo $RABBITSERVER

ENV SUBS=$SUBS
ENV RUNMODE=$RUNMODE
ENV RABBITSERVER=$RABBITSERVER

WORKDIR /app

COPY ./deploy.go .

RUN go mod init deploy
RUN go mod tidy
RUN go build

RUN mkdir -p "/app/logs/${SUBS}-${RUNMODE}-logs"

CMD /app/deploy --subs $SUBS --runmode $RUNMODE --rabbitserver $RABBITSERVER
