#!/bin/bash

mkdir -p /tmp/redis1

redis-server --port 6379 --dir /tmp/redis1 --daemonize yes

sleep 1

./node
