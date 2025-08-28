#!/bin/bash

# Parseable + MinIO Setup Script
echo "🚀 Setting up Parseable + MinIO environment..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker-compose > /dev/null 2>&1 && ! docker compose version > /dev/null 2>&1; then
    echo "❌ Docker Compose is not available. Please install Docker Compose."
    exit 1
fi

echo "✅ Docker is running"

# Start the services
echo "📦 Starting Parseable and MinIO services..."
if command -v docker-compose > /dev/null 2>&1; then
    docker-compose up -d
else
    docker compose up -d
fi

echo "⏳ Waiting for services to be ready..."
sleep 30

# Check if services are running
echo "🔍 Checking service status..."
if command -v docker-compose > /dev/null 2>&1; then
    docker-compose ps
else
    docker compose ps
fi

# Wait for Parseable to be ready
echo "⏳ Waiting for Parseable to be ready..."
timeout=60
counter=0
while ! curl -s http://localhost:8000/api/v1/about > /dev/null; do
    sleep 2
    counter=$((counter + 2))
    if [ $counter -ge $timeout ]; then
        echo "❌ Parseable did not start within $timeout seconds"
        exit 1
    fi
done

echo "✅ Parseable is ready!"

# Wait for MinIO to be ready
echo "⏳ Waiting for MinIO to be ready..."
counter=0
while ! curl -s http://localhost:9000/minio/health/live > /dev/null; do
    sleep 2
    counter=$((counter + 2))
    if [ $counter -ge $timeout ]; then
        echo "❌ MinIO did not start within $timeout seconds"
        exit 1
    fi
done

echo "✅ MinIO is ready!"

# Create log streams in Parseable
echo "📝 Creating log streams in Parseable..."

# Create minio_audit stream
curl -X PUT "http://localhost:8000/api/v1/logstream/minio_audit" \
  -H "Authorization: Basic YWRtaW46YWRtaW4=" \
  -H "Content-Type: application/json" \
  -d '{}'

# Create minio_log stream
curl -X PUT "http://localhost:8000/api/v1/logstream/minio_log" \
  -H "Authorization: Basic YWRtaW46YWRtaW4=" \
  -H "Content-Type: application/json" \
  -d '{}'

echo "✅ Log streams created!"

# Initialize and tidy Go modules if go.mod exists
if [ -f "go.mod" ]; then
    echo "📦 Tidying Go modules..."
    go mod tidy
    echo "✅ Go modules updated!"
fi

echo ""
echo "🎉 Setup completed successfully!"
echo ""
echo "📊 Access Points:"
echo "   • Parseable UI: http://localhost:8000 (admin/admin)"
echo "   • MinIO Console: http://localhost:9001 (minioadmin/minioadmin)"
echo "   • MinIO API: http://localhost:9000"
echo ""
echo "🔧 Next steps:"
echo "   1. Run the go app to generate sample data and audit activity: go run minio_generate_sample_data.go"
echo "   2. Open Parseable dashboard to see audit logs"
echo "   3. Create custom dashboards and alerts"
echo ""
echo "📚 Useful commands:"
echo "   • Stop services: ./stop.sh"
echo "   • View logs: docker-compose logs -f"
