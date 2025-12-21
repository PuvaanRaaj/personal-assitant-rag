# Testing Guide for RAG Personal Assistant

## Container Testing with LocalStack

The entire application can be tested locally using Docker containers with LocalStack simulating AWS S3.

### Complete Stack Testing

1. **Start all containers**:

```bash
docker-compose up -d
```

This starts:

- `backend` - Golang API server (port 8080)
- `frontend` - Next.js application (port 3000)
- `postgres` - PostgreSQL database (port 5432)
- `qdrant` - Vector database (port 6333)
- `localstack` - S3-compatible storage (port 4566)

2. **Verify all services are healthy**:

```bash
docker-compose ps

# All services should show "healthy" or "running"
```

3. **Check LocalStack S3 bucket**:

```bash
docker-compose exec localstack awslocal s3 ls

# Should output: rag-assistant-uploads
```

### Testing S3 Upload Flow

**Test file upload to LocalStack**:

```bash
# Create a test file
echo "This is a test document for RAG testing" > test.txt

# Upload via API
curl -X POST http://localhost:8080/api/documents/upload \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@test.txt"

# Verify file in LocalStack
docker-compose exec localstack awslocal s3 ls s3://rag-assistant-uploads/ --recursive
```

### Testing RAG Query Flow

**End-to-end RAG test**:

```bash
# 1. Register a user
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# 2. Login to get JWT
TOKEN=$(curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  | jq -r '.token')

# 3. Upload a document
curl -X POST http://localhost:8080/api/documents/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@your-document.pdf"

# 4. Query the document
curl -X POST http://localhost:8080/api/query \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"question":"What is this document about?"}'
```

### Backend Unit Tests

```bash
# Run all backend tests
cd backend
go test ./internal/... -v -cover

# Test specific package
go test ./internal/service -v

# Run tests with race detection
go test -race ./...
```

### Frontend Tests

```bash
cd frontend
npm run test

# Watch mode
npm run test:watch

# E2E tests (requires running backend)
npm run test:e2e
```

### Integration Tests

```bash
# Run integration tests (requires docker-compose running)
cd backend
go test ./tests/integration/... -v
```

### Load Testing

**Using k6** (install from https://k6.io/):

```javascript
// load-test.js
import http from "k6/http";
import { check } from "k6";

export let options = {
  vus: 50, // 50 virtual users
  duration: "30s", // for 30 seconds
};

export default function () {
  const payload = JSON.stringify({
    question: "What is RAG?",
  });

  const params = {
    headers: {
      "Content-Type": "application/json",
      Authorization: "Bearer YOUR_TOKEN",
    },
  };

  let res = http.post("http://localhost:8080/api/query", payload, params);
  check(res, { "status is 200": (r) => r.status === 200 });
}
```

Run with: `k6 run load-test.js`

### Security Testing

**1. Test JWT validation**:

```bash
# Should fail without token
curl -X GET http://localhost:8080/api/documents

# Should fail with invalid token
curl -X GET http://localhost:8080/api/documents \
  -H "Authorization: Bearer invalid_token"
```

**2. Test rate limiting**:

```bash
# Send 150 requests rapidly (limit is 100/min)
for i in {1..150}; do
  curl http://localhost:8080/api/health &
done
wait

# Should see 429 Too Many Requests after 100 requests
```

**3. Test file upload validation**:

```bash
# Try uploading a .exe file (should be rejected)
curl -X POST http://localhost:8080/api/documents/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@malicious.exe"

# Should return 400 Bad Request
```

### Debugging

**View logs**:

```bash
# Backend logs
docker-compose logs -f backend

# All services
docker-compose logs -f

# Specific service
docker-compose logs -f localstack
```

**Access containers**:

```bash
# Backend shell
docker-compose exec backend sh

# Check S3 from inside LocalStack
docker-compose exec localstack sh
awslocal s3 ls s3://rag-assistant-uploads/
```

**Database debugging**:

```bash
# Connect to PostgreSQL
docker-compose exec postgres psql -U rag_user -d rag_assistant

# Show tables
\dt

# Query users
SELECT * FROM users;
```

## Production Testing

### AWS S3 Testing

Before deploying to production, test with real AWS S3:

1. **Update `.env`**:

```bash
# Remove AWS_ENDPOINT (don't use LocalStack)
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-real-key
AWS_SECRET_ACCESS_KEY=your-real-secret
S3_BUCKET=your-test-bucket
```

2. **Create test bucket**:

```bash
aws s3 mb s3://your-test-bucket
aws s3api put-bucket-encryption \
  --bucket your-test-bucket \
  --server-side-encryption-configuration '{
    "Rules": [{
      "ApplyServerSideEncryptionByDefault": {
        "SSEAlgorithm": "AES256"
      }
    }]
  }'
```

3. **Run tests against real S3**:

```bash
docker-compose up -d backend
# Run same tests as above
```

### Performance Benchmarks

**Expected performance**:

- ✅ Health check: <10ms
- ✅ Document upload: <500ms (20-page PDF)
- ✅ Embedding generation: 1-3s per document
- ✅ Query response: 500ms-2s (including LLM call)
- ✅ Concurrent users: 100+ with 2 vCPU

**Measure**:

```bash
# Time a query
time curl -X POST http://localhost:8080/api/query \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"question":"Test question"}'
```

## Cleanup

```bash
# Stop all containers
docker-compose down

# Remove volumes (WARNING: deletes all data)
docker-compose down -v

# Remove images
docker-compose down --rmi all
```

## CI/CD Testing

Example GitHub Actions workflow:

```yaml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Start services
        run: docker-compose up -d

      - name: Wait for services
        run: sleep 30

      - name: Run backend tests
        run: |
          cd backend
          go test ./... -v

      - name: Health check
        run: curl -f http://localhost:8080/health

      - name: Cleanup
        run: docker-compose down
```
