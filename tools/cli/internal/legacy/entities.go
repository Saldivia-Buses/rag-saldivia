package legacy

import "database/sql"

// EntityReader creates a reader for REG_CUENTA (clientes + proveedores).
// subsistema_id distinguishes: 01=compras(supplier), 02=ventas(customer), etc.
func EntityReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "REG_CUENTA",
		Target:     "erp_entities",
		DomainName: "entity",
		PKColumn:   "id_regcuenta",
		Columns:    "id_regcuenta, subsistema_id, nombre_cuenta, apellido_cuenta, cod_direccion, localidad_id, provincia_id, iva_id, tel_cuenta, cel_cuenta, email_cuenta, cuit_cuenta, razon_social, alta_cuenta, baja_cuenta, baja, nro_cuenta, nombre_fantasia, creator_user_id",
	}
}

// EmployeeReader creates a reader for PERSONAL (empleados).
func EmployeeReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "PERSONAL",
		Target:     "erp_entities",
		DomainName: "entity",
		PKColumn:   "IdPersona",
		Columns:    "IdPersona, legajo, Nombre, apellido, CUIL, email, telefono, celular, calle, numero, localidad, idProvincia, fecing, fechaEgreso, IdCateg, IdConvenio, Id_area, horario, IdSindicato",
	}
}

// ContactReader creates a reader for MED_CONTACTO.
// Note: MED_CONTACTO has no PK. We use a synthetic ROW_NUMBER approach.
// The reader returns all contacts for a given entity type.
func ContactReader(db *sql.DB) *GenericReader {
	// MED_CONTACTO has no PK, so we read all rows ordered by codigo (entity id).
	// Resume is based on a synthetic rowindex.
	return &GenericReader{
		DB:         db,
		Table:      "MED_CONTACTO",
		Target:     "erp_entity_contacts",
		DomainName: "entity",
		PKColumn:   "codigo", // Not a real PK but the best we have for ordering
		Columns:    "codigo, id_tipo, tipo_ente, valor",
	}
}
