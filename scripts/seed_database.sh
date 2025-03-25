#!/bin/bash

# Check if MongoDB container is running
if ! docker ps | grep -q reconya-mongo-dev; then
    echo "MongoDB container not running. Starting it now..."
    ./dev_start_mongo.sh
    
    # Wait for MongoDB to start
    sleep 3
fi

# Run the seed script
echo "Seeding the database..."
docker exec -i reconya-mongo-dev mongosh "mongodb://localhost:27017/reconya-dev" < seed_database.js

echo "Done! The database has been populated with sample data."