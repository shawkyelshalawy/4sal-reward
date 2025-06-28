# 4sal Reward System - High-Level System Design

## System Architecture Overview

```mermaid
graph TB
    %% External Systems
    Client[Client Applications<br/>Web/Mobile/API]
    Gemini[Google Gemini AI<br/>gemini-1.5-flash]
    
    %% Load Balancer & API Gateway
    LB[Load Balancer<br/>Nginx/HAProxy]
    
    %% Application Layer
    subgraph "Application Cluster"
        App1[Go App Instance 1<br/>Gin Framework]
        App2[Go App Instance 2<br/>Gin Framework]
        App3[Go App Instance N<br/>Gin Framework]
    end
    
    %% Caching Layer
    Redis[(Redis Cache<br/>Search Results<br/>Session Data)]
    
    %% Database Layer
    subgraph "Database Cluster"
        PG_Master[(PostgreSQL Master<br/>Read/Write)]
        PG_Replica1[(PostgreSQL Replica 1<br/>Read Only)]
        PG_Replica2[(PostgreSQL Replica 2<br/>Read Only)]
    end
    
    %% Monitoring & Logging
    subgraph "Observability"
        Logs[Structured Logging<br/>Zap Logger]
        Metrics[Metrics & Monitoring<br/>Prometheus/Grafana]
        Health[Health Checks<br/>Database/Redis/AI]
    end
    
    %% External Integrations
    subgraph "External Services"
        Payment[Payment Gateway<br/>Stripe/PayPal]
        Email[Email Service<br/>SendGrid/SES]
        Storage[File Storage<br/>S3/CloudStorage]
    end
    
    %% Connections
    Client --> LB
    LB --> App1
    LB --> App2
    LB --> App3
    
    App1 --> Redis
    App2 --> Redis
    App3 --> Redis
    
    App1 --> PG_Master
    App1 --> PG_Replica1
    App2 --> PG_Master
    App2 --> PG_Replica2
    App3 --> PG_Master
    App3 --> PG_Replica1
    
    App1 --> Gemini
    App2 --> Gemini
    App3 --> Gemini
    
    App1 --> Logs
    App2 --> Logs
    App3 --> Logs
    
    PG_Master --> PG_Replica1
    PG_Master --> PG_Replica2
    
    App1 --> Payment
    App1 --> Email
    App1 --> Storage
    
    classDef client fill:#e1f5fe
    classDef app fill:#f3e5f5
    classDef db fill:#e8f5e8
    classDef cache fill:#fff3e0
    classDef external fill:#fce4ec
    classDef monitoring fill:#f1f8e9
    
    class Client client
    class App1,App2,App3 app
    class PG_Master,PG_Replica1,PG_Replica2 db
    class Redis cache
    class Gemini,Payment,Email,Storage external
    class Logs,Metrics,Health monitoring
```

## Component Architecture

```mermaid
graph TB
    subgraph "Client Layer"
        Web[Web Application]
        Mobile[Mobile App]
        API[API Clients]
    end
    
    subgraph "API Gateway Layer"
        Gateway[API Gateway<br/>Rate Limiting<br/>Authentication<br/>Request Routing]
    end
    
    subgraph "Application Layer"
        subgraph "Handlers"
            AH[AI Handler<br/>Recommendations]
            CH[Credit Handler<br/>Packages & Purchases]
            PH[Product Handler<br/>Search & Redemption]
            AdminH[Admin Handler<br/>Management]
        end
        
        subgraph "Services"
            AS[AI Service<br/>Gemini Integration]
            CS[Credit Service<br/>Business Logic]
            PS[Product Service<br/>Search & Cache]
            US[User Service<br/>Profile Management]
        end
        
        subgraph "Repositories"
            CR[Credit Repository<br/>Package & Purchase Data]
            PR[Product Repository<br/>Product & Category Data]
            UR[User Repository<br/>User Data]
            CAR[Category Repository<br/>Category Data]
        end
    end
    
    subgraph "Infrastructure Layer"
        subgraph "Database"
            PostgreSQL[(PostgreSQL<br/>ACID Transactions<br/>Full-text Search)]
        end
        
        subgraph "Cache"
            RedisCache[(Redis<br/>Search Results<br/>Session Data)]
        end
        
        subgraph "External APIs"
            GeminiAPI[Google Gemini AI<br/>Smart Recommendations]
            PaymentAPI[Payment Gateway<br/>Credit Purchases]
        end
        
        subgraph "Monitoring"
            Logger[Structured Logging<br/>Zap]
            HealthCheck[Health Monitoring<br/>DB/Redis/AI Status]
        end
    end
    
    %% Connections
    Web --> Gateway
    Mobile --> Gateway
    API --> Gateway
    
    Gateway --> AH
    Gateway --> CH
    Gateway --> PH
    Gateway --> AdminH
    
    AH --> AS
    CH --> CS
    PH --> PS
    AdminH --> CS
    AdminH --> PS
    
    AS --> CAR
    AS --> UR
    AS --> PR
    CS --> CR
    CS --> UR
    PS --> PR
    
    CR --> PostgreSQL
    PR --> PostgreSQL
    UR --> PostgreSQL
    CAR --> PostgreSQL
    
    PS --> RedisCache
    AS --> GeminiAPI
    CS --> PaymentAPI
    
    AH --> Logger
    CH --> Logger
    PH --> Logger
    AdminH --> Logger
    
    PostgreSQL --> HealthCheck
    RedisCache --> HealthCheck
    GeminiAPI --> HealthCheck
```

## Data Flow Architecture

```mermaid
sequenceDiagram
    participant Client
    participant API as API Gateway
    participant Handler as AI Handler
    participant Service as AI Service
    participant UserRepo as User Repository
    participant CategoryRepo as Category Repository
    participant ProductRepo as Product Repository
    participant Gemini as Google Gemini AI
    participant Cache as Redis Cache
    participant DB as PostgreSQL
    participant Logger as Zap Logger
    
    Note over Client,Logger: AI Recommendation Flow
    
    Client->>API: POST /ai/recommendation
    API->>Handler: Route request
    Handler->>Logger: Log request start
    
    Handler->>UserRepo: Get user by ID
    UserRepo->>DB: SELECT user data
    DB-->>UserRepo: User data
    UserRepo-->>Handler: User with point balance
    
    Handler->>CategoryRepo: Get all categories
    CategoryRepo->>DB: SELECT categories
    DB-->>CategoryRepo: Category list
    CategoryRepo-->>Handler: Categories for AI
    
    Handler->>Service: Get AI recommendation
    Service->>Gemini: Send prompt with user data
    Gemini-->>Service: AI recommendation JSON
    Service-->>Handler: Parsed recommendation
    
    Handler->>CategoryRepo: Get category details
    CategoryRepo->>DB: SELECT category by ID
    DB-->>CategoryRepo: Category details
    CategoryRepo-->>Handler: Category info
    
    Handler->>ProductRepo: Get products by category
    ProductRepo->>Cache: Check cache
    Cache-->>ProductRepo: Cache miss
    ProductRepo->>DB: SELECT products
    DB-->>ProductRepo: Product list
    ProductRepo->>Cache: Store in cache
    ProductRepo-->>Handler: Category products
    
    Handler->>Logger: Log success
    Handler-->>API: Structured response
    API-->>Client: JSON response with recommendation + products
```

## Database Schema Design

```mermaid
erDiagram
    USERS {
        uuid id PK
        varchar email UK
        varchar name
        integer point_balance
        timestamp created_at
        timestamp updated_at
    }
    
    CREDIT_PACKAGES {
        uuid id PK
        varchar name
        text description
        decimal price
        integer reward_points
        boolean is_active
        timestamp created_at
        timestamp updated_at
    }
    
    CREDIT_PURCHASES {
        uuid id PK
        uuid user_id FK
        uuid credit_package_id FK
        decimal amount_paid
        integer points_awarded
        timestamp purchase_date
        varchar status
    }
    
    CATEGORIES {
        uuid id PK
        varchar name
        text description
        timestamp created_at
    }
    
    PRODUCTS {
        uuid id PK
        varchar name
        text description
        uuid category_id FK
        integer point_cost
        integer stock_quantity
        boolean is_active
        boolean is_in_offer_pool
        varchar image_url
        timestamp created_at
        timestamp updated_at
    }
    
    POINT_REDEMPTIONS {
        uuid id PK
        uuid user_id FK
        uuid product_id FK
        integer points_used
        integer quantity
        timestamp redemption_date
        varchar status
    }
    
    USERS ||--o{ CREDIT_PURCHASES : "purchases"
    USERS ||--o{ POINT_REDEMPTIONS : "redeems"
    CREDIT_PACKAGES ||--o{ CREDIT_PURCHASES : "purchased_as"
    CATEGORIES ||--o{ PRODUCTS : "contains"
    PRODUCTS ||--o{ POINT_REDEMPTIONS : "redeemed_as"
```

## Technology Stack

### Backend Framework
- **Language**: Go 1.22
- **Framework**: Gin (HTTP router)
- **Architecture**: Clean Architecture with layers

### Database & Storage
- **Primary Database**: PostgreSQL 15
  - ACID transactions
  - Full-text search with GIN indexes
  - Connection pooling
- **Cache**: Redis 7
  - Search result caching
  - Session storage
  - TTL-based expiration

### AI Integration
- **Provider**: Google Gemini 1.5 Flash
- **Use Case**: Smart product recommendations
- **Fallback**: Rule-based recommendations
- **Timeout**: 30 seconds with graceful degradation

### Infrastructure
- **Containerization**: Docker & Docker Compose
- **Logging**: Structured logging with Zap
- **Health Checks**: Database, Redis, and AI service monitoring
- **Environment**: Environment-based configuration

### External Integrations
- **AI**: Google Generative Language API
- **Future**: Payment gateways, email services, file storage

## Scalability Features

### Performance Optimizations
1. **Database Indexing**
   - GIN indexes for full-text search
   - B-tree indexes on foreign keys
   - Composite indexes for common queries

2. **Caching Strategy**
   - Redis for search results (5-minute TTL)
   - Cache invalidation on product updates
   - Connection pooling for database

3. **Horizontal Scaling**
   - Stateless application design
   - Load balancer ready
   - Database read replicas support

### Reliability Features
1. **Error Handling**
   - Structured error responses
   - Graceful degradation for AI failures
   - Comprehensive logging

2. **Health Monitoring**
   - Database connection health
   - Redis availability checks
   - AI service status monitoring

3. **Transaction Safety**
   - ACID transactions for critical operations
   - Optimistic locking for concurrent updates
   - Rollback mechanisms

## Security Considerations

### Current Implementation
- Input validation and sanitization
- SQL injection prevention
- Structured error handling without information leakage
- Environment-based configuration

### Production Recommendations
- JWT authentication
- API rate limiting
- HTTPS/TLS encryption
- API key management
- Request logging and monitoring
- Role-based access control (RBAC)

## Deployment Architecture

### Development Environment
```bash
docker-compose up --build
```

### Production Environment
- **Container Orchestration**: Kubernetes/Docker Swarm
- **Load Balancing**: Nginx/HAProxy
- **Database**: Managed PostgreSQL (AWS RDS/Google Cloud SQL)
- **Cache**: Managed Redis (AWS ElastiCache/Google Memorystore)
- **Monitoring**: Prometheus + Grafana
- **Logging**: ELK Stack or similar

### Scaling Strategy
1. **Horizontal Scaling**: Multiple application instances
2. **Database Scaling**: Read replicas for read-heavy operations
3. **Cache Scaling**: Redis cluster for high availability
4. **CDN**: Static asset delivery optimization
5. **Auto-scaling**: Based on CPU/memory metrics