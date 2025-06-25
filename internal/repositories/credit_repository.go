package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/shawkyelshalawy/4sal-reward/internal/domain/models"
)

type CreditRepository struct {
	db *sql.DB
}

func NewCreditRepository(db *sql.DB) *CreditRepository {
	return &CreditRepository{db: db}
}

func (r *CreditRepository) GetPackage(ctx context.Context, id uuid.UUID) (*models.CreditPackage, error) {
	query := `SELECT id, name, description, price, reward_points, is_active, created_at, updated_at 
              FROM credit_packages WHERE id = $1`
	
	row := r.db.QueryRowContext(ctx, query, id)
	
	var pkg models.CreditPackage
	err := row.Scan(
		&pkg.ID,
		&pkg.Name,
		&pkg.Description,
		&pkg.Price,
		&pkg.RewardPoints,
		&pkg.IsActive,
		&pkg.CreatedAt,
		&pkg.UpdatedAt,
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("credit package not found")
		}
		return nil, err
	}
	return &pkg, nil
}

func (r *CreditRepository) CreatePurchase(ctx context.Context, purchase *models.CreditPurchase) error {
	query := `INSERT INTO credit_purchases (
                id, user_id, credit_package_id, amount_paid, points_awarded, purchase_date, status
              ) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	
	_, err := r.db.ExecContext(ctx, query,
		purchase.ID,
		purchase.UserID,
		purchase.CreditPackageID,
		purchase.AmountPaid,
		purchase.PointsAwarded,
		purchase.PurchaseDate,
		purchase.Status,
	)
	return err
}

func (r *CreditRepository) CreatePackage(ctx context.Context, pkg *models.CreditPackage) error {
	query := `INSERT INTO credit_packages (
                id, name, description, price, reward_points, is_active, created_at, updated_at
              ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	
	_, err := r.db.ExecContext(ctx, query,
		pkg.ID,
		pkg.Name,
		pkg.Description,
		pkg.Price,
		pkg.RewardPoints,
		pkg.IsActive,
		pkg.CreatedAt,
		pkg.UpdatedAt,
	)
	return err
}
