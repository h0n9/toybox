#!/bin/sh

FLAGS=

if [ ! -z "$HOST" ]; then
    FLAGS="$FLAGS --host $HOST"
fi

if [ ! -z "$PORT" ]; then
    FLAGS="$FLAGS --port $PORT"
fi

if [ ! -z "$TOPIC_PUB" ]; then
    FLAGS="$FLAGS --topic-pub $TOPIC_PUB"
fi

if [ ! -z "$TOPIC_SUB" ]; then
    FLAGS="$FLAGS --topic-sub $TOPIC_SUB"
fi

/usr/bin/app/sample-client $FLAGS
