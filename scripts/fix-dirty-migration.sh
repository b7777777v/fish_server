#!/bin/bash
# Fix dirty database migration script
# Usage: ./scripts/fix-dirty-migration.sh [version]
# If no version specified, it will force to version 5 (before the dirty migration)

set -e

echo "==================================="
echo "Dirty Migration Fix Script"
echo "==================================="
echo ""

# Check if version argument is provided
if [ -z "$1" ]; then
    echo "No version specified. Will force to version 5 (rollback before dirty migration 6)."
    VERSION=5
else
    VERSION=$1
    echo "Will force migration to version $VERSION"
fi

echo ""
echo "Current migration status:"
go run cmd/migrator/main.go version || echo "Failed to get version (database might not be accessible)"

echo ""
echo "-----------------------------------"
echo "IMPORTANT: Before proceeding, you should:"
echo "1. Check if migration 6 (create_users_table) was partially applied"
echo "2. Manually verify the database state"
echo "3. Decide whether to:"
echo "   - Force to version 5 (rollback) if migration failed early"
echo "   - Force to version 6 (complete) if migration mostly succeeded"
echo "-----------------------------------"
echo ""

read -p "Do you want to force migration to version $VERSION? (yes/no): " confirmation

if [ "$confirmation" != "yes" ]; then
    echo "Operation cancelled."
    exit 1
fi

echo ""
echo "Forcing migration to version $VERSION..."
go run cmd/migrator/main.go force $VERSION

echo ""
echo "✓ Migration forced to version $VERSION"
echo ""
echo "Current migration status:"
go run cmd/migrator/main.go version

echo ""
echo "-----------------------------------"
echo "Next steps:"
if [ "$VERSION" -eq 5 ]; then
    echo "1. Run: go run cmd/migrator/main.go up"
    echo "   This will re-apply migration 6 and subsequent migrations"
elif [ "$VERSION" -eq 6 ]; then
    echo "1. Verify migration 6 was completed correctly"
    echo "2. Run: go run cmd/migrator/main.go up"
    echo "   This will apply any remaining migrations"
fi
echo "-----------------------------------"
