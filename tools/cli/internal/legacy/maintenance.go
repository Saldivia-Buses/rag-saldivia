package legacy

import "database/sql"

// ---------------------------------------------------------------------------
// Maintenance Assets (MANT_EQUIPOS)
// ---------------------------------------------------------------------------

// MaintenanceAssetReader creates a reader for MANT_EQUIPOS (maintainable equipment).
// Has auto-increment id_equipo PK. 2.6K rows.
// tipoequipo_id = equipment type, articulo_id = linked stock article,
// numero_serie = serial number, anio_fabricacion = year of manufacture,
// deposito_id = warehouse location, repuestos_criticos = critical spare parts notes.
func MaintenanceAssetReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "MANT_EQUIPOS",
		Target:     "erp_maintenance_assets",
		DomainName: "maintenance",
		PKColumn:   "id_equipo",
		Columns: "id_equipo, tipoequipo_id, nombre_equipo, numero_serie, " +
			"fecha_alta, fecha_baja, deposito_id, articulo_id, " +
			"anio_fabricacion, repuestos_criticos",
	}
}

// ---------------------------------------------------------------------------
// Maintenance Plans (MANT_PLAN)
// ---------------------------------------------------------------------------

// MaintenancePlanReader creates a reader for MANT_PLAN (scheduled maintenance plans).
// Has auto-increment id_plan PK. 1.5K rows.
// equipo_id = target equipment, accion_id = maintenance action type,
// rrule = recurrence rule (iCal format for periodic maintenance).
// Tracks scheduling: fecha_accion/hora_accion = planned start,
// fecha_terminacion/hora_terminacion = planned end,
// fecha_salida/hora_salida = equipment release,
// fecha_iniciotarea/hora_iniciotarea = actual task start.
func MaintenancePlanReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "MANT_PLAN",
		Target:     "erp_maintenance_plans",
		DomainName: "maintenance",
		PKColumn:   "id_plan",
		Columns: "id_plan, equipo_id, fecha_accion, hora_accion, " +
			"fecha_terminacion, hora_terminacion, fecha_salida, hora_salida, " +
			"accion_id, login, observaciones, rrule, " +
			"fecha_iniciotarea, hora_iniciotarea",
	}
}

// ---------------------------------------------------------------------------
// Maintenance Events (MANT_PLAN_EVENTOS)
// ---------------------------------------------------------------------------

// MaintenanceEventReader creates a reader for MANT_PLAN_EVENTOS (maintenance event instances).
// Has auto-increment id_planevento PK. 15K rows.
// Each row is one occurrence of a maintenance plan execution.
// plan_id = parent plan, accion_id = action performed.
func MaintenanceEventReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "MANT_PLAN_EVENTOS",
		Target:     "erp_work_orders",
		DomainName: "maintenance",
		PKColumn:   "id_planevento",
		Columns:    "id_planevento, plan_id, fecha_accion, accion_id, login, observaciones",
	}
}

// ---------------------------------------------------------------------------
// Vehicle Work (TRABAJOS_COCHE)
// ---------------------------------------------------------------------------

// VehicleWorkReader creates a reader for TRABAJOS_COCHE (third-party vehicle work).
// Composite PK (nrofab, idTrabajo). 7.3K rows.
// nrofab = vehicle/unit number, idTrabajo = work type code,
// realizador = contractor/person who did the work, importe = cost.
func VehicleWorkReader(db *sql.DB) *CompositeKeyReader {
	return &CompositeKeyReader{
		DB:         db,
		Table:      "TRABAJOS_COCHE",
		Target:     "erp_work_orders",
		DomainName: "maintenance",
		PKColumns:  []string{"nrofab", "idTrabajo"},
		Columns:    "nrofab, idTrabajo, fechaTrabajo, fechaPago, realizador, importe",
	}
}

// ---------------------------------------------------------------------------
// Fuel Logs (COMBUSTIBLE)
// ---------------------------------------------------------------------------

// FuelReader creates a reader for COMBUSTIBLE (fuel dispensing records).
// Has auto-increment id PK. 4K rows.
// nrofab_id = vehicle/unit, egreso_litros = liters dispensed,
// nombre/apellido = driver name, usuario = system user who recorded.
// Note: cometario (sic) is misspelled in the legacy schema.
func FuelReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "COMBUSTIBLE",
		Target:     "erp_fuel_logs",
		DomainName: "maintenance",
		PKColumn:   "id",
		Columns: "id, fecha, hora, egreso_litros, unidad, cometario, " +
			"nombre, apellido, usuario, nrofab_id",
	}
}
