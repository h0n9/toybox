#!/usr/bin/dumb-init /bin/sh

FLAGS=

if [ ! -z "$AGENT_HOST" ]; then
    FLAGS="$FLAGS --agent-host $AGENT_HOST"
fi

if [ ! -z "$AGENT_PORT" ]; then
    FLAGS="$FLAGS --agent-port $AGENT_PORT"
fi

if [ ! -z "$TOPIC_PUB" ]; then
    FLAGS="$FLAGS --topic-pub $TOPIC_PUB"
fi

if [ ! -z "$TOPIC_SUB" ]; then
    FLAGS="$FLAGS --topic-sub $TOPIC_SUB"
fi

echo $FLAGS

/usr/bin/app/sample-client $FLAGS
