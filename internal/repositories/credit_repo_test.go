package repositories

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/shawkyelshalawy/4sal-reward/internal/domain/models"
	"github.com/stretchr/testify/assert"
)

func setupCreditMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *CreditRepository) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("failed to open sqlmock database: %v", err)
    }
    repo := NewCreditRepository(db)
    return db, mock, repo
}

func TestGetPackage_Success(t *testing.T) {
    db, mock, repo := setupCreditMockDB(t)
    defer db.Close()

    id := uuid.New()
    now := time.Now()
    rows := sqlmock.NewRows([]string{
        "id", "name", "description", "price", "reward_points", "is_active", "created_at", "updated_at",
    }).AddRow(
        id, "pkg", "desc", 99.99, 1000, true, now, now,
    )

    mock.ExpectQuery(regexp.QuoteMeta(
        `SELECT id, name, description, price, reward_points, is_active, created_at, updated_at 
              FROM credit_packages WHERE id = $1`)).
        WithArgs(id).
        WillReturnRows(rows)

    pkg, err := repo.GetPackage(context.Background(), id)
    assert.NoError(t, err)
    assert.Equal(t, id, pkg.ID)
    assert.Equal(t, "pkg", pkg.Name)
}

func TestGetPackage_NotFound(t *testing.T) {
    db, mock, repo := setupCreditMockDB(t)
    defer db.Close()

    id := uuid.New()
    mock.ExpectQuery("SELECT id, name, description, price, reward_points, is_active, created_at, updated_at.*FROM credit_packages WHERE id = \\$1").
        WithArgs(id).
        WillReturnError(sql.ErrNoRows)

    pkg, err := repo.GetPackage(context.Background(), id)
    assert.Nil(t, pkg)
    assert.EqualError(t, err, "credit package not found")
}

func TestCreatePackage_Success(t *testing.T) {
    db, mock, repo := setupCreditMockDB(t)
    defer db.Close()

    now := time.Now()
    pkg := &models.CreditPackage{
        ID:           uuid.New(),
        Name:         "pkg",
        Description:  "desc",
        Price:        99.99,
        RewardPoints: 1000,
        IsActive:     true,
        CreatedAt:    now,
        UpdatedAt:    now,
    }

    mock.ExpectExec("INSERT INTO credit_packages").
        WithArgs(
            pkg.ID,
            pkg.Name,
            pkg.Description,
            pkg.Price,
            pkg.RewardPoints,
            pkg.IsActive,
            pkg.CreatedAt,
            pkg.UpdatedAt,
        ).
        WillReturnResult(sqlmock.NewResult(1, 1))

    err := repo.CreatePackage(context.Background(), pkg)
    assert.NoError(t, err)
}

func TestPurchasePackage_Success(t *testing.T) {
    db, mock, repo := setupCreditMockDB(t)
    defer db.Close()

    userID := uuid.New()
    packageID := uuid.New()
    amountPaid := 99.99
    now := time.Now()

    // Begin transaction
    mock.ExpectBegin()

    // Lock and validate the credit package
    rows := sqlmock.NewRows([]string{
        "id", "name", "description", "price", "reward_points", "is_active", "created_at", "updated_at",
    }).AddRow(
        packageID, "pkg", "desc", amountPaid, 1000, true, now, now,
    )
    mock.ExpectQuery("SELECT id, name, description, price, reward_points, is_active, created_at, updated_at.*FOR UPDATE").
        WithArgs(packageID).
        WillReturnRows(rows)

    // Lock user row
    mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM users WHERE id = \\$1\\) FOR UPDATE").
        WithArgs(userID).
        WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

    // Update user's point balance
    mock.ExpectExec("UPDATE users SET point_balance = point_balance \\+").
        WithArgs(1000, userID).
        WillReturnResult(sqlmock.NewResult(1, 1))

    // Insert purchase record
    mock.ExpectExec("INSERT INTO credit_purchases").
        WillReturnResult(sqlmock.NewResult(1, 1))

    // Commit
    mock.ExpectCommit()

    err := repo.PurchasePackage(context.Background(), userID, packageID, amountPaid)
    assert.NoError(t, err)
}

func TestPurchasePackage_PackageNotActive(t *testing.T) {
    db, mock, repo := setupCreditMockDB(t)
    defer db.Close()

    userID := uuid.New()
    packageID := uuid.New()
    amountPaid := 99.99
    now := time.Now()

    mock.ExpectBegin()
    rows := sqlmock.NewRows([]string{
        "id", "name", "description", "price", "reward_points", "is_active", "created_at", "updated_at",
    }).AddRow(
        packageID, "pkg", "desc", amountPaid, 1000, false, now, now,
    )
    mock.ExpectQuery("SELECT id, name, description, price, reward_points, is_active, created_at, updated_at.*FOR UPDATE").
        WithArgs(packageID).
        WillReturnRows(rows)
    mock.ExpectRollback()

    err := repo.PurchasePackage(context.Background(), userID, packageID, amountPaid)
    assert.EqualError(t, err, "credit package is not active")
}

func TestPurchasePackage_AmountMismatch(t *testing.T) {
    db, mock, repo := setupCreditMockDB(t)
    defer db.Close()

    userID := uuid.New()
    packageID := uuid.New()
    now := time.Now()

    mock.ExpectBegin()
    rows := sqlmock.NewRows([]string{
        "id", "name", "description", "price", "reward_points", "is_active", "created_at", "updated_at",
    }).AddRow(
        packageID, "pkg", "desc", 50.00, 1000, true, now, now,
    )
    mock.ExpectQuery("SELECT id, name, description, price, reward_points, is_active, created_at, updated_at.*FOR UPDATE").
        WithArgs(packageID).
        WillReturnRows(rows)
    mock.ExpectRollback()

    err := repo.PurchasePackage(context.Background(), userID, packageID, 99.99)
    assert.EqualError(t, err, "amount paid does not match package price")
}

func TestPurchasePackage_UserNotFound(t *testing.T) {
    db, mock, repo := setupCreditMockDB(t)
    defer db.Close()

    userID := uuid.New()
    packageID := uuid.New()
    amountPaid := 99.99
    now := time.Now()

    mock.ExpectBegin()
    rows := sqlmock.NewRows([]string{
        "id", "name", "description", "price", "reward_points", "is_active", "created_at", "updated_at",
    }).AddRow(
        packageID, "pkg", "desc", amountPaid, 1000, true, now, now,
    )
    mock.ExpectQuery("SELECT id, name, description, price, reward_points, is_active, created_at, updated_at.*FOR UPDATE").
        WithArgs(packageID).
        WillReturnRows(rows)
    mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM users WHERE id = \\$1\\) FOR UPDATE").
        WithArgs(userID).
        WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
    mock.ExpectRollback()

    err := repo.PurchasePackage(context.Background(), userID, packageID, amountPaid)
    assert.EqualError(t, err, "user not found")
}