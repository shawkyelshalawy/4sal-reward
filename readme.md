# 4sal Reward System

A modern reward system for managing users, credit packages, products, and AI-powered product recommendations.

---

## Table of Contents

- [Overview](#overview)
- [Business Logic](#business-logic)
- [Architecture](#architecture)
- [Setup & Running (Docker Compose)](#setup--running-docker-compose)
- [API Endpoints](#api-endpoints)
  - [Admin Endpoints](#admin-endpoints)
  - [User Endpoints](#user-endpoints)
  - [AI Recommendation Endpoint](#ai-recommendation-endpoint)
  - [Health Check](#health-check)
- [Sample Test Data](#sample-test-data)
- [AI Prompt Explanation](#ai-prompt-explanation)
- [Troubleshooting](#troubleshooting)

---

## Overview

**4sal Reward System** is a backend service for a loyalty/reward platform.  
It allows users to purchase credit packages, redeem products using points, and get AI-powered product recommendations based on their point balance and available categories.

---

## Business Logic

- **Users** earn points by purchasing credit packages.
- **Products** can be redeemed using points if the user has enough balance.
- **Categories** organize products and are used for AI recommendations.
- **Admins** can create credit packages and products, and manage offer statuses.
- **AI Recommendation**: The system uses an LLM (Large Language Model) to suggest the best product category and point range for a user, based on their current point balance and available categories.(In progress)

---

## Architecture

- **Go (Golang)** backend (Gin framework)
- **PostgreSQL** for persistent storage
- **Redis** for caching
- **Docker Compose** for easy setup
- **LLM/AI** integration for smart recommendations

---

## Setup & Running (Docker Compose)

### Prerequisites

- [Docker](https://www.docker.com/products/docker-desktop) and [Docker Compose](https://docs.docker.com/compose/) installed

### Steps

1. **Clone the repository:**
   ```sh
   git clone https://github.com/yourusername/4sal-reward.git
   cd 4sal-reward
   ```

2. **Start all services:**
   ```sh
   docker-compose up --build
   ```

   This will:
   - Build and run the Go app
   - Start PostgreSQL and Redis
   - Run database migrations and seed sample data

3. **Access the API:**
   - The API will be available at: [http://localhost:8080](http://localhost:8080)
   - Health check: [http://localhost:8080/health](http://localhost:8080/health)

---

## API Endpoints

### Admin Endpoints

| Method | Endpoint                        | Description                        |
|--------|---------------------------------|------------------------------------|
| POST   | `/admin/packages`               | Create a new credit package        |
| POST   | `/admin/products`               | Create a new product               |
| PUT    | `/admin/products/:id/offer-status` | Update product offer status    |

#### Example: Create Credit Package

```http
POST /admin/packages
Content-Type: application/json

{
  "name": "Gold Bundle",
  "description": "Best value credits",
  "price": 99.99,
  "reward_points": 1000
}
```

#### Example: Create Product

```http
POST /admin/products
Content-Type: application/json

{
  "name": "Wireless Headphones",
  "description": "Premium headphones",
  "category_id": "1a2b3c4d-5e6f-7a8b-9c0d-1e2f00010000",
  "point_cost": 500,
  "stock_quantity": 20,
  "is_in_offer_pool": true,
  "image_url": "https://example.com/headphones.jpg"
}
```

---

### User Endpoints

| Method | Endpoint                | Description                        |
|--------|-------------------------|------------------------------------|
| POST   | `/credits/purchase`     | Purchase a credit package          |
| POST   | `/products/redeem`      | Redeem a product with points       |
| GET    | `/products/search`      | Search for products                |

#### Example: Purchase Credit Package

```http
POST /credits/purchase
Content-Type: application/json

{
  "user_id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
  "package_id": "f2a34b5c-6d7e-8a9b-0c1d-2e4100030000",
  "amount_paid": 50.00
}
```

#### Example: Redeem Product

```http
POST /products/redeem
Content-Type: application/json

{
  "user_id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
  "product_id": "a1b2c3d4-e5f6-a7b8-c9d0-e1f200010000",
  "quantity": 1
}
```

#### Example: Search Products

```http
GET /products/search?query=wireless&page=1&size=10
```

---

### AI Recommendation Endpoint

| Method | Endpoint                | Description                        |
|--------|-------------------------|------------------------------------|
| POST   | `/ai/recommendation`    | Get AI-powered product suggestion  |

#### Example Request

```http
POST /ai/recommendation
Content-Type: application/json

{
  "user_id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
}
```

#### Example Response

```json
{
  "recommended_category_id": "1a2b3c4d-5e6f-7a8b-9c0d-1e2f00010000",
  "min_points_llm": 400,
  "max_points_llm": 800,
  "reasoning": "Based on your point balance and available categories, Electronics is the best fit."
}
```

---

### Health Check

| Method | Endpoint   | Description         |
|--------|------------|---------------------|
| GET    | `/health`  | Service health info |

---

## Sample Test Data

The app seeds the database with sample users, credit packages, categories, and products.  
See [`internal/infrastructure/db/migrations/000002_seed_db.up.sql`](internal/infrastructure/db/migrations/000002_seed_db.up.sql) for details.

**Sample User:**
- ID: `a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11`
- Name: Alice Smith
- Email: user1@example.com
- Point Balance: 1500

**Sample Categories:**
- Electronics (`1a2b3c4d-5e6f-7a8b-9c0d-1e2f00010000`)
- Books (`1a2b3c4d-5e6f-7a8b-9c0d-1e3000020000`)
- Gift Cards (`1a2b3c4d-5e6f-7a8b-9c0d-1e3100030000`)
- Home Goods (`1a2b3c4d-5e6f-7a8b-9c0d-1e3200040000`)

**Sample Products:**
- Wireless Earbuds (Electronics)
- The Great Novel (Books)
- E-Reader (Electronics)
- 10$ Gift Card (Gift Cards)
- Smart Home Hub (Electronics)
- Cookbook: Italian Delights (Books)

---

## AI Prompt Explanation

When a user requests a recommendation, the backend:

1. Fetches all available categories (with IDs and names).
2. Sends the user's point balance and the categories list to the LLM (AI model).
3. The LLM is prompted to select the **category ID** and a point range for the user, returning a JSON like:
   ```json
   {
     "recommended_category_id": "1a2b3c4d-5e6f-7a8b-9c0d-1e2f00010000",
     "min_points_llm": 400,
     "max_points_llm": 800,
     "reasoning": "Electronics is a good fit for your balance."
   }
   ```
4. The backend uses the recommended category ID and point range to fetch and suggest products.

**Prompt Example:**
```
Given a user with a current point balance of 1500 and the following available categories (each with an id and name): [{"id":"1a2b3c4d-...","name":"Electronics"},...], suggest the most suitable category by its id and a corresponding minimum and maximum point range for products within that category. Ensure the response is a JSON object with the following fields: {"recommended_category_id": "string", "min_points_llm": "integer", "max_points_llm": "integer", "reasoning": "string"}.
```

---

## Troubleshooting

- **Ports in use:** Make sure port 8080 is free or change it in `docker-compose.yml`.
- **Database errors:** Check logs for migration or connection issues.
- **Redis errors:** Ensure Redis is running and accessible.
- **AI errors:** If using a real LLM, ensure API keys and endpoints are configured.


---

**Enjoy building with 4sal Reward System!**