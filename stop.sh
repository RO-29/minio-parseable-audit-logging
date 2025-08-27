#!/bin/bash

echo "🛑 Stopping Parseable + MinIO services..."

if command -v docker-compose > /dev/null 2>&1; then
    docker-compose down
else
    docker compose down
fi

echo "✅ Services stopped!"
echo "💡 To remove all data, run: docker volume prune"
