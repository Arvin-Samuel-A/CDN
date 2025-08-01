#!/bin/bash

mkdir -p /tmp/redis1 /tmp/redis2

redis-server --port 6379 --dir /tmp/redis1 --daemonize yes
redis-server --bind 0.0.0.0 --port 6380 --dir /tmp/redis2 --daemonize yes --protected-mode no

sleep 1

./server
