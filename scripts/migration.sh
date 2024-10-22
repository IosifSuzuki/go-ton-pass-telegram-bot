#!/usr/bin/env bash

source .env

CMD=$1

export HOST=$POSTGRES_HOST
function main {
  source .env
	if [ "$CMD" == "migrate_up" ]; then
		migrate -path db/migration -database "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=${POSTGRES_MODE}" -verbose up
	elif [ "$CMD" == "migrate_down" ]; then
		migrate -path db/migration -database "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=${POSTGRES_MODE}" -verbose down
	else
		migrate -path db/migration -database "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=${POSTGRES_MODE}" -verbose drop
	fi
}

main