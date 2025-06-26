package repositories

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/shawkyelshalawy/4sal-reward/internal/domain/models"
	"github.com/stretchr/testify/assert"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *ProductRepository) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("failed to open sqlmock database: %v", err)
    }
    repo := NewProductRepository(db)
    return db, mock, repo
}

func TestSearch_Success(t *testing.T) {
    db, mock, repo := setupMockDB(t)
    defer db.Close()

    rows := sqlmock.NewRows([]string{"id", "name", "description", "point_cost", "image_url"}).
        AddRow(uuid.New(), "prod1", "desc1", 100, "img1").
        AddRow(uuid.New(), "prod2", "desc2", 200, "img2")

    mock.ExpectQuery("SELECT id, name, description, point_cost, image_url.*FROM products").
        WithArgs("query", 10, 0).
        WillReturnRows(rows)

    products, err := repo.Search(context.Background(), "query", 1, 10)
    assert.NoError(t, err)
    assert.Len(t, products, 2)
}

func TestCreate_Success(t *testing.T) {
    db, mock, repo := setupMockDB(t)
    defer db.Close()

    product := &models.Product{
        ID:            uuid.New(),
        Name:          "prod",
        Description:   "desc",
        CategoryID:    nil,
        PointCost:     100,
        StockQuantity: 10,
        IsActive:      true,
        IsInOfferPool: true,
        ImageURL:      "img",
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
    }

    mock.ExpectExec("INSERT INTO products").
        WithArgs(
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
        ).
        WillReturnResult(sqlmock.NewResult(1, 1))

    err := repo.Create(context.Background(), product)
    assert.NoError(t, err)
}

func TestUpdateOfferStatus_Success(t *testing.T) {
    db, mock, repo := setupMockDB(t)
    defer db.Close()

    id := uuid.New()
    mock.ExpectExec("UPDATE products SET is_in_offer_pool").
        WithArgs(true, id).
        WillReturnResult(sqlmock.NewResult(1, 1))

    err := repo.UpdateOfferStatus(context.Background(), id, true)
    assert.NoError(t, err)
}

func TestRedeemProduct_UserNotFound(t *testing.T) {
    db, mock, repo := setupMockDB(t)
    defer db.Close()

    userID := uuid.New()
    productID := uuid.New()
    quantity := 1

    // Product row
    productRows := sqlmock.NewRows([]string{
        "id", "name", "description", "category_id", "point_cost", "stock_quantity",
        "is_active", "is_in_offer_pool", "image_url", "created_at", "updated_at",
    }).AddRow(
        productID, "prod", "desc", nil, 100, 10, true, true, "img", time.Now(), time.Now(),
    )
    mock.ExpectBegin()
    mock.ExpectQuery("SELECT id, name, description, category_id, point_cost, stock_quantity.*FOR UPDATE").
        WithArgs(productID).
        WillReturnRows(productRows)
    // User row not found
    mock.ExpectQuery("SELECT point_balance FROM users WHERE id = \\$1 FOR UPDATE").
        WithArgs(userID).
        WillReturnError(sql.ErrNoRows)
    mock.ExpectRollback()

    err := repo.RedeemProduct(context.Background(), userID, productID, quantity)
    assert.EqualError(t, err, "user not found")
}

func TestRedeemProduct_InsufficientPoints(t *testing.T) {
    db, mock, repo := setupMockDB(t)
    defer db.Close()

    userID := uuid.New()
    productID := uuid.New()
    quantity := 1

    productRows := sqlmock.NewRows([]string{
        "id", "name", "description", "category_id", "point_cost", "stock_quantity",
        "is_active", "is_in_offer_pool", "image_url", "created_at", "updated_at",
    }).AddRow(
        productID, "prod", "desc", nil, 100, 10, true, true, "img", time.Now(), time.Now(),
    )
    mock.ExpectBegin()
    mock.ExpectQuery("SELECT id, name, description, category_id, point_cost, stock_quantity.*FOR UPDATE").
        WithArgs(productID).
        WillReturnRows(productRows)
    mock.ExpectQuery("SELECT point_balance FROM users WHERE id = \\$1 FOR UPDATE").
        WithArgs(userID).
        WillReturnRows(sqlmock.NewRows([]string{"point_balance"}).AddRow(50))
    mock.ExpectRollback()

    err := repo.RedeemProduct(context.Background(), userID, productID, quantity)
    assert.EqualError(t, err, "insufficient points")
}

func TestRedeemProduct_Success(t *testing.T) {
    db, mock, repo := setupMockDB(t)
    defer db.Close()

    userID := uuid.New()
    productID := uuid.New()
    quantity := 1

    productRows := sqlmock.NewRows([]string{
        "id", "name", "description", "category_id", "point_cost", "stock_quantity",
        "is_active", "is_in_offer_pool", "image_url", "created_at", "updated_at",
    }).AddRow(
        productID, "prod", "desc", nil, 100, 10, true, true, "img", time.Now(), time.Now(),
    )
    mock.ExpectBegin()
    mock.ExpectQuery("SELECT id, name, description, category_id, point_cost, stock_quantity.*FOR UPDATE").
        WithArgs(productID).
        WillReturnRows(productRows)
    mock.ExpectQuery("SELECT point_balance FROM users WHERE id = \\$1 FOR UPDATE").
        WithArgs(userID).
        WillReturnRows(sqlmock.NewRows([]string{"point_balance"}).AddRow(200))
    mock.ExpectExec("UPDATE users SET point_balance = point_balance -").
        WithArgs(100, userID).
        WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectExec("UPDATE products SET stock_quantity = stock_quantity -").
        WithArgs(quantity, productID).
        WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectExec("INSERT INTO point_redemptions").
        WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectCommit()

    err := repo.RedeemProduct(context.Background(), userID, productID, quantity)
    assert.NoError(t, err)
}