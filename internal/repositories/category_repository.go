package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/shawkyelshalawy/4sal-reward/internal/domain/models"
)

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	query := `SELECT id, name, description, created_at FROM categories WHERE id = $1`
	
	row := r.db.QueryRowContext(ctx, query, id)
	
	var category models.Category
	err := row.Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}
	return &category, nil
}

func (r *CategoryRepository) GetAll(ctx context.Context) ([]interface{}, error) {
	query := `SELECT id, name, description, created_at FROM categories ORDER BY name`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var categories []interface{}
	for rows.Next() {
		var category models.Category
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Description,
			&category.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		// Convert to map for AI prompt
		categoryMap := map[string]interface{}{
			"id":   category.ID.String(),
			"name": category.Name,
		}
		categories = append(categories, categoryMap)
	}
	
	return categories, nil
}

func (r *CategoryRepository) Create(ctx context.Context, category *models.Category) error {
	query := `INSERT INTO categories (id, name, description, created_at) VALUES ($1, $2, $3, $4)`
	
	_, err := r.db.ExecContext(ctx, query,
		category.ID,
		category.Name,
		category.Description,
		category.CreatedAt,
	)
	return err
}