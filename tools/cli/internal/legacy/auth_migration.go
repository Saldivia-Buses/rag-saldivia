package legacy

import "database/sql"

// ---------------------------------------------------------------------------
// Legacy Users (HTXUSERS)
// ---------------------------------------------------------------------------

// LegacyUserReader creates a reader for HTXUSERS (Histrix user accounts).
// Has auto-increment Id_usuario PK. 191 rows.
// login = username, password/pass = legacy passwords (plain/hashed),
// Id_perfil = profile/role FK, baja = deactivated flag,
// admin/editor/remote = permission flags, legajo_personal = linked employee.
// emailUser/emailPass = email account credentials (legacy inline storage).
func LegacyUserReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "HTXUSERS",
		Target:     "users",
		DomainName: "auth",
		PKColumn:   "Id_usuario",
		Columns: "Id_usuario, apellido, foto, seccion_id, Id_perfil, login, Nombre, " +
			"password, email, ultimolog, baja, pass, emailUser, emailPass, " +
			"editor, interno, telefono, admin, remote, minHour, maxHour, " +
			"idUsuario, theme, emailSignature, legajo_personal",
	}
}

// ---------------------------------------------------------------------------
// Legacy Profiles / Roles (HTXPROFILES)
// ---------------------------------------------------------------------------

// LegacyProfileReader creates a reader for HTXPROFILES (role/profile definitions).
// PK is Id_perfil (int, NOT auto-increment — manually assigned). 23 rows.
// nombre = profile/role display name.
func LegacyProfileReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "HTXPROFILES",
		Target:     "roles",
		DomainName: "auth",
		PKColumn:   "Id_perfil",
		Columns:    "Id_perfil, nombre",
	}
}

// ---------------------------------------------------------------------------
// Legacy Profile Authorizations (HTXPROFILE_AUTH)
// ---------------------------------------------------------------------------

// LegacyProfileAuthReader creates a reader for HTXPROFILE_AUTH (profile→menu permissions).
// Composite PK (Id_perfil, menuId). 2.1K rows.
// Id_menu = legacy menu path string, orden = display order,
// notifica = notification flag for this menu item.
func LegacyProfileAuthReader(db *sql.DB) *CompositeKeyReader {
	return &CompositeKeyReader{
		DB:         db,
		Table:      "HTXPROFILE_AUTH",
		Target:     "role_permissions",
		DomainName: "auth",
		PKColumns:  []string{"Id_perfil", "menuId"},
		Columns:    "Id_menu, Id_perfil, orden, notifica, menuId",
	}
}

// ---------------------------------------------------------------------------
// Legacy User Authorizations (HTXUSER_AUTH)
// ---------------------------------------------------------------------------

// LegacyUserAuthReader creates a reader for HTXUSER_AUTH (user-level menu overrides).
// Composite PK (login, menu_id). 1K rows.
// deny = 1 means this permission is denied for the user (override from profile).
func LegacyUserAuthReader(db *sql.DB) *CompositeKeyReader {
	return &CompositeKeyReader{
		DB:         db,
		Table:      "HTXUSER_AUTH",
		Target:     "user_roles",
		DomainName: "auth",
		PKColumns:  []string{"login", "menu_id"},
		Columns:    "login, menu_id, deny",
	}
}
