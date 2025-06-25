package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/shawkyelshalawy/4sal-reward/internal/domain/models"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	query := `SELECT id, name, description, category_id, point_cost, stock_quantity, 
                     is_active, is_in_offer_pool, image_url, created_at, updated_at 
              FROM products WHERE id = $1`
	
	row := r.db.QueryRowContext(ctx, query, id)
	
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

func (r *ProductRepository) ReduceStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	query := `UPDATE products SET 
                stock_quantity = stock_quantity - $1,
                updated_at = NOW()
              WHERE id = $2 AND stock_quantity >= $1`
	
	result, err := r.db.ExecContext(ctx, query, quantity, productID)
	if err != nil {
		return err
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("insufficient stock")
	}
	
	return nil
}

func (r *ProductRepository) CreateRedemption(ctx context.Context, redemption *models.PointRedemption) error {
	query := `INSERT INTO point_redemptions (
                id, user_id, product_id, points_used, quantity, redemption_date, status
              ) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	
	_, err := r.db.ExecContext(ctx, query,
		redemption.ID,
		redemption.UserID,
		redemption.ProductID,
		redemption.PointsUsed,
		redemption.Quantity,
		redemption.RedemptionDate,
		redemption.Status,
	)
	return err
}