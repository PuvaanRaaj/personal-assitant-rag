#!/bin/bash

# Initialize LocalStack S3 bucket for local development
echo "Initializing LocalStack S3..."

# Wait for LocalStack to be ready
while ! awslocal s3 ls > /dev/null 2>&1; do
  echo "Waiting for LocalStack to be ready..."
  sleep 2
done

# Create the S3 bucket
awslocal s3 mb s3://rag-assistant-uploads 2>/dev/null || echo "Bucket already exists"

# Enable versioning (optional but recommended)
awslocal s3api put-bucket-versioning \
  --bucket rag-assistant-uploads \
  --versioning-configuration Status=Enabled

# Set bucket encryption
awslocal s3api put-bucket-encryption \
  --bucket rag-assistant-uploads \
  --server-side-encryption-configuration '{
    "Rules": [{
      "ApplyServerSideEncryptionByDefault": {
        "SSEAlgorithm": "AES256"
      }
    }]
  }'

echo "LocalStack S3 bucket 'rag-assistant-uploads' initialized successfully!"
