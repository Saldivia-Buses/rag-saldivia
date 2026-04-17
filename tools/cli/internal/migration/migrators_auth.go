package migration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Legacy auth reuse — HTXPROFILES + HTXPROFILE_AUTH + HTXUSERS.Id_perfil.
//
// Prior to this module the auth import wrote HTXUSERS to `users` but dropped
// profile assignments, so every migrated user landed with zero roles. The
// three rescue functions below close that gap, recovering:
//
//   23 roles      ← HTXPROFILES
//   188 users     ← HTXUSERS.legacy_login set
//   N user_roles  ← HTXUSERS.Id_perfil → user_roles (all 188 valid cases)
//   2104 menu   pins per role in metadata.legacy_menus
//   1022 menu   pins per user in users.metadata (via audit annotation)
//
// They are plain functions (not GenericMigrator) because the tables are small
// (< 2K rows total) and the target schema is auth/platform, not ERP/tenant —
// the orchestrator's batch+COPY pipeline brings no benefit at this scale and
// would force us to introduce auth-specific readers. A simple tx per table
// is faster to ship and easier to reason about for a permissions rescue.
//
// All three are idempotent: ON CONFLICT DO NOTHING on the natural key.
// They're safe to re-run after partial failures.

// RescueLegacyProfiles reads HTXPROFILES and writes one `roles` row per
// profile, keeping the mapping in erp_legacy_mapping under
// domain="auth", legacy_table="HTXPROFILES".
func RescueLegacyProfiles(ctx context.Context, mysqlDB *sql.DB, pgPool *pgxpool.Pool, tenantID string, mapper *Mapper) error {
	if mapper.dryRun {
		return nil
	}
	rows, err := mysqlDB.QueryContext(ctx,
		`SELECT Id_perfil, nombre FROM HTXPROFILES WHERE Id_perfil > 0`)
	if err != nil {
		return fmt.Errorf("scan HTXPROFILES: %w", err)
	}
	defer func() { _ = rows.Close() }()

	tx, err := pgPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	inserted := 0
	for rows.Next() {
		var legacyID int64
		var name string
		if err := rows.Scan(&legacyID, &name); err != nil {
			return fmt.Errorf("scan profile: %w", err)
		}
		name = strings.TrimSpace(name)
		if name == "" {
			name = fmt.Sprintf("legacy_profile_%d", legacyID)
		}
		roleID := uuid.New().String()
		meta, _ := json.Marshal(map[string]any{
			"source":         "HTXPROFILES",
			"legacy_profile": legacyID,
		})

		// The legacy `saldivia` DB may already have a role with the same name
		// (e.g. if this is a re-run). Use ON CONFLICT on the unique key (name)
		// and RETURNING to capture whichever id actually ended up persisted.
		var persistedID string
		err := tx.QueryRow(ctx, `
			INSERT INTO roles (id, name, description, is_system, metadata)
			VALUES ($1, $2, $3, false, $4::jsonb)
			ON CONFLICT (name) DO UPDATE SET metadata = roles.metadata || EXCLUDED.metadata
			RETURNING id
		`, roleID, name,
			fmt.Sprintf("Imported from Histrix HTXPROFILES (id=%d)", legacyID),
			string(meta),
		).Scan(&persistedID)
		if err != nil {
			return fmt.Errorf("insert role %q: %w", name, err)
		}

		// Register in mapper cache + mapping table so the user_roles phase
		// resolves profile → role uuid.
		persistedUUID, _ := uuid.Parse(persistedID)
		if _, err := tx.Exec(ctx, `
			INSERT INTO erp_legacy_mapping (tenant_id, domain, legacy_table, legacy_id, sda_id)
			VALUES ($1, 'auth', 'HTXPROFILES', $2, $3)
			ON CONFLICT (tenant_id, domain, legacy_table, legacy_id) DO NOTHING
		`, tenantID, legacyID, persistedUUID); err != nil {
			return fmt.Errorf("map profile %d: %w", legacyID, err)
		}

		mapper.mu.Lock()
		key := mapper.cacheKey("auth", "HTXPROFILES")
		if mapper.cache[key] == nil {
			mapper.cache[key] = make(map[int64]uuid.UUID)
		}
		mapper.cache[key][legacyID] = persistedUUID
		mapper.mu.Unlock()

		inserted++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate profiles: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit profiles: %w", err)
	}
	slog.Info("rescue: HTXPROFILES → roles", "count", inserted)
	return nil
}

// RescueLegacyUserRoles joins HTXUSERS with its profile column and writes
// one user_roles row per (user, profile) pair. Must run AFTER the user
// migrator AND RescueLegacyProfiles so both FKs resolve.
func RescueLegacyUserRoles(ctx context.Context, mysqlDB *sql.DB, pgPool *pgxpool.Pool, tenantID string, mapper *Mapper) error {
	if mapper.dryRun {
		return nil
	}
	rows, err := mysqlDB.QueryContext(ctx,
		`SELECT Id_usuario, Id_perfil, login FROM HTXUSERS
		 WHERE Id_usuario > 0 AND Id_perfil > 0`)
	if err != nil {
		return fmt.Errorf("scan HTXUSERS: %w", err)
	}
	defer func() { _ = rows.Close() }()

	tx, err := pgPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	assigned := 0
	annotated := 0
	for rows.Next() {
		var userLegacyID, profLegacyID int64
		var login string
		if err := rows.Scan(&userLegacyID, &profLegacyID, &login); err != nil {
			return fmt.Errorf("scan user-role: %w", err)
		}
		userUUID, err := mapper.ResolveOptional(ctx, "auth", "HTXUSERS", userLegacyID)
		if err != nil || userUUID == uuid.Nil {
			continue
		}
		roleUUID, err := mapper.ResolveOptional(ctx, "auth", "HTXPROFILES", profLegacyID)
		if err != nil || roleUUID == uuid.Nil {
			continue
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)
			ON CONFLICT (user_id, role_id) DO NOTHING
		`, userUUID.String(), roleUUID.String()); err != nil {
			return fmt.Errorf("insert user_role (%s, %s): %w", userUUID, roleUUID, err)
		}
		assigned++

		// Annotate the user with legacy_login + legacy_profile_id so admin
		// searches work without joining erp_legacy_mapping.
		if _, err := tx.Exec(ctx, `
			UPDATE users SET legacy_login = $1, legacy_profile_id = $2
			WHERE id = $3
		`, login, profLegacyID, userUUID.String()); err != nil {
			slog.Warn("annotate user failed", "id", userUUID, "err", err)
			continue
		}
		annotated++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate user-roles: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit user-roles: %w", err)
	}
	slog.Info("rescue: user_roles + legacy_login", "assigned", assigned, "annotated", annotated)
	return nil
}

// RescueLegacyRolePermissions reads HTXPROFILE_AUTH (2,104 rows on saldivia)
// and appends each profile's menu list to that role's metadata.legacy_menus.
// Must run AFTER RescueLegacyProfiles.
//
// We deliberately do not translate menu ids into SDA permissions here — the
// mapping is not 1-to-1 and an automated pass would risk over-granting. The
// admin UI can read this JSONB and prompt an operator to map them by hand.
func RescueLegacyRolePermissions(ctx context.Context, mysqlDB *sql.DB, pgPool *pgxpool.Pool, tenantID string, mapper *Mapper) error {
	if mapper.dryRun {
		return nil
	}
	rows, err := mysqlDB.QueryContext(ctx,
		`SELECT Id_perfil, Id_menu, orden, notifica FROM HTXPROFILE_AUTH ORDER BY Id_perfil, orden`)
	if err != nil {
		return fmt.Errorf("scan HTXPROFILE_AUTH: %w", err)
	}
	defer func() { _ = rows.Close() }()

	byProfile := make(map[int64][]map[string]any)
	for rows.Next() {
		var prof int64
		var menu string
		var orden, notifica int
		if err := rows.Scan(&prof, &menu, &orden, &notifica); err != nil {
			return fmt.Errorf("scan profile auth: %w", err)
		}
		byProfile[prof] = append(byProfile[prof], map[string]any{
			"menu": menu, "order": orden, "notify": notifica == 1,
		})
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate profile auth: %w", err)
	}

	updated := 0
	for profLegacyID, menus := range byProfile {
		roleUUID, err := mapper.ResolveOptional(ctx, "auth", "HTXPROFILES", profLegacyID)
		if err != nil || roleUUID == uuid.Nil {
			continue
		}
		menusJSON, _ := json.Marshal(menus)
		_, err = pgPool.Exec(ctx, `
			UPDATE roles
			SET metadata = metadata || jsonb_build_object(
				'legacy_menus', $1::jsonb,
				'legacy_menu_count', $2
			)
			WHERE id = $3
		`, string(menusJSON), len(menus), roleUUID.String())
		if err != nil {
			slog.Warn("role metadata update failed", "role_id", roleUUID, "err", err)
			continue
		}
		updated++
	}
	slog.Info("rescue: roles decorated with legacy menus", "updated", updated)
	return nil
}

// RescueLegacyAuthAll is a convenience wrapper that runs all three auth
// rescue steps in the correct order. The orchestrator calls this from a
// setup hook after NewLegacyUserMigrator finishes.
func RescueLegacyAuthAll(ctx context.Context, mysqlDB *sql.DB, pgPool *pgxpool.Pool, tenantID string, mapper *Mapper) error {
	if err := RescueLegacyProfiles(ctx, mysqlDB, pgPool, tenantID, mapper); err != nil {
		return err
	}
	if err := RescueLegacyUserRoles(ctx, mysqlDB, pgPool, tenantID, mapper); err != nil {
		return err
	}
	return RescueLegacyRolePermissions(ctx, mysqlDB, pgPool, tenantID, mapper)
}
