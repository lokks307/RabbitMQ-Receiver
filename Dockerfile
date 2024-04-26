FROM golang:1.21

ARG SUBS="empty"
ARG RUNMODE="prod"

RUN echo "subs and runmode"
RUN echo $SUBS
RUN echo $RUNMODE

ENV SUBS=$SUBS
ENV RUNMODE=$RUNMODE

WORKDIR /app

COPY ./deploy.go .

RUN go mod init deploy
RUN go mod tidy
RUN go build

RUN mkdir -p "/app/logs/${SUBS}-${RUNMODE}-logs"

CMD /app/deploy --subs $SUBS --runmode $RUNMODE
