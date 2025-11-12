# 🔧 Dirty Migration Fix Guide

## Problem

You're seeing this error:
```
Dirty database version 6. Fix and force version.
```

This occurs when a migration fails partway through execution, leaving the database in an inconsistent state.

## Important: Database Name Configuration Issue

⚠️ **Configuration Mismatch Detected:**
- `configs/config.dev.yaml` expects database: `fish_db`
- `deployments/docker-compose.dev.yml` creates database: `fish_db_dev`

### Quick Fix Option 1: Update config.dev.yaml

```yaml
# configs/config.dev.yaml
data:
  database:
    dbname: "fish_db_dev"  # Change from fish_db to fish_db_dev
```

### Quick Fix Option 2: Update docker-compose.dev.yml

```yaml
# deployments/docker-compose.dev.yml
services:
  postgres:
    environment:
      POSTGRES_DB: fish_db  # Change from fish_db_dev to fish_db
```

## Understanding the Dirty Migration

Migration 6 (`000006_create_users_table`) creates:
- The `users` table with multiple columns
- Several indexes
- Check constraints
- A trigger function and trigger

The migration could have failed at any of these steps.

## Solution Steps

### Step 1: Ensure Database is Running

#### Using Docker Compose:
```bash
# Start PostgreSQL and Redis
make run-dev

# Or directly:
docker-compose -f deployments/docker-compose.yml up -d postgres redis
```

#### Check database connectivity:
```bash
pg_isready -h localhost -p 5432 -U user
```

### Step 2: Check Current Migration Status

```bash
go run cmd/migrator/main.go version
```

This will show:
- Current version (6)
- Dirty state (true)

### Step 3: Inspect Database State

Connect to the database to see what was actually created:

```bash
# Using psql
PGPASSWORD=password psql -h localhost -p 5432 -U user -d fish_db

# Then in psql:
\dt                          # List tables
\d users                     # Check if users table exists
\di                          # List indexes
SELECT * FROM schema_migrations;  # Check migration state
```

### Step 4: Decide on the Fix Strategy

#### Option A: Rollback to Version 5 (Recommended if migration failed early)

If the `users` table doesn't exist or is incomplete:

```bash
# Using the provided script:
./scripts/fix-dirty-migration.sh 5

# Or manually:
go run cmd/migrator/main.go force 5
```

Then re-run migrations:
```bash
go run cmd/migrator/main.go up
```

#### Option B: Force to Version 6 (If migration mostly completed)

If the `users` table exists with all columns, indexes, and triggers:

```bash
# Using the provided script:
./scripts/fix-dirty-migration.sh 6

# Or manually:
go run cmd/migrator/main.go force 6
```

Then run remaining migrations:
```bash
go run cmd/migrator/main.go up
```

#### Option C: Manual Database Cleanup + Force to 5

If the migration partially applied some changes:

```bash
# Connect to database
PGPASSWORD=password psql -h localhost -p 5432 -U user -d fish_db

# In psql, manually clean up partial changes:
DROP TRIGGER IF EXISTS trigger_update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_users_updated_at();
DROP TABLE IF EXISTS users;
\q

# Then force to version 5:
go run cmd/migrator/main.go force 5

# Re-run migrations:
go run cmd/migrator/main.go up
```

### Step 5: Verify Fix

```bash
# Check version (should not be dirty)
go run cmd/migrator/main.go version

# Verify all migrations are applied
go run cmd/migrator/main.go up
```

## Prevention

To avoid dirty migrations in the future:

1. **Test migrations on a separate database first:**
   ```bash
   # Create test database
   PGPASSWORD=password psql -h localhost -p 5432 -U user -c "CREATE DATABASE fish_db_test;"

   # Test migration
   migrate -database "postgresql://user:password@localhost:5432/fish_db_test?sslmode=disable" \
           -path ./storage/migrations up
   ```

2. **Use transactions in migrations** (where possible):
   - Migration 6 uses multiple DDL statements
   - PostgreSQL supports transactional DDL
   - The golang-migrate tool wraps each migration in a transaction by default

3. **Review migration files before applying:**
   ```bash
   # Check what will be applied
   ls storage/migrations/
   cat storage/migrations/000006_create_users_table.up.sql
   ```

4. **Always backup before migrations:**
   ```bash
   pg_dump -h localhost -U user fish_db > backup_before_migration.sql
   ```

## Quick Reference Commands

```bash
# Check migration status
go run cmd/migrator/main.go version

# Force to specific version
go run cmd/migrator/main.go force <version>

# Apply all pending migrations
go run cmd/migrator/main.go up

# Rollback one migration
go run cmd/migrator/main.go down

# Start database
make run-dev

# Connect to database
PGPASSWORD=password psql -h localhost -p 5432 -U user -d fish_db
```

## Troubleshooting

### Issue: "Failed to create migrate instance: dial tcp 127.0.0.1:5432: connection refused"

**Solution:** Database is not running. Start it with:
```bash
make run-dev
```

### Issue: Database name not found

**Solution:** Check the database name mismatch issue at the top of this guide.

### Issue: Still dirty after forcing version

**Solution:**
1. Check you forced to the correct version
2. Verify with: `go run cmd/migrator/main.go version`
3. The dirty flag should be `false` after forcing

### Issue: Migration fails again after fixing

**Solution:**
1. Check the error message carefully
2. Verify database schema matches expected state
3. Check for conflicts (e.g., table already exists)
4. Consider manual cleanup before re-running

## Additional Resources

- [golang-migrate Documentation](https://github.com/golang-migrate/migrate)
- [PostgreSQL Migration Best Practices](https://www.postgresql.org/docs/current/ddl.html)
