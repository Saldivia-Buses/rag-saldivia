package legacy

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
)

// ---------------------------------------------------------------------------
// Vehicles / Chassis (CHASIS)
// ---------------------------------------------------------------------------

// VehicleReader creates a reader for CHASIS (vehicle/chassis registry).
// PK is nrocha (mediumint, NOT auto-increment — it's the fabrication number). 4K rows.
// Tracks chassis from entry to exit: entrada/salida = factory entry/exit dates,
// fecalt = registration date, fecent = delivery date, fecter = completion date.
// chasis = VIN, nromotor = engine number, marcod = brand, modcod = model.
// ctacod = customer entity, conces = dealer/concession.
// chequeo = inspection status, devolucion = return date if applicable.
func VehicleReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "CHASIS",
		Target:     "erp_units",
		DomainName: "production",
		PKColumn:   "nrocha",
		Columns: "nrocha, chades, chaobs, chasis, conces, ctacod, entrada, " +
			"fecalt, fecent, fecter, modcha, salida, marcod, modcod, nromotor, " +
			"chequeo, idTaco, nroTaco, diasTaco, fechaActualiza, idMarcaMotor, " +
			"devolucion, chequeo_arranque, rfid_tarjeta_id, referencia_factura",
	}
}

// ---------------------------------------------------------------------------
// Production Inspections (PROD_CONTROLES)
// ---------------------------------------------------------------------------

// ProductionInspectionReader creates a reader for PROD_CONTROLES (quality control checkpoints).
// Has auto-increment id_prodcontrol PK. 2K rows.
// Each row defines a control point in the production line.
// seccion_id = production section, tipo_control = control type,
// critico = critical flag, accionable = can trigger actions,
// legajo_defecto = default employee, aviso_produccion = production alert flag.
func ProductionInspectionReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "PROD_CONTROLES",
		Target:     "erp_production_inspections",
		DomainName: "production",
		PKColumn:   "id_prodcontrol",
		Columns: "id_prodcontrol, seccion_id, nombre_control, seccion_coche_id, " +
			"tipo_control, habilitado, critico, orden_control, accionable, " +
			"obs_control, modelo_control, legajo_defecto, usuario_habilitado, " +
			"ver_ft, aviso_produccion",
	}
}

// ---------------------------------------------------------------------------
// Production Inspection Details (PROD_CONTROL_MOVIM)
// ---------------------------------------------------------------------------

// ProductionInspectionDetailReader creates a reader for PROD_CONTROL_MOVIM
// (individual inspection results per vehicle per control point).
// Has auto-increment id_controlmovim PK. 1.7M rows.
// Uses a JOIN with PROD_CONTROLES to bring in the control's orden_control (production order context).
// conforme/no_conforme/no_aplica = result flags, horas_retrabajo = rework hours,
// controlcausal_id = root cause category, legajo_personal/legajo_realizo = employee IDs.
type ProductionInspectionDetailReader struct {
	DB *sql.DB
}

func (r *ProductionInspectionDetailReader) LegacyTable() string { return "PROD_CONTROL_MOVIM" }
func (r *ProductionInspectionDetailReader) SDATable() string     { return "erp_production_inspections" }
func (r *ProductionInspectionDetailReader) Domain() string       { return "production" }

func (r *ProductionInspectionDetailReader) ReadBatch(ctx context.Context, resumeKey string, limit int) ([]LegacyRow, string, error) {
	lastID := int64(0)
	if resumeKey != "" {
		lastID, _ = strconv.ParseInt(resumeKey, 10, 64)
	}

	query := `SELECT m.id_controlmovim, m.prodcontrol_id, m.nrofab_id, m.valor_control,
		m.observacion_control, m.fecha_control, m.conforme, m.no_conforme, m.no_aplica,
		m.hora_control, m.controlado, m.user_id, m.foto, m.horas_retrabajo,
		m.controlcausal_id, m.legajo_personal, m.acciones_control, m.legajo_realizo,
		m.seccion_origen_id, m.seccion_calidad_id,
		c.orden_control as control_orden_id
		FROM PROD_CONTROL_MOVIM m
		LEFT JOIN PROD_CONTROLES c ON m.prodcontrol_id = c.id_prodcontrol
		WHERE m.id_controlmovim > ?
		ORDER BY m.id_controlmovim LIMIT ?`

	rows, err := r.DB.QueryContext(ctx, query, lastID, limit)
	if err != nil {
		return nil, "", fmt.Errorf("read PROD_CONTROL_MOVIM: %w", err)
	}
	defer func() { _ = rows.Close() }()

	cols, err := rows.Columns()
	if err != nil {
		return nil, "", fmt.Errorf("columns PROD_CONTROL_MOVIM: %w", err)
	}

	var result []LegacyRow
	var lastKey int64
	for rows.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, "", fmt.Errorf("scan PROD_CONTROL_MOVIM: %w", err)
		}
		row := make(LegacyRow, len(cols))
		for i, col := range cols {
			row[col] = values[i]
		}
		result = append(result, row)
		lastKey = row.Int64("id_controlmovim")
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("iterate PROD_CONTROL_MOVIM: %w", err)
	}

	return result, strconv.FormatInt(lastKey, 10), nil
}

// NewProductionInspectionDetailReader creates a ProductionInspectionDetailReader.
func NewProductionInspectionDetailReader(db *sql.DB) *ProductionInspectionDetailReader {
	return &ProductionInspectionDetailReader{DB: db}
}

// ---------------------------------------------------------------------------
// Production Steps / Processes (PROD_PROCESOS)
// ---------------------------------------------------------------------------

// ProductionStepReader creates a reader for PROD_PROCESOS (production process definitions).
// Has auto-increment id_proceso PK. 1.6K rows.
// procesopadre_id = parent process (tree structure), stkdeposito_id = linked warehouse,
// homologacion_id = vehicle homologation type this process applies to.
func ProductionStepReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "PROD_PROCESOS",
		Target:     "erp_production_steps",
		DomainName: "production",
		PKColumn:   "id_proceso",
		Columns: "id_proceso, nombre_proceso, descripcion_proceso, orden_proceso, " +
			"stkdeposito_id, homologacion_id, procesopadre_id",
	}
}

// ---------------------------------------------------------------------------
// Production Requests (MRP_PEDIDO_PRODUCCION)
// ---------------------------------------------------------------------------

// ProductionRequestReader creates a reader for MRP_PEDIDO_PRODUCCION (production material requests).
// Has auto-increment id_mrp_pedido_prod PK. 13K rows.
// Links a specific article to a vehicle (nrofab_id) at a manufacturing position.
// Tracks workflow: creacion → inicio → terminado with timestamps and user IDs.
// carpetaplanos_id = drawing folder, posicionfab_alarma_id = position alarm threshold.
func ProductionRequestReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "MRP_PEDIDO_PRODUCCION",
		Target:     "erp_production_orders",
		DomainName: "production",
		PKColumn:   "id_mrp_pedido_prod",
		Columns: "id_mrp_pedido_prod, carpetaplanos_id, seccion_id, nrofab_id, " +
			"posicionfab_alarma_id, posicionfab_id, stkarticulo_id, " +
			"cantidad_pieza_unidad, creacion_usuario_id, creacion_hora, " +
			"inicio_usuario_id, inicio_hora, terminado_usuario_id, terminado_hora",
	}
}
