#!/bin/bash
# Cleanup partial migration 6 (create_users_table)
# This script removes all objects that might have been partially created by migration 6

echo "=========================================="
echo "Cleaning up partial migration 6"
echo "=========================================="
echo ""

echo "This script will remove the following objects if they exist:"
echo "- users table"
echo "- All indexes on users table"
echo "- All constraints on users table"
echo "- update_users_updated_at() function"
echo "- trigger_update_users_updated_at trigger"
echo ""

read -p "Do you want to proceed? (yes/no): " confirmation

if [ "$confirmation" != "yes" ]; then
    echo "Operation cancelled."
    exit 1
fi

echo ""
echo "Connecting to database and cleaning up..."
echo ""

# Use the same config as migrator
PGPASSWORD=password psql -h localhost -p 5432 -U user -d fish_db << 'EOF'
-- Drop trigger
DROP TRIGGER IF EXISTS trigger_update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_users_updated_at();

-- Drop all constraints
ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS check_third_party;
ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS check_regular_user;

-- Drop all indexes (explicitly)
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_third_party;
DROP INDEX IF EXISTS idx_users_is_guest;
DROP INDEX IF EXISTS idx_users_created_at;

-- Drop the table
DROP TABLE IF EXISTS users;

-- Verify cleanup
\dt users
EOF

if [ $? -eq 0 ]; then
    echo ""
    echo "✓ Cleanup completed successfully!"
    echo ""
    echo "Next steps:"
    echo "1. Force migration to version 5:"
    echo "   go run cmd/migrator/main.go force 5"
    echo ""
    echo "2. Re-apply migrations:"
    echo "   go run cmd/migrator/main.go up"
    echo ""
else
    echo ""
    echo "ERROR: Cleanup failed!"
    echo "Please check the error messages above."
    exit 1
fi
