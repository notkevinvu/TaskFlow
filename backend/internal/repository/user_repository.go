package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/sqlc"
)

// UserRepository handles database operations for users using sqlc
type UserRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

// Helper function to convert string UUID to pgtype.UUID
func stringToPgtypeUUID(s string) (pgtype.UUID, error) {
	parsed, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}

	var pguuid pgtype.UUID
	pguuid.Bytes = parsed
	pguuid.Valid = true
	return pguuid, nil
}

// Helper function to convert pgtype.UUID to string
func pgtypeUUIDToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return uuid.UUID(u.Bytes).String()
}

// Helper function to convert time.Time to pgtype.Timestamptz
func timeToPgtypeTimestamptz(t time.Time) pgtype.Timestamptz {
	var ts pgtype.Timestamptz
	ts.Time = t
	ts.Valid = true
	return ts
}

// Helper function to convert pgtype.Timestamptz to time.Time
func pgtypeTimestamptzToTime(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time
}

// Convert sqlc.User to domain.User
func sqlcUserToDomain(u sqlc.User) domain.User {
	return domain.User{
		ID:           pgtypeUUIDToString(u.ID),
		Email:        u.Email,
		Name:         u.Name,
		PasswordHash: u.PasswordHash,
		CreatedAt:    pgtypeTimestamptzToTime(u.CreatedAt),
		UpdatedAt:    pgtypeTimestamptzToTime(u.UpdatedAt),
	}
}

// Create inserts a new user into the database
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	pguuid, err := stringToPgtypeUUID(user.ID)
	if err != nil {
		return err
	}

	params := sqlc.CreateUserParams{
		ID:           pguuid,
		Email:        user.Email,
		Name:         user.Name,
		PasswordHash: user.PasswordHash,
		CreatedAt:    timeToPgtypeTimestamptz(user.CreatedAt),
		UpdatedAt:    timeToPgtypeTimestamptz(user.UpdatedAt),
	}

	return r.queries.CreateUser(ctx, params)
}

// FindByEmail retrieves a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	sqlcUser, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // User not found
		}
		return nil, err
	}

	domainUser := sqlcUserToDomain(sqlcUser)
	return &domainUser, nil
}

// FindByID retrieves a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	pguuid, err := stringToPgtypeUUID(id)
	if err != nil {
		return nil, err
	}

	sqlcUser, err := r.queries.GetUserByID(ctx, pguuid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // User not found
		}
		return nil, err
	}

	domainUser := sqlcUserToDomain(sqlcUser)
	return &domainUser, nil
}

// EmailExists checks if an email already exists
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	return r.queries.CheckEmailExists(ctx, email)
}
