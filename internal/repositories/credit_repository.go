package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

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

func (r *CreditRepository) PurchasePackage(ctx context.Context, userID, packageID uuid.UUID, amountPaid float64) error {
    tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Lock and validate the credit package
    creditPackage, err := r.getPackageForUpdate(ctx, tx, packageID)
    if err != nil {
        return err
    }
    if !creditPackage.IsActive {
        return errors.New("credit package is not active")
    }
    if creditPackage.Price != amountPaid {
        return errors.New("amount paid does not match package price")
    }

    // Lock user row to ensure consistency
    var userExists bool
    err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1) FOR UPDATE", userID).Scan(&userExists)
    if err != nil {
        return err
    }
    if !userExists {
        return errors.New("user not found")
    }

    // Update user's point balance
    _, err = tx.ExecContext(ctx, "UPDATE users SET point_balance = point_balance + $1, updated_at = NOW() WHERE id = $2", 
        creditPackage.RewardPoints, userID)
    if err != nil {
        return err
    }

    // Create the purchase record
    purchase := &models.CreditPurchase{
        ID:               uuid.New(),
        UserID:           userID,
        CreditPackageID:  packageID,
        AmountPaid:       amountPaid,
        PointsAwarded:    creditPackage.RewardPoints,
        PurchaseDate:     time.Now(),
        Status:           "completed",
    }

    _, err = tx.ExecContext(ctx, `INSERT INTO credit_purchases (
        id, user_id, credit_package_id, amount_paid, points_awarded, purchase_date, status
    ) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
        purchase.ID,
        purchase.UserID,
        purchase.CreditPackageID,
        purchase.AmountPaid,
        purchase.PointsAwarded,
        purchase.PurchaseDate,
        purchase.Status,
    )
    if err != nil {
        return err
    }

    return tx.Commit()
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

func (r *CreditRepository) getPackageForUpdate(ctx context.Context, tx *sql.Tx, id uuid.UUID) (*models.CreditPackage, error) {
    query := `SELECT id, name, description, price, reward_points, is_active, created_at, updated_at 
              FROM credit_packages WHERE id = $1 FOR UPDATE`
    
    row := tx.QueryRowContext(ctx, query, id)
    
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