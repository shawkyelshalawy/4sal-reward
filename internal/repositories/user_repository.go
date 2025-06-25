package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/shawkyelshalawy/4sal-reward/internal/domain/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `SELECT id, email, name, point_balance, created_at, updated_at 
              FROM users WHERE id = $1`
	
	row := r.db.QueryRowContext(ctx, query, id)
	
	var user models.User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PointBalance,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) AddPoints(ctx context.Context, userID uuid.UUID, points int) error {
	query := `UPDATE users SET 
                point_balance = point_balance + $1, 
                updated_at = NOW() 
              WHERE id = $2`
	
	_, err := r.db.ExecContext(ctx, query, points, userID)
	return err
}

func (r *UserRepository) DeductPoints(ctx context.Context, userID uuid.UUID, points int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	var currentBalance int
	err = tx.QueryRowContext(ctx, "SELECT point_balance FROM users WHERE id = $1 FOR UPDATE", userID).Scan(&currentBalance)
	if err != nil {
		return err
	}
	
	if currentBalance < points {
		return errors.New("insufficient points")
	}
	
	_, err = tx.ExecContext(ctx, 
		"UPDATE users SET point_balance = point_balance - $1, updated_at = NOW() WHERE id = $2", 
		points, userID)
	if err != nil {
		return err
	}
	
	return tx.Commit()
}