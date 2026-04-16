//go:build integration

// Profile integration tests for the auth service.
// Requires Docker (testcontainers-go spins up PostgreSQL automatically).
// Run: go test -tags=integration -v ./internal/service/ -timeout 120s

package service

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestUpdateProfile_Name_UpdatesDB verifies that UpdateProfile with a valid name
// persists the change and returns updated UserInfo from the database.
func TestUpdateProfile_Name_UpdatesDB(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userID := seedTestUser(t, pool, "profile-name@test.com", "password123", "role-user")
	auth := NewAuth(pool, testJWTCfg(t), "t-profile", "profile-tenant")

	info, err := auth.UpdateProfile(context.Background(), userID, UpdateProfileRequest{
		Name: "Updated Name",
	})
	require.NoError(t, err)
	require.Equal(t, "Updated Name", info.Name)
	require.Equal(t, "profile-name@test.com", info.Email)

	// Verify the change is persisted in DB (not just in the returned struct)
	var dbName string
	err = pool.QueryRow(context.Background(),
		`SELECT name FROM users WHERE id = $1`, userID,
	).Scan(&dbName)
	require.NoError(t, err)
	require.Equal(t, "Updated Name", dbName)
}

// TestUpdateProfile_EmptyName_ReturnsValidationError verifies that an empty name
// is rejected with ErrValidation.
func TestUpdateProfile_EmptyName_ReturnsValidationError(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userID := seedTestUser(t, pool, "profile-empty@test.com", "password123", "role-user")
	auth := NewAuth(pool, testJWTCfg(t), "t-profile", "profile-tenant")

	_, err := auth.UpdateProfile(context.Background(), userID, UpdateProfileRequest{
		Name: "",
	})
	require.ErrorIs(t, err, ErrValidation, "empty name must return ErrValidation")
}

// TestUpdateProfile_NameTooLong_ReturnsValidationError verifies that a name
// exceeding 200 characters is rejected. This acts as the length-validation test
// (analogous to URL validation — UpdateProfileRequest only has Name, not AvatarURL).
// TDD-ANCHOR: UpdateProfileRequest has only Name field — AvatarURL is not implemented
// in the service layer. If added in future, add TestUpdateProfile_AvatarURL_Validates.
func TestUpdateProfile_NameTooLong_ReturnsValidationError(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userID := seedTestUser(t, pool, "profile-long@test.com", "password123", "role-user")
	auth := NewAuth(pool, testJWTCfg(t), "t-profile", "profile-tenant")

	longName := strings.Repeat("a", 201)
	_, err := auth.UpdateProfile(context.Background(), userID, UpdateProfileRequest{
		Name: longName,
	})
	require.ErrorIs(t, err, ErrValidation, "name over 200 chars must return ErrValidation")
}

// TestUpdateProfile_NonexistentUser_ReturnsNotFound verifies that updating a
// user that does not exist returns ErrUserNotFound.
func TestUpdateProfile_NonexistentUser_ReturnsNotFound(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	auth := NewAuth(pool, testJWTCfg(t), "t-profile", "profile-tenant")

	_, err := auth.UpdateProfile(context.Background(), "nonexistent-user-id", UpdateProfileRequest{
		Name: "Valid Name",
	})
	require.ErrorIs(t, err, ErrUserNotFound)
}

// TestListUsers_ReturnsPaginated verifies that ListUsers respects limit and offset.
// Seeds 3 users, fetches first 2, then the remaining 1 via offset.
func TestListUsers_ReturnsPaginated(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	auth := NewAuth(pool, testJWTCfg(t), "t-profile", "profile-tenant")

	// Seed 3 users with unique emails
	seedTestUser(t, pool, "list-user1@test.com", "pass", "role-user")
	seedTestUser(t, pool, "list-user2@test.com", "pass", "role-user")
	seedTestUser(t, pool, "list-user3@test.com", "pass", "role-admin")

	// First page: limit 2, offset 0
	page1, err := auth.ListUsers(context.Background(), 2, 0)
	require.NoError(t, err)
	require.Len(t, page1, 2, "first page must have exactly 2 users")

	// Second page: limit 2, offset 2
	page2, err := auth.ListUsers(context.Background(), 2, 2)
	require.NoError(t, err)
	require.Len(t, page2, 1, "second page must have exactly 1 user")

	// No user should appear in both pages (no duplicates)
	seenIDs := make(map[string]bool)
	for _, u := range page1 {
		seenIDs[u.ID] = true
	}
	for _, u := range page2 {
		require.False(t, seenIDs[u.ID],
			"user %s appears in both pages — pagination overlap", u.ID)
	}

	// Each returned item must have required fields
	for _, u := range append(page1, page2...) {
		require.NotEmpty(t, u.ID)
		require.NotEmpty(t, u.Email)
		require.NotEmpty(t, u.Name)
		require.NotEmpty(t, u.Role)
	}
}

// TestListUsers_EmptyTable_ReturnsEmptySlice verifies that ListUsers on an empty
// table returns an empty (non-nil) slice without error.
func TestListUsers_EmptyTable_ReturnsEmptySlice(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	auth := NewAuth(pool, testJWTCfg(t), "t-profile", "profile-tenant")

	users, err := auth.ListUsers(context.Background(), 10, 0)
	require.NoError(t, err)
	require.NotNil(t, users, "result must be non-nil even when empty")
	require.Empty(t, users, "expected empty slice for empty table")
}
