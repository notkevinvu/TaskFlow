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

// Note: pgtypeTimestamptzToTimePtr and timePtrToPgtypeTimestamptz are defined in task_repository.go

// Create inserts a new user into the database (for registered users - uses legacy sqlc)
// For anonymous users, use CreateAnonymous instead
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	// For registered users, use the original sqlc method
	if user.UserType == domain.UserTypeRegistered {
		pguuid, err := stringToPgtypeUUID(user.ID)
		if err != nil {
			return err
		}

		// Use raw SQL to include user_type column
		query := `
			INSERT INTO users (id, email, name, password_hash, user_type, expires_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err = r.db.Exec(ctx, query,
			pguuid,
			user.Email,
			user.Name,
			user.PasswordHash,
			user.UserType,
			timePtrToPgtypeTimestamptz(user.ExpiresAt),
			timeToPgtypeTimestamptz(user.CreatedAt),
			timeToPgtypeTimestamptz(user.UpdatedAt),
		)
		return err
	}

	// For anonymous users
	return r.CreateAnonymous(ctx, user)
}

// CreateAnonymous creates a new anonymous user
func (r *UserRepository) CreateAnonymous(ctx context.Context, user *domain.User) error {
	pguuid, err := stringToPgtypeUUID(user.ID)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO users (id, email, name, password_hash, user_type, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = r.db.Exec(ctx, query,
		pguuid,
		user.Email,        // nil for anonymous
		user.Name,         // nil for anonymous
		user.PasswordHash, // nil for anonymous
		user.UserType,
		timePtrToPgtypeTimestamptz(user.ExpiresAt),
		timeToPgtypeTimestamptz(user.CreatedAt),
		timeToPgtypeTimestamptz(user.UpdatedAt),
	)
	return err
}

// FindByEmail retrieves a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, name, password_hash, user_type, expires_at, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	return r.scanUser(ctx, query, email)
}

// FindByID retrieves a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	pguuid, err := stringToPgtypeUUID(id)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, email, name, password_hash, user_type, expires_at, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	return r.scanUser(ctx, query, pguuid)
}

// scanUser is a helper to scan a user row from any query
func (r *UserRepository) scanUser(ctx context.Context, query string, args ...interface{}) (*domain.User, error) {
	row := r.db.QueryRow(ctx, query, args...)

	var (
		id           pgtype.UUID
		email        *string
		name         *string
		passwordHash *string
		userType     string
		expiresAt    pgtype.Timestamptz
		createdAt    pgtype.Timestamptz
		updatedAt    pgtype.Timestamptz
	)

	err := row.Scan(&id, &email, &name, &passwordHash, &userType, &expiresAt, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &domain.User{
		ID:           pgtypeUUIDToString(id),
		Email:        email,
		Name:         name,
		PasswordHash: passwordHash,
		UserType:     domain.UserType(userType),
		ExpiresAt:    pgtypeTimestamptzToTimePtr(expiresAt),
		CreatedAt:    pgtypeTimestamptzToTime(createdAt),
		UpdatedAt:    pgtypeTimestamptzToTime(updatedAt),
	}, nil
}

// EmailExists checks if an email already exists
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	return r.queries.CheckEmailExists(ctx, email)
}

// ConvertToRegistered converts an anonymous user to a registered user
func (r *UserRepository) ConvertToRegistered(ctx context.Context, userID string, email, name, passwordHash string) error {
	pguuid, err := stringToPgtypeUUID(userID)
	if err != nil {
		return err
	}

	query := `
		UPDATE users
		SET email = $2,
			name = $3,
			password_hash = $4,
			user_type = 'registered',
			expires_at = NULL,
			updated_at = $5
		WHERE id = $1 AND user_type = 'anonymous'
	`
	result, err := r.db.Exec(ctx, query, pguuid, email, name, passwordHash, time.Now())
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("user not found or not anonymous")
	}

	return nil
}

// FindExpiredAnonymous returns all anonymous users whose expires_at is in the past
func (r *UserRepository) FindExpiredAnonymous(ctx context.Context) ([]*domain.User, error) {
	query := `
		SELECT id, email, name, password_hash, user_type, expires_at, created_at, updated_at
		FROM users
		WHERE user_type = 'anonymous'
		  AND expires_at IS NOT NULL
		  AND expires_at < NOW()
		ORDER BY expires_at ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var (
			id           pgtype.UUID
			email        *string
			name         *string
			passwordHash *string
			userType     string
			expiresAt    pgtype.Timestamptz
			createdAt    pgtype.Timestamptz
			updatedAt    pgtype.Timestamptz
		)

		err := rows.Scan(&id, &email, &name, &passwordHash, &userType, &expiresAt, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}

		users = append(users, &domain.User{
			ID:           pgtypeUUIDToString(id),
			Email:        email,
			Name:         name,
			PasswordHash: passwordHash,
			UserType:     domain.UserType(userType),
			ExpiresAt:    pgtypeTimestamptzToTimePtr(expiresAt),
			CreatedAt:    pgtypeTimestamptzToTime(createdAt),
			UpdatedAt:    pgtypeTimestamptzToTime(updatedAt),
		})
	}

	return users, rows.Err()
}

// Delete removes a user by ID (cascade deletes tasks via FK)
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	pguuid, err := stringToPgtypeUUID(id)
	if err != nil {
		return err
	}

	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.Exec(ctx, query, pguuid)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("user not found")
	}

	return nil
}

// CountTasksByUserID returns the number of tasks owned by a user
func (r *UserRepository) CountTasksByUserID(ctx context.Context, userID string) (int, error) {
	pguuid, err := stringToPgtypeUUID(userID)
	if err != nil {
		return 0, err
	}

	query := `SELECT COUNT(*) FROM tasks WHERE user_id = $1`
	var count int
	err = r.db.QueryRow(ctx, query, pguuid).Scan(&count)
	return count, err
}

// LogAnonymousCleanup records the cleanup of an anonymous user for audit purposes
func (r *UserRepository) LogAnonymousCleanup(ctx context.Context, userID string, taskCount int, userCreatedAt time.Time) error {
	pguuid, err := stringToPgtypeUUID(userID)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO anonymous_user_cleanups (user_id, task_count, created_at, deleted_at)
		VALUES ($1, $2, $3, NOW())
	`
	_, err = r.db.Exec(ctx, query, pguuid, taskCount, timeToPgtypeTimestamptz(userCreatedAt))
	return err
}
