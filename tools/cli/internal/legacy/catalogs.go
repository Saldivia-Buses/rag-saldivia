package legacy

import "database/sql"

// CatalogMapping defines how a legacy lookup table maps to erp_catalogs.
type CatalogMapping struct {
	LegacyTable string
	CatalogType string
	PKColumn    string // AI int column for ordering/resume
	CodeColumn  string
	NameColumn  string
	ExtraWhere  string
}

// DefaultCatalogMappings returns the standard Histrix → SDA catalog mappings.
// PKColumn is the auto-increment integer used for ordering and resume.
func DefaultCatalogMappings() []CatalogMapping {
	return []CatalogMapping{
		{"GEN_PROVINCIAS", "province", "id_provincia", "id_provincia", "nombre_provincia", ""},
		{"GEN_LOCALIDADES", "city", "id_localidad", "id_localidad", "nombre_localidad", ""},
		{"GEN_MONEDAS", "currency", "id_moneda", "id_moneda", "descripcion_moneda", ""},
		{"FORMAPAGO", "payment_method", "formaPago", "formaPago", "nombrePago", ""},
		{"GEN_IVA", "iva_condition", "id_iva", "id_iva", "nombre_iva", ""},
		{"GEN_IVA_ALICUOTAS", "iva_rate", "id_iva_alicuota", "id_iva_alicuota", "alicuota", ""},
		{"GEN_UNIDADES", "unit", "id_unidad", "id_unidad", "nombre_unidad", ""},
		{"GEN_COMPROBANTES", "voucher_type", "id_codigo", "id_codigo", "descripcion", ""},
		{"GEN_TIPOS_DOCUMENTOS", "document_type", "id_tipo_documento", "id_tipo_documento", "nombre_tipo_documento", ""},
		{"GEN_ESTADOCIVIL", "civil_status", "id_estadocivil", "id_estadocivil", "nombre_estadocivil", ""},
		{"GEN_NACIONALIDADES", "nationality", "id_nacionalidad", "id_nacionalidad", "nacionalidad", ""},
		{"GEN_OBRASSOCIALES", "health_plan", "id_osocial", "id_osocial", "nombre_osocial", ""},
		{"GEN_PARENTESCOS", "kinship", "id_parentesco", "id_parentesco", "nombre_parentesco", ""},
		{"GEN_PROFESIONES", "profession", "id_profesion", "id_profesion", "nombre", ""},
		{"GEN_CALLES", "street", "id_calle", "id_calle", "nombre_calle", ""},
		{"GEN_TIPO_CONTACTOS", "contact_type", "id_contacto", "id_contacto", "tipo_contacto", ""},
		{"GEN_IIBB", "iibb_regime", "id_iibb", "id_iibb", "nombre_iibb", ""},
	}
}

// CatalogReader creates a GenericReader for a catalog mapping.
func CatalogReader(db *sql.DB, m CatalogMapping) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      m.LegacyTable,
		Target:     "erp_catalogs",
		DomainName: "catalog",
		PKColumn:   m.PKColumn,
		Columns:    m.PKColumn + ", " + m.CodeColumn + ", " + m.NameColumn,
		ExtraWhere: m.ExtraWhere,
	}
}
