package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pquerna/otp/totp"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/services/auth/internal/repository"
)

// MFASetupResult holds the data needed for a user to configure their authenticator app.
type MFASetupResult struct {
	Secret string `json:"secret"` // base32-encoded TOTP secret
	URI    string `json:"uri"`    // otpauth:// URI for QR code generation
}

// SetupMFA generates a TOTP secret for the user and stores it (inactive until verified).
func (a *Auth) SetupMFA(ctx context.Context, userID string) (*MFASetupResult, error) {
	// Get user email for the TOTP issuer label
	email, err := a.repo.GetUserEmail(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user email: %w", err)
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "SDA Framework",
		AccountName: email,
	})
	if err != nil {
		return nil, fmt.Errorf("generate TOTP key: %w", err)
	}

	// Store secret (not yet activated — mfa_enabled stays false until VerifySetup)
	err = a.repo.SetMFASecret(ctx, repository.SetMFASecretParams{
		ID:        userID,
		MfaSecret: pgtype.Text{String: key.Secret(), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("store MFA secret: %w", err)
	}

	return &MFASetupResult{
		Secret: key.Secret(),
		URI:    key.URL(),
	}, nil
}

// VerifySetup activates MFA after the user proves they can generate valid codes.
func (a *Auth) VerifySetup(ctx context.Context, userID, code string) error {
	secret, err := a.getMFASecret(ctx, userID)
	if err != nil {
		return err
	}
	if secret == "" {
		return fmt.Errorf("MFA not set up — call SetupMFA first")
	}

	if !totp.Validate(code, secret) {
		return ErrInvalidMFACode
	}

	if err = a.repo.EnableMFA(ctx, userID); err != nil {
		return fmt.Errorf("enable MFA: %w", err)
	}

	a.auditor.Write(ctx, audit.Entry{
		UserID: userID, Action: "user.mfa_enabled",
	})
	return nil
}

// VerifyMFA validates a TOTP code for login. Returns nil if valid.
func (a *Auth) VerifyMFA(ctx context.Context, userID, code string) error {
	secret, err := a.getMFASecret(ctx, userID)
	if err != nil {
		return err
	}
	if !totp.Validate(code, secret) {
		return ErrInvalidMFACode
	}
	return nil
}

// DisableMFA removes MFA for the user. Requires a valid TOTP code as confirmation.
func (a *Auth) DisableMFA(ctx context.Context, userID, code string) error {
	secret, err := a.getMFASecret(ctx, userID)
	if err != nil {
		return err
	}
	if !totp.Validate(code, secret) {
		return ErrInvalidMFACode
	}

	if err = a.repo.DisableMFA(ctx, userID); err != nil {
		return fmt.Errorf("disable MFA: %w", err)
	}

	a.auditor.Write(ctx, audit.Entry{
		UserID: userID, Action: "user.mfa_disabled",
	})
	return nil
}

// CheckMFARequired returns true if the user has MFA enabled.
func (a *Auth) CheckMFARequired(ctx context.Context, userID string) (bool, error) {
	enabled, err := a.repo.CheckMFAEnabled(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("check MFA status: %w", err)
	}
	return enabled, nil
}

func (a *Auth) getMFASecret(ctx context.Context, userID string) (string, error) {
	secret, err := a.repo.GetMFASecret(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("get MFA secret: %w", err)
	}
	if !secret.Valid {
		return "", nil
	}
	return secret.String, nil
}
