package models

import (
	"time"

	"github.com/google/uuid"
)





type Product struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	Name          string     `json:"name" db:"name"`
	Description   string     `json:"description" db:"description"`
	CategoryID    *uuid.UUID `json:"category_id" db:"category_id"` 
	Category      *Category  `json:"category,omitempty"`           
	PointCost     int        `json:"point_cost" db:"point_cost"`
	StockQuantity int        `json:"stock_quantity" db:"stock_quantity"`
	IsActive      bool       `json:"is_active" db:"is_active"`
	IsInOfferPool bool       `json:"is_in_offer_pool" db:"is_in_offer_pool"`
	ImageURL      string     `json:"image_url" db:"image_url"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}
