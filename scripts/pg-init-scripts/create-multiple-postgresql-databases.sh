#!/bin/bash

set -e
set -u

function create_user_and_database() {
	local database=$1
	local username=$2
	local password=$3
	echo "  Creating user and database '$database' username '$username' password '$password'"
	psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
	    CREATE USER $username WITH SUPERUSER PASSWORD '$password';
	    CREATE DATABASE $database owner $username;
EOSQL
	psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
	    GRANT ALL PRIVILEGES ON SCHEMA public TO $username;
      GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $username;
      GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO $username;
EOSQL
}
function main() {
  create_user_and_database "temporal" "temporal" "temporal"
}

main