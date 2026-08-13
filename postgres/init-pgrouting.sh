#!/bin/bash
set -e

# Aktifkan pgRouting pada database yang dibuat (POSTGRES_DB) dan pada template1
# agar tersedia juga untuk database baru.
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-'EOSQL'
    CREATE EXTENSION IF NOT EXISTS pgRouting;
    CREATE EXTENSION IF NOT EXISTS postgis;
EOSQL

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname template1 <<-'EOSQL'
    CREATE EXTENSION IF NOT EXISTS pgRouting;
EOSQL