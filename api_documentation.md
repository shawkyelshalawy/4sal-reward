# 4sal Reward System API Documentation

## Base URL
```
http://localhost:8080
```

## Authentication
Currently, the API doesn't require authentication. In production, you would add JWT or API key authentication.

## Endpoints

### Health Check
```http
GET /health
```
Returns the health status of the application and its dependencies.

**Response:**
```json
{
  "status": "UP",
  "db": "OK",
  "redis": "OK"
}
```

---

### Credit Packages

#### List Credit Packages
```http
GET /credits/packages?page=1&size=10
```

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `size` (optional): Items per page (default: 10, max: 100)

**Response:**
```json
{
  "packages": [
    {
      "id": "f2a34b5c-6d7e-8a9b-0c1d-2e3f00010000",
      "name": "Bronze Bundle",
      "description": "Small credit package for beginners",
      "price": 10.00,
      "reward_points": 100,
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "page": 1,
  "size": 10,
  "total": 4,
  "total_pages": 1
}
```

#### Purchase Credit Package
```http
POST /credits/purchase
```

**Request Body:**
```json
{
  "user_id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
  "package_id": "f2a34b5c-6d7e-8a9b-0c1d-2e4100030000",
  "amount_paid": 50.00
}
```

**Response:**
```json
{
  "message": "Credit package purchased successfully"
}
```

---

### Products

#### Search Products
```http
GET /products/search?query=wireless&page=1&size=10
```

**Query Parameters:**
- `query` (required): Search term
- `page` (optional): Page number (default: 1)
- `size` (optional): Items per page (default: 10, max: 100)

**Response:**
```json
{
  "products": [
    {
      "id": "a1b2c3d4-e5f6-a7b8-c9d0-e1f200010000",
      "name": "Wireless Earbuds",
      "description": "High-quality sound with noise cancellation.",
      "point_cost": 500,
      "image_url": "https://placehold.co/300x200/cccccc/333333?text=Earbuds"
    }
  ],
  "page": 1,
  "size": 10
}
```

#### Redeem Product
```http
POST /products/redeem
```

**Request Body:**
```json
{
  "user_id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
  "product_id": "a1b2c3d4-e5f6-a7b8-c9d0-e1f200010000",
  "quantity": 1
}
```

**Response:**
```http
201 Created
```

---

### AI Recommendations (Powered by Google Gemini)

#### Get AI Recommendation with Category Products
```http
POST /ai/recommendation
```

**Request Body:**
```json
{
  "user_id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
}
```

**Response:**
```json
{
  "user_id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
  "point_balance": 1500,
  "recommended_category_id": "1a2b3c4d-5e6f-7a8b-9c0d-1e2f00010000",
  "category_name": "Electronics",
  "min_points_llm": 400,
  "max_points_llm": 800,
  "reasoning": "Based on your point balance of 1500, Electronics category offers the best value with products ranging from 400-800 points, allowing you to make multiple purchases or save for premium items.",
  "recommended_products": [
    {
      "id": "a1b2c3d4-e5f6-a7b8-c9d0-e1f200010000",
      "name": "Wireless Earbuds",
      "description": "High-quality sound with noise cancellation.",
      "point_cost": 500,
      "image_url": "https://placehold.co/300x200/cccccc/333333?text=Earbuds",
      "in_stock": true
    },
    {
      "id": "a1b2c3d4-e5f6-a7b8-c9d0-e1f400030000",
      "name": "E-Reader",
      "description": "Lightweight device for digital reading.",
      "point_cost": 800,
      "image_url": "https://placehold.co/300x200/cccccc/333333?text=E-Reader",
      "in_stock": true
    },
    {
      "id": "a1b2c3d4-e5f6-a7b8-c9d0-e1f600050000",
      "name": "Smart Home Hub",
      "description": "Central control for smart devices.",
      "point_cost": 1200,
      "image_url": "https://placehold.co/300x200/cccccc/333333?text=SmartHub",
      "in_stock": true
    }
  ],
  "total_products": 3
}
```

**Enhanced AI Features:**
- **Complete Category Display**: Shows all available products in the recommended category
- **Stock Information**: Indicates whether each product is currently in stock
- **Sorted by Value**: Products are sorted by point cost (ascending) for easy browsing
- **Category Context**: Includes category name for better user understanding
- **User Context**: Shows user's current point balance for informed decisions

**AI Integration Details:**
- **Model**: Google Gemini 1.5 Flash
- **Fallback**: Rule-based recommendations if AI is unavailable
- **Response Time**: Typically 1-3 seconds
- **Context**: User's point balance and available product categories
- **Product Filtering**: Only shows active products in the offer pool

---

### Admin Endpoints

#### Create Credit Package
```http
POST /admin/packages
```

**Request Body:**
```json
{
  "name": "Platinum Bundle",
  "description": "Premium package with maximum rewards",
  "price": 100.00,
  "reward_points": 1500
}
```

**Response:**
```json
{
  "package_id": "f2a34b5c-6d7e-8a9b-0c1d-2e4200040000"
}
```

#### Update Credit Package
```http
PUT /admin/packages/{package_id}
```

**Request Body:**
```json
{
  "name": "Platinum Bundle Updated",
  "price": 120.00,
  "is_active": true
}
```

**Response:**
```json
{
  "message": "Credit package updated successfully"
}
```

#### Create Product
```http
POST /admin/products
```

**Request Body:**
```json
{
  "name": "Wireless Headphones",
  "description": "Premium noise-cancelling headphones",
  "category_id": "1a2b3c4d-5e6f-7a8b-9c0d-1e2f00010000",
  "point_cost": 800,
  "stock_quantity": 25,
  "is_in_offer_pool": true,
  "image_url": "https://example.com/headphones.jpg"
}
```

**Response:**
```json
{
  "product_id": "a1b2c3d4-e5f6-a7b8-c9d0-e1f800070000"
}
```

#### Update Product
```http
PUT /admin/products/{product_id}
```

**Request Body:**
```json
{
  "name": "Wireless Headphones Pro",
  "stock_quantity": 15,
  "is_active": true
}
```

**Response:**
```json
{
  "message": "Product updated successfully"
}
```

#### Update Product Offer Status
```http
PUT /admin/products/{product_id}/offer-status
```

**Request Body:**
```json
{
  "is_in_offer_pool": false
}
```

**Response:**
```http
200 OK
```

---

## Error Responses

All endpoints return appropriate HTTP status codes and error messages:

**400 Bad Request:**
```json
{
  "error": "Invalid request"
}
```

**404 Not Found:**
```json
{
  "error": "Resource not found"
}
```

**500 Internal Server Error:**
```json
{
  "error": "Internal server error message"
}
```

---

## Sample Test Data

The system comes pre-loaded with sample data:

### Users
- **Alice Smith** (ID: `a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11`) - 1500 points
- **Bob Johnson** (ID: `b1c9de00-0a1b-4c2d-8e3f-9a0b1c2d3e4f`) - 750 points
- **Admin User** (ID: `c2d0ef11-1b2c-5d3e-9f0a-0b1c2d3e4f50`) - 10000 points

### Categories
- **Electronics** (ID: `1a2b3c4d-5e6f-7a8b-9c0d-1e2f00010000`)
- **Books** (ID: `1a2b3c4d-5e6f-7a8b-9c0d-1e3000020000`)
- **Gift Cards** (ID: `1a2b3c4d-5e6f-7a8b-9c0d-1e3100030000`)
- **Home Goods** (ID: `1a2b3c4d-5e6f-7a8b-9c0d-1e3200040000`)

### Credit Packages
- **Bronze Bundle** - $10.00 for 100 points
- **Silver Bundle** - $25.00 for 300 points
- **Gold Bundle** - $50.00 for 750 points
- **Platinum Bundle** - $100.00 for 1500 points

### Products
- **Wireless Earbuds** - 500 points (Electronics)
- **The Great Novel** - 200 points (Books)
- **E-Reader** - 800 points (Electronics)
- **$10 Gift Card** - 1000 points (Gift Cards)
- **Smart Home Hub** - 1200 points (Electronics)
- **Cookbook: Italian Delights** - 300 points (Books)

---

## AI Prompt Engineering

### Gemini Integration Details

The AI recommendation system uses Google Gemini 1.5 Flash with carefully crafted prompts:

**Prompt Structure:**
```
Given a user with a current point balance of {balance} and the following available categories: {categories_json}

Suggest the most suitable category by its id and a corresponding minimum and maximum point range for products within that category.

Consider the user's point balance and recommend a category that offers good value. The point range should be realistic for products in that category.

Ensure the response is a JSON object with the following fields: 
{"recommended_category_id": "string", "min_points_llm": "integer", "max_points_llm": "integer", "reasoning": "string"}

Only return the JSON object, no additional text.
```

**Enhanced Response Processing:**
- JSON parsing with error handling
- Fallback to rule-based logic if AI fails
- Validation of category IDs against database
- Point range validation for reasonableness
- **Category product fetching** for complete user experience
- **Stock status checking** for each product
- **Sorting by point cost** for optimal user browsing

**Performance Considerations:**
- 30-second timeout for AI requests
- Graceful degradation to simple recommendations
- Efficient database queries for category products
- Redis caching for improved performance
- Rate limiting protection (future enhancement)

---

## Complete User Journey

### AI Recommendation Flow
1. **User Request**: User provides their ID for recommendation
2. **Context Gathering**: System fetches user's point balance and available categories
3. **AI Processing**: Gemini AI analyzes and recommends optimal category
4. **Product Fetching**: System retrieves all products in recommended category
5. **Enhanced Response**: User receives recommendation with complete product catalog
6. **Informed Decision**: User can immediately browse and redeem products

This enhanced flow provides a seamless experience from AI recommendation to product selection and redemption.