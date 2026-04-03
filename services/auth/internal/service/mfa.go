package service

import (
	"context"
	"fmt"

	"github.com/pquerna/otp/totp"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
)

// MFASetupResult holds the data needed for a user to configure their authenticator app.
type MFASetupResult struct {
	Secret string `json:"secret"` // base32-encoded TOTP secret
	URI    string `json:"uri"`    // otpauth:// URI for QR code generation
}

// SetupMFA generates a TOTP secret for the user and stores it (inactive until verified).
func (a *Auth) SetupMFA(ctx context.Context, userID string) (*MFASetupResult, error) {
	// Get user email for the TOTP issuer label
	var email string
	err := a.db.QueryRow(ctx, `SELECT email FROM users WHERE id = $1`, userID).Scan(&email)
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
	_, err = a.db.Exec(ctx,
		`UPDATE users SET mfa_secret = $2 WHERE id = $1`,
		userID, key.Secret(),
	)
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

	_, err = a.db.Exec(ctx,
		`UPDATE users SET mfa_enabled = true WHERE id = $1`,
		userID,
	)
	if err != nil {
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

	_, err = a.db.Exec(ctx,
		`UPDATE users SET mfa_enabled = false, mfa_secret = NULL WHERE id = $1`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("disable MFA: %w", err)
	}

	a.auditor.Write(ctx, audit.Entry{
		UserID: userID, Action: "user.mfa_disabled",
	})
	return nil
}

// CheckMFARequired returns true if the user has MFA enabled.
func (a *Auth) CheckMFARequired(ctx context.Context, userID string) (bool, error) {
	var enabled bool
	err := a.db.QueryRow(ctx,
		`SELECT COALESCE(mfa_enabled, false) FROM users WHERE id = $1`,
		userID,
	).Scan(&enabled)
	if err != nil {
		return false, fmt.Errorf("check MFA status: %w", err)
	}
	return enabled, nil
}

func (a *Auth) getMFASecret(ctx context.Context, userID string) (string, error) {
	var secret *string
	err := a.db.QueryRow(ctx,
		`SELECT mfa_secret FROM users WHERE id = $1`,
		userID,
	).Scan(&secret)
	if err != nil {
		return "", fmt.Errorf("get MFA secret: %w", err)
	}
	if secret == nil {
		return "", nil
	}
	return *secret, nil
}
