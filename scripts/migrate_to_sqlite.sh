#!/bin/bash

# Change to the backend directory
cd "$(dirname "$0")/../backend" || exit

# Compile and run the migration tool
echo "Building migration tool..."
go build -o /tmp/migrator ../scripts/migrate_mongo_to_sqlite.go

echo "Running migration from MongoDB to SQLite..."
/tmp/migrator

# Clean up
rm /tmp/migrator

echo "Migration completed. Update your .env file with DATABASE_TYPE=sqlite to use SQLite."