package legacy

import "database/sql"

// EmployeeDetailReader reads extended PERSONAL fields for erp_employee_details.
func EmployeeDetailReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "PERSONAL",
		Target:     "erp_employee_details",
		DomainName: "hr",
		PKColumn:   "IdPersona",
		Columns:    "IdPersona, Id_area, horario, fecing, fechaEgreso, IdConvenio, IdSindicato, IdCateg, Id_perfil",
	}
}

// AbsenceReader creates a reader for FRANCOS_PER (ausencias/licencias).
func AbsenceReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "FRANCOS_PER",
		Target:     "erp_hr_events",
		DomainName: "hr",
		PKColumn:   "id",
		Columns:    "id, IdPersona, idfranco, fechainicio, fechafin, diashabiles, horas, observaciones, tipo, franco_tipo_id",
	}
}

// TrainingReader creates a reader for RH_CURSOS (capacitaciones).
func TrainingReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "RH_CURSOS",
		Target:     "erp_training",
		DomainName: "hr",
		PKColumn:   "Id_curso",
		Columns:    "Id_curso, descripcion, inicio, fin, situacion, contenido, modalidad",
	}
}

// TrainingAttendeeReader creates a reader for RH_CURSO_REALIZADO.
func TrainingAttendeeReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "RH_CURSO_REALIZADO",
		Target:     "erp_training_attendees",
		DomainName: "hr",
		PKColumn:   "id_curso_realizado",
		Columns:    "id_curso_realizado, Id_curso, IdPersona, calificacion, presente",
	}
}
