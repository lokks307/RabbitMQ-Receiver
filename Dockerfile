FROM golang:1.21

ARG SUBS="empty"
ARG RUNMODE="test"

RUN echo "subs and runmode"
RUN echo $SUBS
RUN echo $RUNMODE

ENV SUBS=$SUBS
ENV RUNMODE=$RUNMODE

WORKDIR /app

COPY deploy .

RUN mkdir -p "/app/logs/${SUBS}-${RUNMODE}-logs"

CMD /app/deploy --subs $SUBS --runmode $RUNMODE
