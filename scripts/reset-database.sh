#!/bin/bash
# Complete database reset script
# WARNING: This will DELETE all data and recreate the database from scratch

set -e

echo "=========================================="
echo "⚠️  DATABASE COMPLETE RESET"
echo "=========================================="
echo ""
echo "This script will:"
echo "1. Stop all connections to the database"
echo "2. DROP the entire database"
echo "3. CREATE a fresh database"
echo "4. Run all migrations from scratch"
echo ""
echo "⚠️  WARNING: ALL DATA WILL BE LOST!"
echo ""

read -p "Are you ABSOLUTELY SURE you want to continue? (type 'yes' to proceed): " confirmation

if [ "$confirmation" != "yes" ]; then
    echo "Operation cancelled."
    exit 1
fi

echo ""
echo "Starting database reset..."
echo ""

# Database connection details
DB_HOST="localhost"
DB_PORT="5432"
DB_USER="user"
DB_NAME="fish_db"
DB_PASSWORD="password"

export PGPASSWORD=$DB_PASSWORD

echo "Step 1/4: Terminating existing connections..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres << EOF
-- Terminate all connections to the database
SELECT pg_terminate_backend(pg_stat_activity.pid)
FROM pg_stat_activity
WHERE pg_stat_activity.datname = '$DB_NAME'
  AND pid <> pg_backend_pid();
EOF

echo "✓ Connections terminated"
echo ""

echo "Step 2/4: Dropping database..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres << EOF
DROP DATABASE IF EXISTS $DB_NAME;
EOF

echo "✓ Database dropped"
echo ""

echo "Step 3/4: Creating fresh database..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres << EOF
CREATE DATABASE $DB_NAME;
EOF

echo "✓ Database created"
echo ""

echo "Step 4/4: Running all migrations..."
go run cmd/migrator/main.go up

if [ $? -eq 0 ]; then
    echo ""
    echo "=========================================="
    echo "✓ Database reset completed successfully!"
    echo "=========================================="
    echo ""
    echo "Checking migration status:"
    go run cmd/migrator/main.go version
    echo ""
else
    echo ""
    echo "ERROR: Migration failed!"
    echo "Please check the error messages above."
    exit 1
fi
