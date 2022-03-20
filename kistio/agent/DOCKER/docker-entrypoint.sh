#!/bin/sh

FLAGS=

if [ ! -z "$BOOTSTRAP" ]; then
    FLAGS="$FLAGS --bootstrap $BOOTSTRAP"
fi

if [ ! -z "$LISTEN" ]; then
    FLAGS="$FLAGS --listen $LISTEN"
fi

if [ ! -z "$SEED" ]; then
    FLAGS="$FLAGS --seed $SEED"
fi

/usr/bin/app/kistio-agent $FLAGS
