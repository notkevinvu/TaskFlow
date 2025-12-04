package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/sqlc"
)

// UserPreferencesRepository handles database operations for user preferences
type UserPreferencesRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

// NewUserPreferencesRepository creates a new user preferences repository
func NewUserPreferencesRepository(db *pgxpool.Pool) *UserPreferencesRepository {
	return &UserPreferencesRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

// Helper: Convert sqlc.UserPreference to domain.UserPreferences
func sqlcUserPreferencesToDomain(p sqlc.UserPreference) domain.UserPreferences {
	return domain.UserPreferences{
		UserID:                    pgtypeUUIDToString(p.UserID),
		DefaultDueDateCalculation: sqlcCalculationToDomain(p.DefaultDueDateCalculation),
		CreatedAt:                 pgtypeTimestamptzToTime(p.CreatedAt),
		UpdatedAt:                 pgtypeTimestamptzToTime(p.UpdatedAt),
	}
}

// Helper: Convert sqlc.CategoryPreference to domain.CategoryPreference
func sqlcCategoryPreferenceToDomain(p sqlc.CategoryPreference) domain.CategoryPreference {
	return domain.CategoryPreference{
		ID:                 pgtypeUUIDToString(p.ID),
		UserID:             pgtypeUUIDToString(p.UserID),
		Category:           p.Category,
		DueDateCalculation: sqlcCalculationToDomain(p.DueDateCalculation),
		CreatedAt:          pgtypeTimestamptzToTime(p.CreatedAt),
		UpdatedAt:          pgtypeTimestamptzToTime(p.UpdatedAt),
	}
}

// GetUserPreferences retrieves user preferences by user ID
func (r *UserPreferencesRepository) GetUserPreferences(ctx context.Context, userID string) (*domain.UserPreferences, error) {
	pguuid, err := stringToPgtypeUUID(userID)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.GetUserPreferences(ctx, pguuid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPreferencesNotFound
		}
		return nil, err
	}

	prefs := sqlcUserPreferencesToDomain(row)
	return &prefs, nil
}

// UpsertUserPreferences creates or updates user preferences
func (r *UserPreferencesRepository) UpsertUserPreferences(ctx context.Context, prefs *domain.UserPreferences) error {
	userID, err := stringToPgtypeUUID(prefs.UserID)
	if err != nil {
		return err
	}

	return r.queries.UpsertUserPreferences(ctx, sqlc.UpsertUserPreferencesParams{
		UserID:                    userID,
		DefaultDueDateCalculation: domainCalculationToSqlc(prefs.DefaultDueDateCalculation),
	})
}

// DeleteUserPreferences removes user preferences
func (r *UserPreferencesRepository) DeleteUserPreferences(ctx context.Context, userID string) error {
	pguuid, err := stringToPgtypeUUID(userID)
	if err != nil {
		return err
	}

	return r.queries.DeleteUserPreferences(ctx, pguuid)
}

// GetCategoryPreference retrieves a specific category preference
func (r *UserPreferencesRepository) GetCategoryPreference(ctx context.Context, userID, category string) (*domain.CategoryPreference, error) {
	userUUID, err := stringToPgtypeUUID(userID)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.GetCategoryPreference(ctx, sqlc.GetCategoryPreferenceParams{
		UserID:   userUUID,
		Category: category,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCategoryPrefNotFound
		}
		return nil, err
	}

	pref := sqlcCategoryPreferenceToDomain(row)
	return &pref, nil
}

// GetCategoryPreferencesByUserID retrieves all category preferences for a user
func (r *UserPreferencesRepository) GetCategoryPreferencesByUserID(ctx context.Context, userID string) ([]domain.CategoryPreference, error) {
	pguuid, err := stringToPgtypeUUID(userID)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.GetCategoryPreferencesByUserID(ctx, pguuid)
	if err != nil {
		return nil, err
	}

	prefs := make([]domain.CategoryPreference, len(rows))
	for i, row := range rows {
		prefs[i] = sqlcCategoryPreferenceToDomain(row)
	}

	return prefs, nil
}

// UpsertCategoryPreference creates or updates a category preference
func (r *UserPreferencesRepository) UpsertCategoryPreference(ctx context.Context, pref *domain.CategoryPreference) error {
	var id string
	if pref.ID == "" {
		id = uuid.New().String()
	} else {
		id = pref.ID
	}

	idUUID, err := stringToPgtypeUUID(id)
	if err != nil {
		return err
	}
	userID, err := stringToPgtypeUUID(pref.UserID)
	if err != nil {
		return err
	}

	return r.queries.UpsertCategoryPreference(ctx, sqlc.UpsertCategoryPreferenceParams{
		ID:                 idUUID,
		UserID:             userID,
		Category:           pref.Category,
		DueDateCalculation: domainCalculationToSqlc(pref.DueDateCalculation),
	})
}

// DeleteCategoryPreference removes a category preference
func (r *UserPreferencesRepository) DeleteCategoryPreference(ctx context.Context, userID, category string) error {
	userUUID, err := stringToPgtypeUUID(userID)
	if err != nil {
		return err
	}

	return r.queries.DeleteCategoryPreference(ctx, sqlc.DeleteCategoryPreferenceParams{
		UserID:   userUUID,
		Category: category,
	})
}

// DeleteAllCategoryPreferences removes all category preferences for a user
func (r *UserPreferencesRepository) DeleteAllCategoryPreferences(ctx context.Context, userID string) error {
	pguuid, err := stringToPgtypeUUID(userID)
	if err != nil {
		return err
	}

	return r.queries.DeleteAllCategoryPreferences(ctx, pguuid)
}

// GetAllPreferences retrieves both user and category preferences
func (r *UserPreferencesRepository) GetAllPreferences(ctx context.Context, userID string) (*domain.AllPreferences, error) {
	userPrefs, err := r.GetUserPreferences(ctx, userID)
	if err != nil && !errors.Is(err, domain.ErrPreferencesNotFound) {
		return nil, err
	}

	categoryPrefs, err := r.GetCategoryPreferencesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &domain.AllPreferences{
		UserPreferences:     userPrefs,
		CategoryPreferences: categoryPrefs,
	}, nil
}
