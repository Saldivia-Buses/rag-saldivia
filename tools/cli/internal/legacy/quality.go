package legacy

import "database/sql"

// ---------------------------------------------------------------------------
// Nonconformities (CAL_NOCONFORMIDADES)
// ---------------------------------------------------------------------------

// NonconformityReader creates a reader for CAL_NOCONFORMIDADES (quality nonconformities).
// Has auto-increment id_noconformidad PK. 720 rows.
// descripcion_nconf = problem description, contencion_nconf = containment action,
// alcance_nconf = scope/extent, causa_nconf = root cause analysis.
// estadonconf_id = workflow state, origennconf_id = origin type,
// tiponconf_id = nonconformity type, eficaz = effectiveness flag.
// costo_nconf = associated cost, regcuenta_id = supplier/customer entity.
// auditoriaobs_id = linked audit finding (if originated from audit).
func NonconformityReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "CAL_NOCONFORMIDADES",
		Target:     "erp_nonconformities",
		DomainName: "quality",
		PKColumn:   "id_noconformidad",
		Columns: "id_noconformidad, fecha_nconf, rhseccion_id, descripcion_nconf, " +
			"contencion_nconf, alcance_nconf, login_id, estadonconf_id, " +
			"causa_nconf, responsable_nconf, fecha_analisisnconf, origennconf_id, " +
			"fecha_cierrenconf, cuenta_id, stkarticulo_id, subsistema_id, " +
			"regcuenta_id, auditoriaobs_id, costo_nconf, eficaz, tiponconf_id, " +
			"unidad_id, demerito_id, tarea_servicie_id",
	}
}

// ---------------------------------------------------------------------------
// Corrective Actions (CAL_ACCIONES_NCONF)
// ---------------------------------------------------------------------------

// CorrectiveActionReader creates a reader for CAL_ACCIONES_NCONF (corrective/preventive actions).
// Has auto-increment id_accionnconf PK. 2.8K rows.
// Linked to a nonconformity via noconformidad_id.
// tipo_accionnconf = action type (corrective vs preventive),
// control_accionnconf = verification status, terminada = completed flag.
// resp_rhpersonal_id = responsible employee.
func CorrectiveActionReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "CAL_ACCIONES_NCONF",
		Target:     "erp_corrective_actions",
		DomainName: "quality",
		PKColumn:   "id_accionnconf",
		Columns: "id_accionnconf, noconformidad_id, descripcion_accionnconf, " +
			"fecha_accionnconf, fecha_realizacion, fecha_terminada, " +
			"resp_rhpersonal_id, control_accionnconf, tipo_accionnconf, " +
			"login, terminada",
	}
}

// ---------------------------------------------------------------------------
// Internal Audits (CAL_AUDITORIA_INT)
// ---------------------------------------------------------------------------

// AuditReader creates a reader for CAL_AUDITORIA_INT (internal quality audits).
// Has auto-increment id_auditoriaint PK. 273 rows.
// documento_id = audited document/process reference,
// realizada = completed flag, auditoriaintpunt_id = audit score/rating.
func AuditReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "CAL_AUDITORIA_INT",
		Target:     "erp_audits",
		DomainName: "quality",
		PKColumn:   "id_auditoriaint",
		Columns: "id_auditoriaint, fecha_auditoriaint, codigo_auditorint, " +
			"documento_id, realizada, auditoriaintpunt_id",
	}
}

// ---------------------------------------------------------------------------
// Audit Findings (CAL_AUDITORIA_OBS)
// ---------------------------------------------------------------------------

// AuditFindingReader creates a reader for CAL_AUDITORIA_OBS (audit observations/findings).
// Has auto-increment id_auditoriaobs PK. 487 rows.
// Linked to audit via auditoriaint_id, respauditoriaint_id = responsible auditor,
// hallazgotipo_id = finding type (observation, minor NC, major NC),
// realizada = resolved flag.
func AuditFindingReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "CAL_AUDITORIA_OBS",
		Target:     "erp_audit_findings",
		DomainName: "quality",
		PKColumn:   "id_auditoriaobs",
		Columns: "id_auditoriaobs, auditoriaint_id, respauditoriaint_id, " +
			"hallazgotipo_id, observacion, realizada",
	}
}

// ---------------------------------------------------------------------------
// Controlled Documents (CAL_DOCUMENTOS)
// ---------------------------------------------------------------------------

// ControlledDocumentReader creates a reader for CAL_DOCUMENTOS (ISO controlled documents).
// Has auto-increment id_documento PK. 559 rows.
// Tracks document lifecycle: codigo_documento = doc code, vigencia_documento = validity date,
// estado_documento = status, revision_documento = current revision.
// emisor_documento/aprueba_documento = issuer/approver (employee IDs).
// distribuidor_documento = distribution list, conservacion_documento = retention policy.
func ControlledDocumentReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "CAL_DOCUMENTOS",
		Target:     "erp_controlled_documents",
		DomainName: "quality",
		PKColumn:   "id_documento",
		Columns: "id_documento, codigo_documento, tipodocumento_id, nombre_documento, " +
			"vigencia_documento, distribuidor_documento, aplicacion_documento, " +
			"emisor_documento, conservacion_documento, guarda_documento, " +
			"revision_documento, estado_documento, archivo_documento, " +
			"aprueba_documento, menuId",
	}
}
