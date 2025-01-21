#!/usr/bin/env bash

CMD=$1

parameters=("dev" "prod")

function main() {
  if [[ -z $CMD ]]; then
    echo "Error: Parameter must not be empty. Please provide a valid input." >&2
    exit 1
  elif ! element_in_array "${CMD}" "${parameters[@]}"; then
    echo "Error: the parameter is not available" >&2
    exit 1
  fi
  if [ "$CMD" == "dev" ]; then
    docker-compose -f ./docker-compose-dev.yml build --no-cache
    docker-compose -f ./docker-compose-dev.yml up -d
  elif [ "$CMD" == "prod" ]; then
    docker-compose -f ./docker-compose-prod.yml build --no-cache
    docker-compose -f ./docker-compose-prod.yml up -d --remove-orphans
  fi
}

function element_in_array() {
  local element="$1"
  shift
  local array=("$@")

  for item in "${array[@]}"; do
    if [[ "${item}" == "${element}" ]]; then
      return 0
    fi
  done
  echo "Element not found"
  return 1
}

#start point
main