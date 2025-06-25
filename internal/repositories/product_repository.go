package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shawkyelshalawy/4sal-reward/internal/domain/models"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) GetForUpdate(ctx context.Context, tx *sql.Tx, id uuid.UUID) (*models.Product, error) {
	query := `SELECT id, name, description, category_id, point_cost, stock_quantity, 
                     is_active, is_in_offer_pool, image_url, created_at, updated_at 
              FROM products WHERE id = $1 FOR UPDATE`
	row := tx.QueryRowContext(ctx, query, id)

	var product models.Product
	err := row.Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.CategoryID,
		&product.PointCost,
		&product.StockQuantity,
		&product.IsActive,
		&product.IsInOfferPool,
		&product.ImageURL,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("product not found")
		}
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) Search(ctx context.Context, query string, page, size int) ([]models.Product, error) {
	offset := (page - 1) * size
	sqlQuery := `
        SELECT id, name, description, point_cost, image_url
        FROM products 
        WHERE to_tsvector('english', name || ' ' || description) @@ to_tsquery($1)
        AND is_in_offer_pool = TRUE
        LIMIT $2 OFFSET $3
    `

	rows, err := r.db.QueryContext(ctx, sqlQuery, query, size, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.PointCost,
			&p.ImageURL,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

func (r *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	query := `INSERT INTO products (
                id, name, description, category_id, point_cost, stock_quantity, 
                is_active, is_in_offer_pool, image_url, created_at, updated_at
              ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.ExecContext(ctx, query,
		product.ID,
		product.Name,
		product.Description,
		product.CategoryID,
		product.PointCost,
		product.StockQuantity,
		product.IsActive,
		product.IsInOfferPool,
		product.ImageURL,
		product.CreatedAt,
		product.UpdatedAt,
	)
	return err
}

func (r *ProductRepository) UpdateOfferStatus(ctx context.Context, productID uuid.UUID, isInOfferPool bool) error {
	query := `UPDATE products SET 
                is_in_offer_pool = $1,
                updated_at = NOW()
              WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, isInOfferPool, productID)
	return err
}

func (r *ProductRepository) RedeemProduct(ctx context.Context, userID, productID uuid.UUID, quantity int) error {
    tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Lock product row
    product, err := r.GetForUpdate(ctx, tx, productID)
    if err != nil {
        return err
    }
    if !product.IsActive {
        return errors.New("product is not active")
    }
    if !product.IsInOfferPool {
        return errors.New("product not available for redemption")
    }
    if product.StockQuantity < quantity {
        return errors.New("insufficient stock")
    }
    pointsNeeded := product.PointCost * quantity

    // Lock user row and check points
    var currentBalance int
    err = tx.QueryRowContext(ctx, "SELECT point_balance FROM users WHERE id = $1 FOR UPDATE", userID).Scan(&currentBalance)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return errors.New("user not found")
        }
        return err
    }
    if currentBalance < pointsNeeded {
        return errors.New("insufficient points")
    }

    _, err = tx.ExecContext(ctx, "UPDATE users SET point_balance = point_balance - $1, updated_at = NOW() WHERE id = $2", pointsNeeded, userID)
    if err != nil {
        return err
    }

    _, err = tx.ExecContext(ctx, "UPDATE products SET stock_quantity = stock_quantity - $1, updated_at = NOW() WHERE id = $2", quantity, productID)
    if err != nil {
        return err
    }

    redemption := &models.PointRedemption{
        ID:             uuid.New(),
        UserID:         userID,
        ProductID:      productID,
        PointsUsed:     pointsNeeded,
        Quantity:       quantity,
        RedemptionDate: time.Now(),
        Status:         "completed",
    }
    _, err = tx.ExecContext(ctx, `INSERT INTO point_redemptions (
        id, user_id, product_id, points_used, quantity, redemption_date, status
    ) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
        redemption.ID,
        redemption.UserID,
        redemption.ProductID,
        redemption.PointsUsed,
        redemption.Quantity,
        redemption.RedemptionDate,
        redemption.Status,
    )
    if err != nil {
        return err
    }

    return tx.Commit()
}