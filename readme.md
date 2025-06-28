# 4sal Reward System

A modern, scalable reward system built with Go, designed for managing users, credit packages, products, and AI-powered product recommendations. This system supports hundreds of thousands of products with fast search capabilities and intelligent recommendations.

---

## 🚀 Features

- **Credit Package Management**: Users can purchase credit packages and earn reward points
- **Product Redemption**: Redeem products using earned points from the offer pool
- **Advanced Search**: Fast, full-text search across product names and descriptions with pagination
- **AI Recommendations**: Intelligent product suggestions based on user's point balance and available categories
- **Admin Panel**: Complete CRUD operations for packages and products
- **Scalable Architecture**: Built with performance and scalability in mind
- **Docker Support**: Fully containerized with Docker Compose
- **Caching**: Redis integration for improved performance
- **Health Monitoring**: Built-in health checks for all services

---

## 🏗️ Architecture

### Tech Stack
- **Backend**: Go 1.22 with Gin framework
- **Database**: PostgreSQL 15 with full-text search
- **Cache**: Redis 7 for search result caching
- **AI Integration**: OpenAI GPT-3.5-turbo for recommendations
- **Containerization**: Docker & Docker Compose
- **Testing**: Go testing with sqlmock

### Project Structure
```
4sal-reward/
├── cmd/                    # Application entry point
├── internal/
│   ├── domain/models/      # Domain models
│   ├── handlers/           # HTTP handlers
│   ├── repositories/       # Data access layer
│   ├── services/           # Business logic
│   └── infrastructure/     # External services (DB, Redis, Logger)
├── docker-compose.yml      # Container orchestration
├── Dockerfile             # Application container
└── api_documentation.md   # Complete API docs
```

---

## 🚀 Quick Start

### Prerequisites
- [Docker](https://www.docker.com/products/docker-desktop) and [Docker Compose](https://docs.docker.com/compose/)
- (Optional) [Go 1.22+](https://golang.org/dl/) for local development

### Setup & Run

1. **Clone the repository:**
   ```bash
   git clone https://github.com/yourusername/4sal-reward.git
   cd 4sal-reward
   ```

2. **Configure environment (optional):**
   ```bash
   cp .env.example .env
   # Edit .env to add your OpenAI API key for AI recommendations
   ```

3. **Start all services:**
   ```bash
   docker-compose up --build
   ```

4. **Verify the setup:**
   ```bash
   curl http://localhost:8080/health
   ```

The API will be available at `http://localhost:8080` with sample data pre-loaded.

---

## 📚 API Documentation

### Core Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/credits/packages` | List credit packages (paginated) |
| `POST` | `/credits/purchase` | Purchase credit package |
| `GET` | `/products/search` | Search products (full-text) |
| `POST` | `/products/redeem` | Redeem product with points |
| `POST` | `/ai/recommendation` | Get AI-powered recommendations |

### Admin Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/admin/packages` | Create credit package |
| `PUT` | `/admin/packages/:id` | Update credit package |
| `POST` | `/admin/products` | Create product |
| `PUT` | `/admin/products/:id` | Update product |
| `PUT` | `/admin/products/:id/offer-status` | Update offer status |

For detailed API documentation with request/response examples, see [api_documentation.md](api_documentation.md).

---

## 🧪 Testing

### Run Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -v -cover ./...

# Run specific package tests
go test ./internal/repositories/...
```

### Test Coverage
The project includes comprehensive unit tests for:
- Repository layer with sqlmock
- Service layer with mock repositories
- Handler layer integration tests

---

## 🤖 AI Integration

### How It Works
1. User requests recommendation via `/ai/recommendation`
2. System fetches user's point balance and available categories
3. Sends structured prompt to OpenAI GPT-3.5-turbo
4. AI returns recommended category ID and point range
5. System provides intelligent product suggestions

### Sample AI Prompt
```
Given a user with a current point balance of 1500 and the following available categories:
[{"id":"1a2b3c4d-...","name":"Electronics"}, {"id":"1a2b3c4d-...","name":"Books"}]

Suggest the most suitable category by its id and a corresponding minimum and maximum point range for products within that category.

Response format: {"recommended_category_id": "string", "min_points_llm": integer, "max_points_llm": integer, "reasoning": "string"}
```

### Fallback Strategy
If OpenAI is unavailable, the system falls back to rule-based recommendations ensuring 100% uptime.

---

## 🔍 Search Implementation

### Full-Text Search Features
- **PostgreSQL GIN indexes** for fast text search
- **Multi-field search** across product names and descriptions
- **Pagination support** for large datasets
- **Redis caching** for frequently searched terms
- **Offer pool filtering** (only redeemable products)

### Performance Optimizations
- Indexed search fields with `to_tsvector` and `to_tsquery`
- Redis caching with 5-minute TTL
- Connection pooling for database
- Efficient pagination with LIMIT/OFFSET

---

## 📊 Sample Data

The system comes with pre-loaded test data:

### Users
- **Alice Smith**: 1,500 points
- **Bob Johnson**: 750 points  
- **Admin User**: 10,000 points

### Categories
- Electronics, Books, Gift Cards, Home Goods

### Products
- Wireless Earbuds (500 pts), E-Reader (800 pts), Gift Cards (1000 pts), etc.

### Credit Packages
- Bronze (100 pts/$10), Silver (300 pts/$25), Gold (750 pts/$50), Platinum (1500 pts/$100)

---

## 🐳 Docker Configuration

### Services
- **app**: Go application (port 8080)
- **postgres**: PostgreSQL database (port 5432)
- **redis**: Redis cache (port 6379)
- **migrate**: Database migration runner

### Health Checks
All services include health checks for reliable startup and monitoring.

---

## 🔧 Development

### Local Development Setup
```bash
# Install dependencies
go mod download

# Run database migrations
make migrate_up

# Start the server
make server

# Run tests
make test
```

### Environment Variables
```bash
DATABASE_URL=postgresql://admin:secret@localhost:5432/rewarddb?sslmode=disable
REDIS_ADDR=localhost:6379
OPENAI_API_KEY=your_openai_api_key_here
GIN_MODE=release
```

---

## 🚀 Scalability Features

### Database Optimizations
- **Indexed searches** with GIN indexes for full-text search
- **Connection pooling** (25 max connections)
- **Read replicas ready** architecture
- **Efficient pagination** for large datasets

### Caching Strategy
- **Search result caching** with Redis
- **Cache invalidation** on product updates
- **TTL-based expiration** (5 minutes)

### Performance Monitoring
- **Structured logging** with Zap
- **Request timing** and metrics
- **Health check endpoints**
- **Database connection monitoring**

---

## 🔒 Security Considerations

### Current Implementation
- Input validation and sanitization
- SQL injection prevention with parameterized queries
- Error handling without information leakage
- Health check endpoints for monitoring

### Production Recommendations
- Add JWT authentication
- Implement rate limiting
- Add API key management
- Enable HTTPS/TLS
- Add request logging and monitoring

---

## 🚀 Deployment

### Production Deployment
```bash
# Build for production
docker-compose -f docker-compose.prod.yml up --build

# Scale the application
docker-compose up --scale app=3
```

### Environment Configuration
- Set `OPENAI_API_KEY` for AI features
- Configure `DATABASE_URL` for external database
- Set `REDIS_ADDR` for external Redis
- Enable `GIN_MODE=release` for production

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🆘 Troubleshooting

### Common Issues

**Port conflicts:**
```bash
# Check if ports are in use
lsof -i :8080
lsof -i :5432
lsof -i :6379
```

**Database connection issues:**
```bash
# Check database logs
docker-compose logs postgres

# Verify database is ready
docker-compose exec postgres pg_isready -U admin -d rewarddb
```

**Redis connection issues:**
```bash
# Check Redis logs
docker-compose logs redis

# Test Redis connection
docker-compose exec redis redis-cli ping
```

**Migration failures:**
```bash
# Check migration logs
docker-compose logs migrate

# Manual migration
docker-compose exec app ./reward-system migrate up
```

---

## 📞 Support

For questions and support:
- Create an issue in the GitHub repository
- Check the [API documentation](api_documentation.md)
- Review the troubleshooting section above

---

**Built with ❤️ for 4sal by the development team**