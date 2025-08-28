# Parseable + MinIO Real-Time Audit Logging Pipeline

This repository demonstrates a complete, production-ready integration between **MinIO object storage** and **Parseable log analytics platform** for real-time audit logging and monitoring.

## What This Is

A complete observability solution that captures every MinIO operation (uploads, downloads, bucket operations, errors) and streams them in real-time to Parseable for analysis, alerting, and compliance monitoring.

## Architecture Overview

```
┌─────────────────┐    Real-time    ┌─────────────────┐    S3 Storage   ┌─────────────────┐
│   MinIO Server  │────Webhooks────▶│   Parseable     │──────────────▶│  Object Storage │
│                 │    (JSON logs)  │   Platform      │   (log data)   │   (MinIO/S3)    │
│  • File Ops     │                 │                 │                │                 │
│  • API Calls    │                 │  • SQL Queries  │                │  • Persistent   │
│  • Audit Logs   │                 │  • Dashboards   │                │  • Scalable     │
│  • Errors       │                 │  • Alerting     │                │  • Searchable   │
└─────────────────┘                 └─────────────────┘                └─────────────────┘
```

### How It Works

1. **MinIO Operations**: Every API call to MinIO (PutObject, GetObject, ListObjects, etc.) generates audit events
2. **Real-time Streaming**: MinIO's webhook feature sends these audit logs immediately to Parseable
3. **Structured Ingestion**: Parseable receives JSON-formatted audit logs and automatically creates searchable schema
4. **Persistent Storage**: Logs are stored in S3-compatible storage for long-term retention and analysis
5. **Rich Analytics**: Use SQL queries, dashboards, and alerts to monitor operations, security, and performance

## Key Features

### Real-Time Audit Logging
- **Different API operation types** (GetObject, PutObject, ListObjects, etc.)
- **Zero-latency** streaming from MinIO to Parseable via webhooks

### Production-Ready Components
- **MinIO**: S3-compatible object storage with audit webhook configuration
- **Parseable**: High-performance log analytics with SQL interface
- **Docker Compose**: Orchestrated multi-service deployment
- **Demo Application**: Node.js app simulating real-world file operations

### Rich Analytics Capabilities
- **SQL Queries**: Query audit logs with full SQL syntax
- **Schema Discovery**: Automatic field extraction and indexing
- **Time-series Analysis**: Timestamp-based filtering and aggregation
- **Dashboard Ready**: Pre-built query examples for common use cases

## What Gets Logged

Every MinIO operation generates detailed audit logs including:

```json
{
  "api_name": "PutObject",
  "api_object": "my-file.txt",
  "api_statusCode": 200,
  "api_inputBytes": 15126,
  "remotehost": "192.168.1.100",
  "userAgent": "minio-js/7.1.3",
  "requestID": "REQUEST-ABC123",
  "time": "2025-08-27T09:45:15.123Z",
  "responseHeader_etag": "\"9a0364b9e99bb480dd25e1f0284c8555\"",
  "requestHeader_authorization": "AWS4-HMAC-SHA256...",
  // ...
}
```

### Captured Operations Include:
- **File Operations**: Upload, download, delete, metadata operations
- **Bucket Management**: Create, list, policy changes, location queries
- **Security Events**: Authentication, authorization, access patterns
- **Error Conditions**: 404s, permission denied, invalid requests
- **Performance Metrics**: Response times, data transfer sizes

## Use Cases

### Security & Compliance
- **Audit Trail**: Complete record of all data access and modifications
- **Anomaly Detection**: Unusual access patterns, failed authentications
- **Compliance Reporting**: SOX, GDPR, HIPAA audit requirements
- **Data Governance**: Track data lineage and access patterns

### Operations & Performance
- **Capacity Planning**: Monitor storage growth and usage patterns
- **Performance Monitoring**: Response times, error rates, throughput
- **Cost Optimization**: Identify heavy users, optimize storage tiers
- **SLA Monitoring**: Track availability and performance metrics

### Alerting & Monitoring
- **Real-time Alerts**: High error rates, unusual access patterns
- **Threshold Monitoring**: Storage quotas, performance degradation
- **Security Notifications**: Failed authentications, policy violations
- **Operational Health**: Service availability, backup status

## Quick Start

### Option 1: Full Setup (Recommended)

Get the complete pipeline running in under 2 minutes:

```bash
# Clone and setup
git clone <this-repo>
cd minio-parseable-audit-logging

# Start the entire pipeline
chmod +x *.sh
./setup.sh

# Generate sample data
npm start
```

### Option 2: Manual Step-by-Step

If you prefer to understand each component:

```bash
# 1. Start services
docker compose up -d

# 2. Wait for services (15 seconds)
sleep 15

# 3. Create log streams
curl -X PUT "http://localhost:8000/api/v1/logstream/minio_audit" \
  -H "Authorization: Basic YWRtaW46YWRtaW4=" \
  -H "Content-Type: application/json" -d '{}'

# 4. Configure MinIO audit webhook
docker exec minio mc alias set minio http://localhost:9000 minioadmin minioadmin
docker exec minio mc admin config set minio audit_webhook:parseable \
  endpoint=http://parseable:8000/api/v1/logstream/minio_audit \
  auth_token="Basic YWRtaW46YWRtaW4=" enable=on
docker compose restart minio

# 5. Generate test data
npm install && npm start
```

## Access Points

Once setup is complete, you can access:

- **Parseable Dashboard**: <http://localhost:8000>
  - Username: `admin`
  - Password: `admin`

- **MinIO Console**: <http://localhost:9001>
  - Username: `minioadmin`
  - Password: `minioadmin`

- **MinIO API**: <http://localhost:9000>

## Sample Queries & Dashboards

### Operations Dashboard

```sql
-- Most common operations
SELECT api_name, COUNT(*) as operations
FROM minio_audit
GROUP BY api_name
ORDER BY operations DESC

-- File upload/download activity
SELECT
  DATE(time) as date,
  api_name,
  COUNT(*) as operations,
  SUM("api_inputBytes") as bytes_transferred
FROM minio_audit
WHERE api_name IN ('PutObject', 'GetObject')
GROUP BY DATE(time), api_name
ORDER BY date DESC
```

### Security Monitoring

```sql
-- Error analysis
SELECT
  "api_statusCode" as status_code,
  COUNT(*) as error_count,
  api_name
FROM minio_audit
WHERE "api_statusCode" >= 400
GROUP BY "api_statusCode", api_name
ORDER BY error_count DESC

-- Access patterns by IP
SELECT
  remotehost,
  COUNT(*) as requests,
  COUNT(DISTINCT api_object) as unique_objects
FROM minio_audit
GROUP BY remotehost
ORDER BY requests DESC
```

### Performance Analysis

```sql
-- Data transfer analysis
SELECT
  api_name,
  COUNT(*) as operations,
  AVG("api_inputBytes") as avg_input_bytes,
  SUM("api_inputBytes") as total_bytes
FROM minio_audit
WHERE "api_inputBytes" > 0
GROUP BY api_name

-- Response time analysis (if available)
SELECT
  api_name,
  AVG("api_timeToResponse") as avg_response_time,
  MAX("api_timeToResponse") as max_response_time
FROM minio_audit
WHERE "api_timeToResponse" IS NOT NULL
GROUP BY api_name
```

## Management Commands

```bash

# Stop services
./stop.sh
```

## Technical Implementation

### MinIO Audit Webhook Configuration

The integration uses MinIO's built-in audit webhook feature:

```bash
# Webhook endpoint configuration
endpoint=http://parseable:8000/api/v1/logstream/minio_audit
auth_token="Basic YWRtaW46YWRtaW4="
enable=on
```

### Parseable Ingestion

Logs are ingested via HTTP POST to the logstream endpoint:

```bash
POST /api/v1/logstream/minio_audit
Authorization: Basic YWRtaW46YWRtaW4=
Content-Type: application/json

[{audit_log_json}]
```

### Schema Auto-Discovery

Parseable automatically discovers and indexes fields:

- **Nested JSON flattening**: `api.name` becomes `api_name`
- **Type inference**: Automatic detection of numbers, strings, timestamps
- **Timestamp handling**: Both `time` and `p_timestamp` for flexibility
- **Case-sensitive fields**: Some fields like `"api_statusCode"` require quotes in queries

## Monitoring

### Check Service Status

```bash
# Docker Compose
docker compose ps

# Or Docker directly
docker ps
```

### View Real-time Logs

```bash
# All services
./logs.sh

# Specific service
docker compose logs -f parseable
docker compose logs -f minio
```

### Verify Audit Log Flow

```bash
# Check log count
curl -H "Authorization: Basic YWRtaW46YWRtaW4="
  "http://localhost:8000/api/v1/query"
  -H "Content-Type: application/json"
  -d '{"query": "SELECT count(*) FROM minio_audit", "startTime": "2025-08-27T00:00:00Z", "endTime": "2025-08-28T23:59:59Z"}'

# Generate test activity
npm start

# Check again for new logs
```

## Troubleshooting

### Services Won't Start

```bash
# Check if ports are in use
lsof -i :8000  # Parseable
lsof -i :9000  # MinIO API
lsof -i :9001  # MinIO Console

# Clean up and restart
docker compose down
docker system prune -f
./setup.sh
```

### No Audit Logs Appearing

1. Verify webhook configuration:

```bash
docker exec minio mc admin config get minio audit_webhook:parseable
```

2. Check MinIO logs for webhook errors:

```bash
docker compose logs minio | grep -i webhook
```

3. Ensure Parseable log stream exists:

```bash
curl -H "Authorization: Basic YWRtaW46YWRtaW4="
  "http://localhost:8000/api/v1/logstream"
```

4. Manually test webhook endpoint:

```bash
curl -X POST "http://localhost:8000/api/v1/logstream/minio_audit"
  -H "Authorization: Basic YWRtaW46YWRtaW4="
  -H "Content-Type: application/json"
  -d '[{"test": "log", "time": "'$(date -u +%Y-%m-%dT%H:%M:%S.%3NZ)'"}]'
```

### Demo App Connections

```bash
# Check MinIO is accessible
curl http://localhost:9000/minio/health/live

# Check Parseable is accessible
curl http://localhost:8000/api/v1/about

# Reinstall dependencies
rm -rf node_modules package-lock.json
npm install
```

## Production Deployment

### Scaling Considerations

1. **Multiple MinIO Nodes**: Configure audit webhooks on all MinIO instances
2. **Parseable Clustering**: Use distributed mode for high-volume logging
3. **Load Balancing**: Distribute webhook traffic across multiple Parseable ingest nodes
4. **Storage**: Ensure adequate S3 storage for log retention requirements

### Security Hardening

```bash
# Use proper authentication tokens
PARSEABLE_AUTH=$(echo -n "username:password" | base64)

# Configure TLS/SSL for production
# Use secrets management for credentials
# Implement network security policies
```

### Monitoring & Alerting

Set up alerts for:

- High error rates in MinIO operations
- Webhook delivery failures
- Storage quota thresholds
- Unusual access patterns
- Performance degradation

## Performance Optimization

### MinIO Configuration

```bash
# Optimize webhook settings
  mc admin config set minio audit_webhook:parseable
  batch_size=100
  queue_size=1000000
  max_retry=3
```

### Parseable Configuration

```bash
# Increase ingest performance
P_STAGING_DIR=/fast-ssd/staging
P_CACHE_SIZE=2GB
P_MAX_CONCURRENT_INGESTS=10
```

## Cleanup

To completely remove everything:

```bash
# Stop services
./stop.sh

# Remove containers and volumes
docker compose down -v

# Remove images (optional)
docker rmi parseable/parseable:latest minio/minio:latest minio/mc:latest

# Clean up local directories
rm -rf downloads/ uploads/
```

## Additional Resources

- [Parseable Documentation](https://parseable.io/docs)
- [MinIO Audit Webhook Guide](https://docs.min.io/docs/minio-audit-quickstart-guide.html)
- [Docker Compose Reference](https://docs.docker.com/compose/)
- [SQL Query Examples for Log Analysis](https://parseable.io/docs/sql)



## 📁 File Structure

```
parseable/
├── docker-compose.yaml                    # Full distributed setup
├── parseable-env                          # Environment variables for standalone
├── minio-generate-sample-data.js          # Sample application
├── setup.sh                               # Full setup script
├── stop.sh                                # Full stop script
└── README.md                              # This file
```

## 🧹 Cleanup

To completely remove everything:

```bash
# Stop services
./stop.sh

# Remove containers and volumes
docker-compose down -v

# Remove images (optional)
docker rmi parseable/parseable:latest minio/minio:latest minio/mc:latest

# Clean up local directories
rm -rf /tmp/parseable /tmp/minio-data
```

## Learn More

- [Parseable Documentation](https://parseable.io/docs)
- [MinIO Documentation](https://docs.min.io/)
- [Example Dashboard Configurations](https://github.com/parseablehq/parseable/tree/main/examples)
