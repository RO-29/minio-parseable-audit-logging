#!/bin/bash

# Standalone Parseable Setup Script
echo "🚀 Setting up standalone Parseable with MinIO..."

# Create staging directory
echo "📁 Creating staging directory..."
mkdir -p /tmp/parseable/staging

# Start MinIO first
echo "🗄️  Starting MinIO..."
docker run -d \
  --name minio-standalone \
  -p 9000:9000 \
  -p 9001:9001 \
  -e "MINIO_ROOT_USER=minioadmin" \
  -e "MINIO_ROOT_PASSWORD=minioadmin" \
  -v /tmp/minio-data:/data \
  minio/minio server /data --console-address ":9001"

# Wait for MinIO to start
echo "⏳ Waiting for MinIO to start..."
sleep 10

# Create bucket using mc client
echo "🪣 Creating bucket..."
docker run --rm \
  --network host \
  minio/mc mc alias set local http://localhost:9000 minioadmin minioadmin

docker run --rm \
  --network host \
  minio/mc mc mb local/parseable

# Start Parseable
echo "📊 Starting Parseable..."
docker run -d \
  --name parseable-standalone \
  -p 8000:8000 \
  --env-file parseable-env \
  -v /tmp/parseable/staging:/staging \
  parseable/parseable:latest \
  parseable s3-store

echo "⏳ Waiting for Parseable to start..."
sleep 15

# Create log streams
echo "📝 Creating log streams..."
curl -X PUT "http://localhost:8000/api/v1/logstream/minio_audit" \
  -H "Authorization: Basic YWRtaW46YWRtaW4=" \
  -H "Content-Type: application/json" \
  -d '{}'

curl -X PUT "http://localhost:8000/api/v1/logstream/minio_log" \
  -H "Authorization: Basic YWRtaW46YWRtaW4=" \
  -H "Content-Type: application/json" \
  -d '{}'

echo ""
echo "🎉 Standalone setup completed!"
echo ""
echo "📊 Access Points:"
echo "   • Parseable UI: http://localhost:8000 (admin/admin)"
echo "   • MinIO Console: http://localhost:9001 (minioadmin/minioadmin)"
echo ""
echo "🛑 To stop:"
echo "   docker stop parseable-standalone minio-standalone"
echo "   docker rm parseable-standalone minio-standalone"
