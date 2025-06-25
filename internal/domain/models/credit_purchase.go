package models

import (
	"time"

	"github.com/google/uuid"
)




type CreditPurchase struct {
	ID            uuid.UUID      `json:"id" db:"id"`
	UserID        uuid.UUID      `json:"user_id" db:"user_id"`
	CreditPackageID uuid.UUID    `json:"credit_package_id" db:"credit_package_id"`
	AmountPaid    float64        `json:"amount_paid" db:"amount_paid"`
	PointsAwarded int            `json:"points_awarded" db:"points_awarded"`
	PurchaseDate  time.Time      `json:"purchase_date" db:"purchase_date"`
	Status        string         `json:"status" db:"status"` 
	User          *User          `json:"user,omitempty"`       
	CreditPackage *CreditPackage `json:"credit_package,omitempty"` 
}
