#!/bin/bash
set -e

# Install PostGIS first (required dependency)
psql -v ON_ERROR_STOP=1 --username "" --dbname "" <<-EOSQL
    CREATE EXTENSION IF NOT EXISTS postgis;
    CREATE EXTENSION IF NOT EXISTS postgis_topology;
EOSQL

# Install pgRouting (depends on PostGIS)
psql -v ON_ERROR_STOP=1 --username "" --dbname "" <<-EOSQL
    CREATE EXTENSION IF NOT EXISTS pgRouting CASCADE;
EOSQL

# Create functions in template1 so they are available in all databases
psql -v ON_ERROR_STOP=1 --username "" --dbname template1 <<-EOSQL
    CREATE EXTENSION IF NOT EXISTS postgis;
    CREATE EXTENSION IF NOT EXISTS postgis_topology;
    CREATE EXTENSION IF NOT EXISTS pgRouting CASCADE;
EOSQL
