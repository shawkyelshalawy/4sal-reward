package models

import (
	"time"

	"github.com/google/uuid"
)




type PointRedemption struct {
	ID             uuid.UUID `json:"id" db:"id"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	ProductID      uuid.UUID `json:"product_id" db:"product_id"`
	PointsUsed     int       `json:"points_used" db:"points_used"`
	Quantity       int       `json:"quantity" db:"quantity"`
	RedemptionDate time.Time `json:"redemption_date" db:"redemption_date"`
	Status         string    `json:"status" db:"status"` 
	User           *User     `json:"user,omitempty"`   
	Product        *Product  `json:"product,omitempty"` 
}
