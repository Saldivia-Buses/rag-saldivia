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

// ContactReader is deferred — MED_CONTACTO has no PK, needs synthetic key approach.
// Will be implemented when entity contact migration is prioritized.
