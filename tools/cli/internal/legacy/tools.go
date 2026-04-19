package legacy

import "database/sql"

// ---------------------------------------------------------------------------
// Tools / serialized inventory tags (HERRAMIENTAS + HERRMOVS)
// ---------------------------------------------------------------------------

// ToolReader creates a reader for HERRAMIENTAS (389,253 rows live — Pareto #4
// of the Phase 1 §Data migration gap post-2.0.9, ~15 % of uncovered row
// volume). Despite the "herramientas" name, the table is the serialized
// inventory tag ledger: one row per physical item received, each with its
// own unique code (id_herramienta varchar(25)) stamped on the item and an
// AI id_etiqueta PK. Live XML-form scrape shows it used across recepcion/,
// almacen/, herramientas/, mantenimiento/, help_local/.
//
// artcod is the FK into STK_ARTICULOS (which has a composite PK with
// subsistema). Same pattern as HomologationRevisionLine: resolve via
// articleCompositeCode(artcod, "") default-subsystem lookup at transform
// time; rows whose artcod isn't in the current stock catalog keep
// article_id NULL with article_code preserved.
func ToolReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "HERRAMIENTAS",
		Target:     "erp_tools",
		DomainName: "tools",
		PKColumn:   "id_etiqueta",
		Columns: "id_etiqueta, id_herramienta, artcod, caract, grucod, invcod, nomherr, " +
			"tipoherr, ocpnro, ocpfec, remfec, remnpv, remnro, ctacod, codest, " +
			"pendiente_oc, observacion, nrofab, generada",
	}
}

// ToolMovementReader creates a reader for HERRMOVS (11,680 rows live). Each
// row is one check-out / check-in / damage-return / loan entry. movher
// references the 4-row CONCHERR enum (1=Devol. Rotura, 2=Devolucion,
// 3=A Cargo, 7=Prestamo — movtip 1=IN, 2=OUT). Inlined as concept_code on
// the target rather than a separate lookup table.
//
// id_herramienta joins HERRAMIENTAS.id_herramienta primarily; some rows
// (~13 % — 1,566 of 11,680) join MANT_EQUIPOS.numero_serie instead
// (mixed-use lending ledger). Orphan movements migrate with tool_id NULL
// and the raw tool_code preserved.
func ToolMovementReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "HERRMOVS",
		Target:     "erp_tool_movements",
		DomainName: "tools",
		PKColumn:   "id_herrmovs",
		Columns:    "id_herrmovs, cantidad, id_herramienta, movfec, movher, usuario",
	}
}
