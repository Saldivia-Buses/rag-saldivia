package legacy

import "database/sql"

// ---------------------------------------------------------------------------
// Departments / Org Chart (ORGANIGRAMA)
// ---------------------------------------------------------------------------

// DepartmentReader creates a reader for ORGANIGRAMA (organizational chart / departments).
// Composite PK (id_seccion, idPadre) — both varchar. 14 rows.
// id_seccion = department/section name, idPadre = parent department (tree structure),
// idUsuario = responsible user, color = UI display color.
func DepartmentReader(db *sql.DB) *CompositeKeyReader {
	return &CompositeKeyReader{
		DB:         db,
		Table:      "ORGANIGRAMA",
		Target:     "erp_departments",
		DomainName: "hr",
		PKColumns:  []string{"id_seccion", "idPadre"},
		Columns:    "id_seccion, idPadre, idUsuario, color",
	}
}

// ---------------------------------------------------------------------------
// Attendance (FICHADADIA)
// ---------------------------------------------------------------------------

// AttendanceReader creates a reader for FICHADADIA (daily attendance / time clock).
// Composite PK (tarjeta, fecha). 933K rows.
// tarjeta = badge/card number, legajo = employee file number.
// h1..h4 = time clock punches (up to 4 per day), trabajadas = hours worked,
// notrabajadas = hours not worked, hingreso/hegreso = entry/exit times.
// feriado = holiday flag, max_extras_50 = max overtime hours at 50% rate.
func AttendanceReader(db *sql.DB) *CompositeKeyReader {
	return &CompositeKeyReader{
		DB:         db,
		Table:      "FICHADADIA",
		Target:     "erp_attendance",
		DomainName: "hr",
		PKColumns:  []string{"tarjeta", "fecha"},
		Columns: "tarjeta, legajo, fecha, h1, h2, h3, h4, " +
			"trabajadas, notrabajadas, hingreso, hegreso, feriado, max_extras_50",
	}
}

// ---------------------------------------------------------------------------
// Training Attendees (RH_CURSO_REALIZADO)
// ---------------------------------------------------------------------------

// TrainingAttendeeReader creates a reader for RH_CURSO_REALIZADO (training course attendance).
// Has auto-increment id_curso_realizado PK. 6.1K rows.
// Links employees to training courses with their results.
// calificacion = employee grade, calif_curso = course overall grade,
// presente = attendance flag, Id_centro = training center.
func TrainingAttendeeReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "RH_CURSO_REALIZADO",
		Target:     "erp_training_attendees",
		DomainName: "hr",
		PKColumn:   "id_curso_realizado",
		Columns: "id_curso_realizado, calificacion, calif_curso, Id_curso, " +
			"Id_usuario, IdPersona, presente, Id_centro",
	}
}

// ---------------------------------------------------------------------------
// Demerits (MOVDEMERITO)
// ---------------------------------------------------------------------------

// DemeritReader creates a reader for MOVDEMERITO (demerit point movements).
// Has auto-increment id_movdemerito PK. 284K rows.
// coddem = demerit type code, demfec = demerit date, movfec = movement date,
// artcod = article code (if material-related), prvcod = supplier code.
// nrocomp/tipocomp = voucher reference, cpsmovim_id = linked purchase movement.
func DemeritReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "MOVDEMERITO",
		Target:     "erp_hr_events",
		DomainName: "hr",
		PKColumn:   "id_movdemerito",
		Columns: "id_movdemerito, artcod, coddem, demfec, movfec, nrocomp, " +
			"prvcod, puesto, tipocomp, cpsmovim_id",
	}
}

// ---------------------------------------------------------------------------
// Deductions (RHDESCUENTOS)
// ---------------------------------------------------------------------------

// DeductionReader creates a reader for RHDESCUENTOS (payroll deductions).
// Has auto-increment idDesc PK. 16.7K rows.
// legajo = employee file number, importe = deduction amount,
// idMotivoDesc = deduction reason code, sobre_extras = applies to overtime flag.
func DeductionReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "RHDESCUENTOS",
		Target:     "erp_hr_events",
		DomainName: "hr",
		PKColumn:   "idDesc",
		Columns:    "idDesc, legajo, fecha, importe, idMotivoDesc, sobre_extras",
	}
}

// ---------------------------------------------------------------------------
// Additional Pay (RRHH_ADICIONALES)
// ---------------------------------------------------------------------------

// AdditionalPayReader creates a reader for RRHH_ADICIONALES (additional pay / bonuses).
// Composite PK (legajo, desdeFecha). 1.9K rows.
// porcentaje = percentage-based bonus, fijo = fixed amount,
// maximo = cap amount, hastaFecha = validity end date.
func AdditionalPayReader(db *sql.DB) *CompositeKeyReader {
	return &CompositeKeyReader{
		DB:         db,
		Table:      "RRHH_ADICIONALES",
		Target:     "erp_hr_events",
		DomainName: "hr",
		PKColumns:  []string{"legajo", "desdeFecha"},
		Columns:    "legajo, descripcion, porcentaje, fijo, maximo, desdeFecha, hastaFecha",
	}
}

// ---------------------------------------------------------------------------
// Employee cards (PERSONAL_TARJETA) — date-versioned card-to-employee assignments
// ---------------------------------------------------------------------------

// EmployeeCardReader creates a reader for PERSONAL_TARJETA (card-to-employee
// assignments, 1,403 rows). The table has no single PK — it's a composite
// (idPersona, tarjeta, fechaDesde) where the same card can reassign across
// employees over time. FICHADAS (raw clock events) depend on this table to
// resolve card_code + event_date → employee at that point in time.
func EmployeeCardReader(db *sql.DB) *CompositeKeyReader {
	return &CompositeKeyReader{
		DB:         db,
		Table:      "PERSONAL_TARJETA",
		Target:     "erp_employee_cards",
		DomainName: "hr",
		PKColumns:  []string{"idPersona", "tarjeta", "fechaDesde"},
		Columns:    "idPersona, tarjeta, fechaDesde",
	}
}

// ---------------------------------------------------------------------------
// Raw clock-punch events (FICHADAS)
// ---------------------------------------------------------------------------

// TimeClockEventReader creates a reader for FICHADAS (raw clock-punch events,
// 1,465,002 rows — Pareto #2 of the Phase 1 §Data migration gap, 41 % of
// uncovered row volume post-2.0.8). Each row is one marcaje from a physical
// terminal. AI PK id_fichada + UNIQUE insertkey (MD5-ish dedup hash). Employee
// resolution goes through tarjeta + fecha → PERSONAL_TARJETA (date-versioned)
// → PERSONAL → erp_entities; orphan tarjetas migrate with entity_id NULL.
func TimeClockEventReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "FICHADAS",
		Target:     "erp_time_clock_events",
		DomainName: "hr",
		PKColumn:   "id_fichada",
		Columns:    "id_fichada, tarjeta, fecha, hora, reloj, codigo, marca, borrado, insertkey",
	}
}
