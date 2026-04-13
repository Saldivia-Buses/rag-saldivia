package legacy

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
)

// ---------------------------------------------------------------------------
// Accidents (ACCIDENTES)
// ---------------------------------------------------------------------------

// AccidentReader creates a reader for ACCIDENTES (workplace accident statistics).
// Composite PK (idAnio, idmes). 13 rows.
// Monthly aggregates: accidentes = count, hstrab = hours worked,
// diasper = days lost, diasacc = accident days, persoprod = production staff count.
func AccidentReader(db *sql.DB) *CompositeKeyReader {
	return &CompositeKeyReader{
		DB:         db,
		Table:      "ACCIDENTES",
		Target:     "erp_work_accidents",
		DomainName: "safety",
		PKColumns:  []string{"idAnio", "idmes"},
		Columns:    "idAnio, idmes, accidentes, hstrab, diasper, diasacc, persoprod",
	}
}

// ---------------------------------------------------------------------------
// Accident Persons (ACCIDENTE_PER)
// ---------------------------------------------------------------------------

// AccidentPersonReader creates a reader for ACCIDENTE_PER (accident-affected employees).
// No auto-increment PK — uses synthetic row ordering. 22 rows.
// legajo = employee file number, IdPersona = person ID, idaccidente = accident ID,
// sector = department, fechaini/fechafin = injury period.
type AccidentPersonReader struct {
	DB *sql.DB
}

func (r *AccidentPersonReader) LegacyTable() string { return "ACCIDENTE_PER" }
func (r *AccidentPersonReader) SDATable() string     { return "erp_work_accidents" }
func (r *AccidentPersonReader) Domain() string       { return "safety" }

func (r *AccidentPersonReader) ReadBatch(ctx context.Context, resumeKey string, limit int) ([]LegacyRow, string, error) {
	// ACCIDENTE_PER has no auto-increment PK and no unique ordering column.
	// With only 22 rows we use LIMIT/OFFSET via a synthetic row counter.
	offset := int64(0)
	if resumeKey != "" {
		offset, _ = strconv.ParseInt(resumeKey, 10, 64)
	}

	query := `SELECT legajo, idaccidente, fechaini, sector, IdPersona, fechafin, observaciones
		FROM ACCIDENTE_PER ORDER BY IdPersona, fechaini LIMIT ? OFFSET ?`

	rows, err := r.DB.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, "", fmt.Errorf("read ACCIDENTE_PER: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, "", fmt.Errorf("columns ACCIDENTE_PER: %w", err)
	}

	var result []LegacyRow
	for rows.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, "", fmt.Errorf("scan ACCIDENTE_PER: %w", err)
		}
		row := make(LegacyRow, len(cols))
		for i, col := range cols {
			row[col] = values[i]
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("iterate ACCIDENTE_PER: %w", err)
	}

	nextOffset := offset + int64(len(result))
	return result, strconv.FormatInt(nextOffset, 10), nil
}

// NewAccidentPersonReader creates an AccidentPersonReader.
func NewAccidentPersonReader(db *sql.DB) *AccidentPersonReader {
	return &AccidentPersonReader{DB: db}
}

// ---------------------------------------------------------------------------
// Risk Agents (RIESGOS)
// ---------------------------------------------------------------------------

// RiskAgentReader creates a reader for RIESGOS (occupational risk agent catalog).
// PK is idRiesgo (int, NOT auto-increment — manually assigned). 212 rows.
// agente = risk agent name/description.
func RiskAgentReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "RIESGOS",
		Target:     "erp_risk_agents",
		DomainName: "safety",
		PKColumn:   "idRiesgo",
		Columns:    "idRiesgo, agente",
	}
}

// ---------------------------------------------------------------------------
// Risk Exposures (RIESGO_PERSONAL)
// ---------------------------------------------------------------------------

// RiskExposureReader creates a reader for RIESGO_PERSONAL (employee risk exposure periods).
// Has auto-increment id_riesgopersonal PK. 2.9K rows.
// riesgo_id = risk agent, persona_id / IdPersona = employee (dual FK columns),
// fecha_desde/fecha_hasta = exposure period.
func RiskExposureReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "RIESGO_PERSONAL",
		Target:     "erp_employee_risk_exposures",
		DomainName: "safety",
		PKColumn:   "id_riesgopersonal",
		Columns:    "id_riesgopersonal, riesgo_id, persona_id, fecha_desde, fecha_hasta, stamp, IdPersona",
	}
}

// ---------------------------------------------------------------------------
// Medical Leaves (PARTE_MEDICO_DIARIO)
// ---------------------------------------------------------------------------

// MedicalLeaveReader creates a reader for PARTE_MEDICO_DIARIO (daily medical reports).
// Has auto-increment id PK. 59 rows.
// sintomatologia = symptoms, prescripcion = prescription/treatment,
// nombre = patient name (denormalized), usuario = recording user.
func MedicalLeaveReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "PARTE_MEDICO_DIARIO",
		Target:     "erp_medical_leaves",
		DomainName: "safety",
		PKColumn:   "id",
		Columns:    "id, fecha, sintomatologia, prescripcion, hora, usuario, nombre",
	}
}
