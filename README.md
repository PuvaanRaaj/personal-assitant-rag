# RAG Personal Assistant

A production-grade, AI-powered personal assistant that answers questions using your uploaded documents (PDFs, Markdown, JSON, etc.) with **Retrieval-Augmented Generation (RAG)**.

Built with **Golang backend**, **Next.js frontend**, **Qdrant vector database**, and **AWS S3 storage**. Fully Dockerized with LocalStack for local testing.

## ✨ Features

- 📄 **Multi-format Support**: Upload PDFs, Markdown, JSON, CSV, and text files
- 🔍 **Semantic Search**: Find relevant information across all your documents
- 💬 **Intelligent Q&A**: Get accurate answers with source citations
- 🔒 **Secure**: JWT authentication, user data isolation, rate limiting
- 🎨 **Beautiful UI**: Modern dark mode interface with glassmorphism
- ⚡ **High Performance**: Golang backend handles 100+ concurrent users
- 🐳 **Fully Dockerized**: One command to run locally with LocalStack S3

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- OpenAI API key

### Run Locally

```bash
# 1. Clone and configure
cp .env.example .env
# Edit .env and add your OPENAI_API_KEY

# 2. Start all services (Postgres, Qdrant, LocalStack S3, Backend, Frontend)
docker-compose up -d

# 3. Access the application
# Frontend: http://localhost:3000
# Backend API: http://localhost:8080
# Qdrant Dashboard: http://localhost:6333/dashboard
```

## 📦 What's Included

- **Backend**: Golang API with RAG implementation
- **Frontend**: Next.js with beautiful UI
- **Vector DB**: Qdrant for semantic search
- **Database**: PostgreSQL for metadata
- **Storage**: AWS S3 (LocalStack for local dev)
- **Auth**: JWT authentication
- **Docker**: Complete containerized stack

For complete implementation details, see [Implementation Plan](file:///.gemini/antigravity/brain/deafef04-9412-4435-967f-b46704d1ea34/implementation_plan.md).
