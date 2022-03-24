#!/bin/sh

FLAGS=

if [ ! -z "$BOOTSTRAPS" ]; then
    FLAGS="$FLAGS --bootstraps $BOOTSTRAPS"
fi

if [ ! -z "$LISTENS" ]; then
    FLAGS="$FLAGS --listens $LISTENS"
fi

if [ ! -z "$SEED" ]; then
    FLAGS="$FLAGS --seed $SEED"
fi

if [ ! -z "$GRPC_LISTEN" ]; then
    FLAGS="$FLAGS --grpc-listen $GRPC_LISTEN"
fi

if [ ! -z "$RENDEZ_VOUS" ]; then
    FLAGS="$FLAGS --rendez-vous $RENDEZ_VOUS"
fi

if [ ! -z "$ENABLE_DHT_SERVER" ]; then
    FLAGS="$FLAGS --enable-dht-server $ENABLE_DHT_SERVER"
fi

/usr/bin/app/kistio-agent $FLAGS
