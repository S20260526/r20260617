#!/bin/sh

docker compose exec registrator psql --user=postgres --file=/r20260921-init.sql

docker compose exec -T config etcdctl txn <<EOD

put root.input.port 8089
put root.storage storage:9333
put root.pushing.host rmq-entry:5672
put root.pushing.queue working
put root.pulling.host rmq1:5672,rmq2:5672,rmq3:5672
put root.pulling.queue working
put root.pulling.dlq dlq
put root.pulling.retry.t 10s
put root.script.socket.d /tmp
put root.script.file /opt/python/main.py
put root.registrator.driver postgres
put root.registrator.dsn "host=registrator sslmode=disable user=postgres password=1234"
put root.registrator.table events
put root.metrics.port 8099


EOD
