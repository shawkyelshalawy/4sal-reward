package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
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

func (r *ProductRepository) UpdateProduct(ctx context.Context, productID uuid.UUID, name, description *string, categoryID *uuid.UUID, pointCost, stockQuantity *int, isActive, isInOfferPool *bool, imageURL *string) error {
	var setParts []string
	var args []interface{}
	argIndex := 1

	if name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *name)
		argIndex++
	}
	
	if description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *description)
		argIndex++
	}
	
	if categoryID != nil {
		setParts = append(setParts, fmt.Sprintf("category_id = $%d", argIndex))
		args = append(args, *categoryID)
		argIndex++
	}
	
	if pointCost != nil {
		setParts = append(setParts, fmt.Sprintf("point_cost = $%d", argIndex))
		args = append(args, *pointCost)
		argIndex++
	}
	
	if stockQuantity != nil {
		setParts = append(setParts, fmt.Sprintf("stock_quantity = $%d", argIndex))
		args = append(args, *stockQuantity)
		argIndex++
	}
	
	if isActive != nil {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *isActive)
		argIndex++
	}
	
	if isInOfferPool != nil {
		setParts = append(setParts, fmt.Sprintf("is_in_offer_pool = $%d", argIndex))
		args = append(args, *isInOfferPool)
		argIndex++
	}
	
	if imageURL != nil {
		setParts = append(setParts, fmt.Sprintf("image_url = $%d", argIndex))
		args = append(args, *imageURL)
		argIndex++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	// Add updated_at field
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add product ID for WHERE clause
	args = append(args, productID)

	query := fmt.Sprintf("UPDATE products SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)
	
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// GetProducts retrieves products with pagination and filters
func (r *ProductRepository) GetProducts(ctx context.Context, page, size int, isActive, isInOfferPool, categoryID string) ([]models.Product, int, error) {
	// Build WHERE clause dynamically
	var whereClauses []string
	var args []interface{}
	argIndex := 1

	if isActive != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, isActive == "true")
		argIndex++
	}

	if isInOfferPool != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("is_in_offer_pool = $%d", argIndex))
		args = append(args, isInOfferPool == "true")
		argIndex++
	}

	if categoryID != "" {
		if categoryID == "null" {
			whereClauses = append(whereClauses, "category_id IS NULL")
		} else {
			whereClauses = append(whereClauses, fmt.Sprintf("category_id = $%d", argIndex))
			if parsedUUID, err := uuid.Parse(categoryID); err == nil {
				args = append(args, parsedUUID)
				argIndex++
			}
		}
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Add pagination parameters
	args = append(args, size, (page-1)*size)
	limitOffset := fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)

	// Get products with pagination
	query := fmt.Sprintf(`SELECT id, name, description, category_id, point_cost, stock_quantity, 
                                 is_active, is_in_offer_pool, image_url, created_at, updated_at 
                          FROM products %s%s`, whereClause, limitOffset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(
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
			return nil, 0, err
		}
		products = append(products, product)
	}

	return products, total, nil
}

// GetProductsByCategory retrieves products by category ID with pagination
func (r *ProductRepository) GetProductsByCategory(ctx context.Context, categoryID uuid.UUID, page, size int) ([]models.Product, int, error) {
	// Get total count for the category
	countQuery := `SELECT COUNT(*) FROM products WHERE category_id = $1 AND is_active = true AND is_in_offer_pool = true`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, categoryID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get products with pagination
	offset := (page - 1) * size
	query := `SELECT id, name, description, category_id, point_cost, stock_quantity, 
                     is_active, is_in_offer_pool, image_url, created_at, updated_at 
              FROM products 
              WHERE category_id = $1 AND is_active = true AND is_in_offer_pool = true
              ORDER BY point_cost ASC, name ASC
              LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, categoryID, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(
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
			return nil, 0, err
		}
		products = append(products, product)
	}

	return products, total, nil
}