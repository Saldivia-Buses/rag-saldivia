package intelligence

// defaultDomains returns the full domain registry.
// 12 root domains + key subdomains. Structured as Go literals (type-safe, zero deps).
func defaultDomains() []Domain {
	// Shared technique lists
	predictiveCore := []string{TechTransits, TechProfections, TechFirdaria, TechZR, TechSolarArc, TechPrimaryDir}
	predictiveExpected := []string{
		TechProgressions, TechSolarReturn, TechLunarReturn, TechEclipses,
		TechFixedStars, TechStations, TechConvergence, TechDecennials,
		TechTertiaryProg, TechFastTransits, TechLunations, TechPlanetCycles,
		TechAlmuten, TechLots, TechDisposition, TechSect, TechMidpoints,
		TechDeclinations, TechActivChains, TechTimingWindows,
	}
	predictiveBrief := []TechniqueWeight{
		{TechPrimaryDir, 0.9, ""}, {TechSolarArc, 0.9, ""},
		{TechTransits, 0.8, ""}, {TechProfections, 0.8, ""},
		{TechFirdaria, 0.7, ""}, {TechZR, 0.7, ""},
		{TechDecennials, 0.6, ""}, {TechProgressions, 0.6, ""},
		{TechEclipses, 0.6, ""}, {TechSolarReturn, 0.5, ""},
		{TechStations, 0.5, ""}, {TechLunations, 0.4, ""},
		{TechPlanetCycles, 0.4, ""}, {TechMidpoints, 0.3, ""},
		{TechConvergence, 1.0, ""}, {TechActivChains, 0.8, ""},
	}

	return []Domain{
		// ══════════════════════════════════════════════════════════
		// ROOT DOMAINS (Parent = "")
		// ══════════════════════════════════════════════════════════

		{
			ID: "natal", Name: "Carta Natal", Parent: "",
			TechniquesRequired: []string{TechNatal, TechAlmuten, TechDisposition, TechSect},
			TechniquesExpected: []string{TechFixedStars, TechLots, TechHyleg, TechTemperament, TechMelothesia, TechMidpoints, TechDeclinations},
			TechniquesBrief: []TechniqueWeight{
				{TechAlmuten, 1.0, ""}, {TechDisposition, 0.9, ""}, {TechSect, 0.8, ""},
				{TechFixedStars, 0.7, ""}, {TechLots, 0.6, ""}, {TechHyleg, 0.5, ""},
				{TechMidpoints, 0.5, ""}, {TechDeclinations, 0.4, ""},
			},
			Keywords: []string{"carta natal", "carta", "natal", "nacimiento", "personalidad", "quien soy", "como soy"},
		},
		{
			ID: "predictivo", Name: "Predictivo General", Parent: "",
			TechniquesRequired: predictiveCore,
			TechniquesExpected: predictiveExpected,
			TechniquesBrief:    predictiveBrief,
			Keywords:           []string{"prediccion", "año", "que viene", "futuro", "pronostico", "que me espera", "como viene"},
		},
		{
			ID: "empresa", Name: "Empresa/Negocio", Parent: "",
			TechniquesRequired: append(predictiveCore, TechLots),
			TechniquesExpected: append(predictiveExpected, TechSynastry, TechComposite),
			TechniquesBrief: []TechniqueWeight{
				{TechTransits, 1.0, "casa_2,casa_8,casa_10"},
				{TechSolarArc, 0.9, "casa_10,mc"},
				{TechPrimaryDir, 0.9, "mc,casa_2"},
				{TechProfections, 0.8, ""}, {TechFirdaria, 0.7, ""},
				{TechLots, 0.8, "comercio,prosperidad"},
				{TechEclipses, 0.6, ""}, {TechPlanetCycles, 0.7, "saturno,jupiter"},
				{TechConvergence, 1.0, ""}, {TechTimingWindows, 0.9, ""},
			},
			Precautions: []string{"No dar cifras exactas de dinero", "Timing de negocios es orientativo"},
			Keywords:    []string{"empresa", "negocio", "compañia", "sociedad", "comercio", "ventas", "facturacion"},
		},
		{
			ID: "carrera", Name: "Carrera Profesional", Parent: "predictivo",
			TechniquesRequired: []string{TechProfections},
			TechniquesBrief: []TechniqueWeight{
				{TechPrimaryDir, 1.0, "mc,casa_10"}, {TechSolarArc, 1.0, "mc,casa_10"},
				{TechTransits, 0.9, "casa_10,saturno"}, {TechProfections, 0.9, ""},
				{TechFirdaria, 0.7, ""}, {TechSolarReturn, 0.6, "mc"},
				{TechConvergence, 1.0, ""}, {TechTimingWindows, 0.8, ""},
			},
			Precautions: []string{"Saturno transitando Casa 10 = reestructuración, no necesariamente negativo"},
			Keywords:    []string{"trabajo", "carrera", "profesion", "empleo", "jefe", "ascenso", "promocion", "vocacion", "casa 10"},
		},
		{
			ID: "salud", Name: "Salud", Parent: "predictivo",
			TechniquesRequired: []string{TechTemperament, TechMelothesia},
			TechniquesBrief: []TechniqueWeight{
				{TechTransits, 1.0, "casa_6,casa_12,casa_8"}, {TechSolarArc, 0.9, "casa_6"},
				{TechTemperament, 0.9, ""}, {TechMelothesia, 0.8, ""},
				{TechHyleg, 0.7, ""}, {TechProfections, 0.7, ""},
				{TechEclipses, 0.6, "casa_6,casa_12"}, {TechLunations, 0.5, ""},
				{TechConvergence, 1.0, ""},
			},
			Precautions: []string{
				"NUNCA diagnosticar enfermedades — solo tendencias energéticas",
				"Siempre recomendar consultar con profesional de salud",
				"Melotesia es orientativa, no médica",
			},
			Keywords: []string{"salud", "enfermedad", "cuerpo", "cirugia", "hospital", "medico", "dolor", "casa 6", "casa 12"},
		},
		{
			ID: "amor", Name: "Amor y Relaciones", Parent: "predictivo",
			TechniquesBrief: []TechniqueWeight{
				{TechTransits, 1.0, "casa_7,venus"}, {TechSolarArc, 0.9, "casa_7,venus"},
				{TechPrimaryDir, 0.8, "venus,casa_7"}, {TechProfections, 0.8, ""},
				{TechSolarReturn, 0.7, "casa_7"}, {TechLots, 0.7, "eros,matrimonio"},
				{TechEclipses, 0.5, "casa_7"}, {TechConvergence, 1.0, ""},
			},
			Keywords: []string{"amor", "pareja", "relacion", "matrimonio", "divorcio", "novio", "novia", "casa 7", "venus"},
		},
		{
			ID: "dinero", Name: "Dinero y Finanzas", Parent: "predictivo",
			TechniquesBrief: []TechniqueWeight{
				{TechTransits, 1.0, "casa_2,casa_8,jupiter"}, {TechSolarArc, 0.9, "casa_2"},
				{TechProfections, 0.8, ""}, {TechLots, 0.8, "prosperidad,adquisicion"},
				{TechPlanetCycles, 0.7, "jupiter"}, {TechSolarReturn, 0.6, "casa_2"},
				{TechConvergence, 1.0, ""},
			},
			Precautions: []string{"No dar cifras exactas", "Indicar tendencias, no garantías"},
			Keywords:    []string{"dinero", "plata", "finanzas", "inversion", "ahorro", "deuda", "casa 2", "casa 8"},
		},
		{
			ID: "familia", Name: "Familia y Hogar", Parent: "predictivo",
			TechniquesBrief: []TechniqueWeight{
				{TechTransits, 1.0, "casa_4,luna,casa_5"}, {TechSolarArc, 0.9, "casa_4"},
				{TechProfections, 0.8, ""}, {TechLunations, 0.7, "casa_4"},
				{TechLots, 0.6, "padre,madre,hijos"}, {TechConvergence, 1.0, ""},
			},
			Keywords: []string{"familia", "hijo", "hija", "padre", "madre", "hogar", "casa", "mudanza", "casa 4", "casa 5"},
		},
		{
			ID: "competitivo", Name: "Competencia y Estrategia", Parent: "predictivo",
			TechniquesBrief: []TechniqueWeight{
				{TechTransits, 1.0, "marte,casa_7"}, {TechSolarArc, 0.9, "marte"},
				{TechProfections, 0.8, ""}, {TechTimingWindows, 0.9, ""},
				{TechConvergence, 1.0, ""},
			},
			Keywords: []string{"competencia", "rival", "enemigo", "juicio", "demanda", "litigio", "conflicto"},
		},
		{
			ID: "electiva", Name: "Astrología Electiva", Parent: "",
			TechniquesRequired: []string{TechElectional},
			TechniquesBrief: []TechniqueWeight{
				{TechElectional, 1.0, ""}, {TechFastTransits, 0.8, ""},
				{TechLunations, 0.7, ""}, {TechTransits, 0.6, ""},
			},
			Precautions: []string{"Fechas electivas son sugerencias, no garantías"},
			Keywords:    []string{"mejor fecha", "cuando", "electiva", "elegir dia", "buen momento", "momento ideal"},
		},
		{
			ID: "horaria", Name: "Astrología Horaria", Parent: "",
			TechniquesRequired: []string{TechHorary},
			TechniquesBrief: []TechniqueWeight{
				{TechHorary, 1.0, ""},
			},
			Precautions: []string{"Verificar radicalidad antes de juzgar", "Luna VOC = nada se concreta"},
			Keywords:    []string{"horaria", "pregunta si/no", "se va a", "va a pasar", "pregunta concreta"},
		},
		{
			ID: "relocation", Name: "Relocalización", Parent: "",
			TechniquesRequired: []string{TechACG},
			TechniquesExpected: []string{TechSolarReturn},
			TechniquesBrief: []TechniqueWeight{
				{TechACG, 1.0, ""}, {TechSolarReturn, 0.8, ""},
			},
			Keywords: []string{"mudarme", "relocation", "ciudad", "pais", "donde vivir", "emigrar"},
		},

		// ══════════════════════════════════════════════════════════
		// ROOT DOMAINS (additional)
		// ══════════════════════════════════════════════════════════

		{
			ID: "sinastria", Name: "Sinastría", Parent: "",
			TechniquesRequired: []string{TechSynastry, TechSolarArc, TechProfections, TechFirdaria},
			TechniquesExpected: []string{TechComposite, TechNodes, TechPrimaryDir, TechSolarReturn},
			TechniquesBrief: []TechniqueWeight{
				{TechSynastry, 1.0, ""}, {TechNodes, 0.8, ""}, {TechSolarArc, 0.8, ""},
				{TechComposite, 0.7, ""}, {TechProfections, 0.7, ""}, {TechFirdaria, 0.6, ""},
				{TechPrimaryDir, 0.5, ""}, {TechSolarReturn, 0.5, ""},
			},
			Keywords: []string{"sinastria entre", "compatibilidad astral", "carta compuesta"},
		},
		{
			ID: "rectificacion", Name: "Rectificación", Parent: "",
			TechniquesRequired: []string{TechSolarArc, TechPrimaryDir, TechProfections},
			TechniquesExpected: []string{TechFirdaria, TechEclipses},
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechPrimaryDir, 1.0, ""},
				{TechProfections, 0.9, ""}, {TechFirdaria, 0.5, ""}, {TechEclipses, 0.5, ""},
			},
			Keywords: []string{"rectificar", "hora exacta de nacimiento", "corregir hora natal"},
		},

		// ══════════════════════════════════════════════════════════
		// SUBDOMAINS — Natal
		// ══════════════════════════════════════════════════════════

		{
			ID: "personal", Name: "Lectura Personal", Parent: "natal",
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechPrimaryDir, 0.9, ""}, {TechProfections, 0.85, ""},
				{TechFirdaria, 0.8, ""}, {TechSolarReturn, 0.8, ""}, {TechEclipses, 0.75, ""},
				{TechDivisor, 0.7, ""}, {TechTriplicity, 0.65, ""}, {TechConvergence, 0.7, ""},
				{TechProgressions, 0.6, ""}, {TechFastTransits, 0.5, ""},
			},
			Precautions: []string{"Mercury Rx = revisar decisiones importantes"},
			Keywords:    []string{"lectura personal", "que me espera este año", "mi futuro", "prediccion personal"},
		},
		{
			ID: "espiritualidad", Name: "Espiritualidad", Parent: "natal",
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechZR, 0.85, ""}, {TechProfections, 0.85, ""},
				{TechFirdaria, 0.8, ""}, {TechEclipses, 0.8, ""}, {TechPrimaryDir, 0.75, ""},
				{TechSolarReturn, 0.7, ""}, {TechProgressions, 0.7, ""}, {TechPrenatalEcl, 0.5, ""},
				{TechLots, 0.6, ""}, {TechChronocrator, 0.6, ""}, {TechLilith, 0.5, ""},
				{TechAlmuten, 0.5, ""}, {TechSabian, 0.45, ""},
			},
			Keywords: []string{"espiritual", "crecimiento personal", "proposito", "mision de vida", "karma"},
		},
		{
			ID: "viajes", Name: "Viajes y Mudanza", Parent: "natal",
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechRelocation, 0.9, ""}, {TechProfections, 0.85, ""},
				{TechPrimaryDir, 0.8, ""}, {TechSolarReturn, 0.8, ""}, {TechFirdaria, 0.7, ""},
				{TechEclipses, 0.75, ""}, {TechZR, 0.65, ""}, {TechConvergence, 0.7, ""},
				{TechProgressions, 0.6, ""}, {TechVertex, 0.5, ""},
			},
			Precautions: []string{"Mercury Rx = revisar logística de viaje"},
			Keywords:    []string{"viaje", "irse del pais", "cambio de ciudad", "vacaciones", "viajar"},
		},

		// ══════════════════════════════════════════════════════════
		// SUBDOMAINS — Empresa (18 new)
		// ══════════════════════════════════════════════════════════

		{
			ID: "empresa_ventas", Name: "Ventas y Comercio", Parent: "empresa",
			TechniquesBrief: []TechniqueWeight{
				{TechTransits, 1.0, "casa_2,casa_7,mercurio"}, {TechTimingWindows, 1.0, ""},
				{TechFastTransits, 0.7, "mercurio,venus"}, {TechConvergence, 1.0, ""},
			},
			Keywords: []string{"venta", "cliente", "cotizacion", "propuesta"},
		},
		{
			ID: "empresa_rrhh", Name: "RRHH y Equipo", Parent: "empresa",
			TechniquesBrief: []TechniqueWeight{
				{TechSynastry, 1.0, ""}, {TechTransits, 0.8, "casa_6,casa_11"},
				{TechTimingWindows, 0.8, ""}, {TechConvergence, 0.9, ""},
			},
			Keywords: []string{"empleado", "contratar", "despedir", "equipo", "rrhh"},
		},
		{
			ID: "empresa_finanzas", Name: "Finanzas Empresariales", Parent: "empresa",
			TechniquesBrief: []TechniqueWeight{
				{TechTransits, 1.0, "casa_2,casa_8,jupiter,saturno"},
				{TechPlanetCycles, 0.9, "jupiter,saturno"}, {TechLots, 0.8, "prosperidad"},
				{TechConvergence, 1.0, ""},
			},
			Keywords: []string{"flujo de caja", "cash flow", "cobranza empresarial", "deuda empresarial"},
		},
		{
			ID: "negociacion", Name: "Negociación Comercial", Parent: "empresa",
			TechniquesRequired: []string{TechMercuryRx},
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechPrimaryDir, 0.95, ""}, {TechMercuryRx, 0.95, ""},
				{TechCorpHouses, 0.9, ""}, {TechProfections, 0.85, ""}, {TechTimingWindows, 0.85, ""},
				{TechFastTransits, 0.85, ""}, {TechEclipses, 0.85, ""}, {TechElectional, 0.8, ""},
				{TechStations, 0.8, ""}, {TechSolarReturn, 0.8, ""}, {TechConvergence, 0.8, ""},
				{TechFirdaria, 0.75, ""}, {TechLunations, 0.75, ""}, {TechLots, 0.7, ""},
			},
			Precautions: []string{"Mercury Rx = no firmar contratos", "Luna VOC = acuerdo no se concreta"},
			Keywords:    []string{"negociar", "cerrar trato", "acuerdo comercial", "contrato comercial", "negociacion"},
		},
		{
			ID: "deudores", Name: "Deudores y Cobranzas", Parent: "empresa",
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechSynastry, 0.9, ""}, {TechPrimaryDir, 0.9, ""},
				{TechCorpHouses, 0.85, ""}, {TechProfections, 0.85, ""}, {TechFastTransits, 0.85, ""},
				{TechMercuryRx, 0.8, ""}, {TechFirdaria, 0.8, ""}, {TechSolarReturn, 0.8, ""},
				{TechLunations, 0.8, ""}, {TechEclipses, 0.75, ""}, {TechConvergence, 0.7, ""},
				{TechLots, 0.7, ""},
			},
			Precautions: []string{"Mercury Rx = demoras en cobros"},
			Keywords:    []string{"cobrar", "deudor", "moroso", "intimacion", "debe plata"},
		},
		{
			ID: "cash_flow", Name: "Flujo de Caja", Parent: "empresa",
			TechniquesRequired: []string{TechMercuryRx},
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechPrimaryDir, 0.95, ""}, {TechFastTransits, 0.9, ""},
				{TechCorpHouses, 0.9, ""}, {TechPlanetReturns, 0.85, ""}, {TechMercuryRx, 0.85, ""},
				{TechProfections, 0.85, ""}, {TechPlanetCycles, 0.8, ""}, {TechEclipses, 0.8, ""},
				{TechConvergence, 0.75, ""}, {TechSolarReturn, 0.75, ""}, {TechLots, 0.75, ""},
				{TechLunations, 0.7, ""}, {TechFirdaria, 0.7, ""}, {TechStations, 0.7, ""},
			},
			Precautions: []string{"Mercury Rx = cuidar cash flow"},
			Keywords:    []string{"liquidez", "pagos", "cash", "flujo de fondos"},
		},
		{
			ID: "lanzamiento", Name: "Lanzamiento de Producto", Parent: "empresa",
			TechniquesRequired: []string{TechMercuryRx, TechEclipses},
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechMercuryRx, 0.95, ""}, {TechPlanetCycles, 0.9, ""},
				{TechEclipses, 0.9, ""}, {TechTimingWindows, 0.85, ""}, {TechProfections, 0.85, ""},
				{TechCorpHouses, 0.85, ""}, {TechElectional, 0.8, ""}, {TechStations, 0.8, ""},
				{TechSolarReturn, 0.8, ""}, {TechConvergence, 0.8, ""}, {TechLunations, 0.8, ""},
				{TechFirdaria, 0.75, ""}, {TechEclTriggers, 0.75, ""},
			},
			Precautions: []string{"Mercury Rx = no lanzar", "Luna VOC = no lanzar"},
			Keywords:    []string{"lanzar", "lanzamiento", "producto nuevo", "innovacion", "sacar al mercado"},
		},
		{
			ID: "screening", Name: "Evaluación de Candidatos", Parent: "empresa",
			TechniquesRequired: []string{TechSynastry},
			TechniquesBrief: []TechniqueWeight{
				{TechSynastry, 1.0, ""}, {TechSolarArc, 0.9, ""}, {TechCorpHouses, 0.85, ""},
				{TechProfections, 0.85, ""}, {TechPlanetReturns, 0.7, ""},
				{TechFirdaria, 0.7, ""}, {TechSolarReturn, 0.7, ""}, {TechConvergence, 0.6, ""},
			},
			Keywords: []string{"candidato", "evaluar persona", "compatibilidad laboral", "screening"},
		},
		{
			ID: "reunion_socios", Name: "Reunión de Socios", Parent: "empresa",
			TechniquesRequired: []string{TechEclipses},
			TechniquesBrief: []TechniqueWeight{
				{TechSynastry, 0.9, ""}, {TechSolarArc, 1.0, ""}, {TechPrimaryDir, 0.9, ""},
				{TechTimingWindows, 0.85, ""}, {TechCorpHouses, 0.85, ""},
				{TechProfections, 0.85, ""}, {TechEclipses, 0.85, ""}, {TechElectional, 0.9, ""},
				{TechMercuryRx, 0.8, ""}, {TechSolarReturn, 0.8, ""}, {TechComposite, 0.8, ""},
				{TechConvergence, 0.75, ""}, {TechFirdaria, 0.7, ""}, {TechLunations, 0.7, ""},
			},
			Precautions: []string{"Mercury Rx = postergar decisiones clave", "Luna VOC = no decidir"},
			Keywords:    []string{"reunion socios", "junta directiva", "asamblea", "decision grupal"},
		},
		{
			ID: "sucesion", Name: "Sucesión Generacional", Parent: "empresa",
			TechniquesRequired: []string{TechEclipses, TechZR},
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechPlanetReturns, 0.95, ""}, {TechCorpHouses, 0.9, ""},
				{TechEclipses, 0.9, ""}, {TechPrimaryDir, 0.9, ""}, {TechSynastry, 0.9, ""},
				{TechZR, 0.85, ""}, {TechProfections, 0.85, ""}, {TechPlanetCycles, 0.85, ""},
				{TechConvergence, 0.8, ""}, {TechFirdaria, 0.8, ""}, {TechSolarReturn, 0.75, ""},
			},
			Keywords: []string{"sucesion", "generacional", "retirar", "jubilacion", "dejar la empresa", "heredar empresa"},
		},
		{
			ID: "sociedad", Name: "Sociedad y Socios", Parent: "empresa",
			TechniquesRequired: []string{TechSynastry},
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechSynastry, 0.9, ""}, {TechCorpHouses, 0.9, ""},
				{TechPrimaryDir, 0.9, ""}, {TechPlanetReturns, 0.85, ""}, {TechLots, 0.85, ""},
				{TechProfections, 0.85, ""}, {TechPlanetCycles, 0.8, ""}, {TechSolarReturn, 0.8, ""},
				{TechFirdaria, 0.8, ""}, {TechComposite, 0.8, ""}, {TechEclipses, 0.75, ""},
				{TechConvergence, 0.7, ""}, {TechMercuryRx, 0.7, ""},
			},
			Precautions: []string{"Mercury Rx = no firmar acuerdos societarios"},
			Keywords:    []string{"socio", "sociedad", "partner", "socios", "accionista"},
		},
		{
			ID: "timing", Name: "Timing Puro", Parent: "empresa",
			TechniquesRequired: []string{TechFastTransits},
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechPrimaryDir, 0.95, ""}, {TechConvergence, 0.9, ""},
				{TechFastTransits, 0.9, ""}, {TechTimingWindows, 0.85, ""}, {TechEclipses, 0.85, ""},
				{TechMercuryRx, 0.85, ""}, {TechElectional, 0.8, ""}, {TechProfections, 0.8, ""},
				{TechStations, 0.8, ""}, {TechLunarReturn, 0.8, ""}, {TechSolarReturn, 0.7, ""},
				{TechChronocrator, 0.6, ""}, {TechFirdaria, 0.65, ""},
			},
			Precautions: []string{"Mercury Rx = postergar", "Luna VOC = no actuar", "Estaciones = revisar"},
			Keywords:    []string{"cuando hacer", "mejor momento empresarial", "timing comercial", "fecha ideal para negocio"},
		},
		{
			ID: "riesgos", Name: "Riesgos Empresariales", Parent: "empresa",
			TechniquesRequired: []string{TechEclipses, TechStations},
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechEclipses, 0.95, ""}, {TechEclTriggers, 0.9, ""},
				{TechCorpHouses, 0.9, ""}, {TechPrimaryDir, 0.9, ""}, {TechFastTransits, 0.85, ""},
				{TechProfections, 0.85, ""}, {TechPlanetReturns, 0.85, ""}, {TechConvergence, 0.8, ""},
				{TechPlanetCycles, 0.8, ""}, {TechStations, 0.8, ""}, {TechSolarReturn, 0.75, ""},
				{TechFirdaria, 0.7, ""}, {TechMercuryRx, 0.7, ""},
			},
			Keywords: []string{"riesgo", "peligro", "crisis empresarial", "exposicion", "proteger empresa"},
		},
		{
			ID: "enterprise_general", Name: "Análisis Empresarial General", Parent: "empresa",
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechCorpHouses, 0.95, ""}, {TechPlanetReturns, 0.9, ""},
				{TechPrimaryDir, 0.9, ""}, {TechEclipses, 0.85, ""}, {TechEclTriggers, 0.85, ""},
				{TechProfections, 0.85, ""}, {TechPlanetCycles, 0.85, ""}, {TechZR, 0.8, ""},
				{TechFirdaria, 0.8, ""}, {TechSolarReturn, 0.8, ""}, {TechConvergence, 0.8, ""},
				{TechMercuryRx, 0.7, ""}, {TechDivisor, 0.7, ""}, {TechLunations, 0.7, ""},
			},
			Keywords: []string{"empresa en general", "panorama empresarial", "como viene la empresa"},
		},
		{
			ID: "produccion", Name: "Producción y Operaciones", Parent: "empresa",
			TechniquesRequired: []string{TechMercuryRx, TechStations},
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechProfections, 0.9, ""}, {TechMercuryRx, 0.9, ""},
				{TechConvergence, 0.9, ""}, {TechPrimaryDir, 0.8, ""}, {TechFastTransits, 0.8, ""},
				{TechCorpHouses, 0.8, ""}, {TechTimingWindows, 0.8, ""}, {TechTransits, 0.7, ""},
				{TechSolarReturn, 0.7, ""}, {TechFirdaria, 0.7, ""}, {TechStations, 0.7, ""},
				{TechPlanetCycles, 0.7, ""}, {TechLunations, 0.6, ""}, {TechEclipses, 0.6, ""},
			},
			Precautions: []string{"Mercury Rx = paradas de producción", "Estaciones = revisar procesos"},
			Keywords:    []string{"produccion", "operaciones", "planta", "fabricacion", "supply chain"},
		},
		{
			ID: "proveedores", Name: "Proveedores", Parent: "empresa",
			TechniquesBrief: []TechniqueWeight{
				{TechMercuryRx, 1.0, ""}, {TechSolarArc, 0.9, ""}, {TechTimingWindows, 0.9, ""},
				{TechProfections, 0.9, ""}, {TechCorpHouses, 0.8, ""}, {TechFastTransits, 0.8, ""},
				{TechConvergence, 0.8, ""}, {TechSolarReturn, 0.7, ""}, {TechPrimaryDir, 0.7, ""},
				{TechFirdaria, 0.7, ""}, {TechStations, 0.7, ""}, {TechLunations, 0.6, ""},
			},
			Precautions: []string{"Mercury Rx = demoras en entregas"},
			Keywords:    []string{"proveedor", "proveedores", "compras", "abastecimiento"},
		},
		{
			ID: "licitaciones", Name: "Licitaciones", Parent: "empresa",
			TechniquesBrief: []TechniqueWeight{
				{TechTimingWindows, 1.0, ""}, {TechSolarArc, 1.0, ""}, {TechConvergence, 0.9, ""},
				{TechProfections, 0.9, ""}, {TechMercuryRx, 0.9, ""}, {TechSolarReturn, 0.8, ""},
				{TechPrimaryDir, 0.8, ""}, {TechFastTransits, 0.7, ""}, {TechCorpHouses, 0.7, ""},
				{TechEclipses, 0.7, ""}, {TechLunations, 0.7, ""}, {TechFirdaria, 0.7, ""},
				{TechStations, 0.6, ""},
			},
			Precautions: []string{"Mercury Rx = no presentar"},
			Keywords:    []string{"licitacion", "concurso", "pliego", "oferta publica"},
		},
		{
			ID: "expansion", Name: "Expansión Empresarial", Parent: "empresa",
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechConvergence, 0.9, ""}, {TechProfections, 0.9, ""},
				{TechTimingWindows, 0.9, ""}, {TechPlanetCycles, 0.8, ""}, {TechSolarReturn, 0.8, ""},
				{TechTransits, 0.8, ""}, {TechPrimaryDir, 0.8, ""}, {TechMercuryRx, 0.8, ""},
				{TechFirdaria, 0.8, ""}, {TechEclipses, 0.7, ""}, {TechCorpHouses, 0.7, ""},
				{TechZR, 0.6, ""}, {TechLunations, 0.6, ""},
			},
			Precautions: []string{"Mercury Rx = no abrir sucursales", "Eclipses = precaución"},
			Keywords:    []string{"expandir", "sucursal", "abrir nuevo local", "crecer", "escalar"},
		},
		{
			ID: "calidad", Name: "Control de Calidad", Parent: "empresa",
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 0.9, ""}, {TechProfections, 0.9, ""}, {TechMercuryRx, 0.9, ""},
				{TechCorpHouses, 0.8, ""}, {TechConvergence, 0.8, ""}, {TechTimingWindows, 0.8, ""},
				{TechFastTransits, 0.8, ""}, {TechStations, 0.7, ""}, {TechSolarReturn, 0.7, ""},
				{TechFirdaria, 0.6, ""},
			},
			Precautions: []string{"Mercury Rx = auditorías de calidad"},
			Keywords:    []string{"calidad", "control de calidad", "norma", "certificacion", "iso"},
		},
		{
			ID: "logistica", Name: "Logística y Distribución", Parent: "empresa",
			TechniquesBrief: []TechniqueWeight{
				{TechMercuryRx, 1.0, ""}, {TechTimingWindows, 0.9, ""}, {TechSolarArc, 0.9, ""},
				{TechFastTransits, 0.8, ""}, {TechStations, 0.8, ""}, {TechConvergence, 0.8, ""},
				{TechProfections, 0.8, ""}, {TechCorpHouses, 0.7, ""}, {TechSolarReturn, 0.7, ""},
				{TechLunations, 0.6, ""}, {TechFirdaria, 0.6, ""},
			},
			Precautions: []string{"Mercury Rx = demoras logísticas", "Estaciones = problemas de transporte"},
			Keywords:    []string{"logistica", "distribucion", "transporte", "envio", "entrega", "flota"},
		},
		{
			ID: "empresa_competitivo", Name: "Competencia de Mercado", Parent: "empresa",
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechSynastry, 0.9, ""}, {TechPrimaryDir, 0.9, ""},
				{TechProfections, 0.85, ""}, {TechFirdaria, 0.8, ""}, {TechSolarReturn, 0.8, ""},
				{TechEclipses, 0.75, ""}, {TechFastTransits, 0.65, ""}, {TechZR, 0.6, ""},
				{TechConvergence, 0.7, ""},
			},
			Precautions: []string{"Mercury Rx = no lanzar contra competencia"},
			Keywords:    []string{"competencia entre empresas", "market share", "competidor directo", "ganar mercado"},
		},

		// ══════════════════════════════════════════════════════════
		// SUBDOMAINS — Carrera (5 new + 1 existing)
		// ══════════════════════════════════════════════════════════

		{
			ID: "carrera_vocacion", Name: "Vocación", Parent: "carrera",
			TechniquesBrief: []TechniqueWeight{
				{TechAlmuten, 1.0, ""}, {TechVocational, 0.9, ""}, {TechDisposition, 0.9, ""},
				{TechSect, 0.8, ""}, {TechProfections, 0.7, ""},
			},
			Keywords: []string{"vocacion", "que estudiar", "para que soy bueno", "talento"},
		},
		{
			ID: "educacion", Name: "Educación y Estudios", Parent: "carrera",
			TechniquesRequired: []string{TechZR},
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechVocational, 0.9, ""}, {TechPrimaryDir, 0.85, ""},
				{TechZR, 0.85, ""}, {TechProfections, 0.85, ""}, {TechSolarReturn, 0.8, ""},
				{TechFirdaria, 0.8, ""}, {TechEclipses, 0.7, ""}, {TechDivisor, 0.7, ""},
				{TechTriplicity, 0.65, ""}, {TechConvergence, 0.7, ""},
			},
			Precautions: []string{"Mercury Rx = revisar inscripciones"},
			Keywords:    []string{"estudiar", "carrera universitaria", "examen", "rendir", "educacion", "posgrado"},
		},
		{
			ID: "legal", Name: "Legal y Juicios", Parent: "carrera",
			TechniquesRequired: []string{TechEclipses},
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechPrimaryDir, 0.9, ""}, {TechMercuryRx, 0.9, ""},
				{TechHorary, 0.85, ""}, {TechProfections, 0.85, ""}, {TechSolarReturn, 0.8, ""},
				{TechFirdaria, 0.8, ""}, {TechElectional, 0.8, ""}, {TechEclipses, 0.75, ""},
				{TechStations, 0.7, ""}, {TechConvergence, 0.7, ""}, {TechFastTransits, 0.65, ""},
			},
			Precautions: []string{"Mercury Rx = no firmar", "Luna VOC = juicio no procede"},
			Keywords:    []string{"abogado", "legal", "contrato legal", "expediente", "sentencia", "escribano"},
		},
		{
			ID: "creatividad", Name: "Creatividad y Arte", Parent: "carrera",
			TechniquesRequired: []string{TechZR},
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechVocational, 0.9, ""}, {TechZR, 0.85, ""},
				{TechProfections, 0.85, ""}, {TechSolarReturn, 0.8, ""}, {TechFirdaria, 0.8, ""},
				{TechPrimaryDir, 0.75, ""}, {TechProgressions, 0.7, ""}, {TechEclipses, 0.7, ""},
				{TechConvergence, 0.65, ""}, {TechSabian, 0.45, ""},
			},
			Keywords: []string{"arte", "creativo", "proyecto artistico", "musica", "pintura", "escritura"},
		},
		{
			ID: "fama", Name: "Fama y Reputación", Parent: "carrera",
			TechniquesRequired: []string{TechZR},
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechPrimaryDir, 0.9, ""}, {TechSolarReturn, 0.85, ""},
				{TechZR, 0.85, ""}, {TechProfections, 0.85, ""}, {TechEclipses, 0.8, ""},
				{TechFirdaria, 0.8, ""}, {TechTriplicity, 0.6, ""}, {TechDivisor, 0.65, ""},
				{TechConvergence, 0.75, ""}, {TechFixedStars, 0.5, ""},
			},
			Keywords: []string{"fama", "reconocimiento", "reputacion", "imagen publica", "hacerme conocido"},
		},
		{
			ID: "relacion_laboral", Name: "Relación Laboral", Parent: "carrera",
			TechniquesRequired: []string{TechSynastry},
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechProfections, 0.85, ""}, {TechSynastry, 0.8, ""},
				{TechSolarReturn, 0.8, ""}, {TechFirdaria, 0.8, ""}, {TechPrimaryDir, 0.75, ""},
				{TechFastTransits, 0.7, ""}, {TechEclipses, 0.7, ""}, {TechConvergence, 0.65, ""},
				{TechMercuryRx, 0.6, ""}, {TechZR, 0.6, ""},
			},
			Precautions: []string{"Mercury Rx = revisar comunicación con jefe"},
			Keywords:    []string{"jefe", "ambiente laboral", "relacion con jefe", "compañero de trabajo"},
		},

		// ══════════════════════════════════════════════════════════
		// SUBDOMAINS — Salud (5 new + 2 existing)
		// ══════════════════════════════════════════════════════════

		{
			ID: "salud_cirugia", Name: "Cirugía", Parent: "salud",
			TechniquesBrief: []TechniqueWeight{
				{TechElectional, 1.0, ""}, {TechLunations, 0.9, ""}, {TechMercuryRx, 0.9, ""},
				{TechSolarArc, 0.85, ""}, {TechConvergence, 0.8, ""}, {TechMelothesia, 0.8, ""},
			},
			Precautions: []string{"Evitar cirugía con Luna en signo del órgano afectado", "Mercury Rx = revisar diagnóstico"},
			Keywords:    []string{"cirugia", "operacion", "operar", "intervencion quirurgica"},
		},
		{
			ID: "salud_mental", Name: "Salud Mental", Parent: "salud",
			TechniquesBrief: []TechniqueWeight{
				{TechTransits, 1.0, "luna,neptuno,casa_12"}, {TechLunations, 0.8, ""},
				{TechTemperament, 0.7, ""}, {TechConvergence, 0.9, ""}, {TechChiron, 0.7, ""},
			},
			Precautions: []string{"Siempre recomendar terapeuta profesional"},
			Keywords:    []string{"ansiedad", "depresion", "estres", "mental", "emocional", "psicologico"},
		},
		{
			ID: "crisis", Name: "Crisis y Emergencia Personal", Parent: "salud",
			TechniquesRequired: []string{TechFirdaria},
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechHyleg, 0.95, ""}, {TechEclipses, 0.95, ""},
				{TechPrimaryDir, 0.9, ""}, {TechConvergence, 0.85, ""}, {TechProfections, 0.85, ""},
				{TechDecumbitura, 0.8, ""}, {TechStations, 0.8, ""}, {TechChiron, 0.8, ""},
				{TechFirdaria, 0.8, ""}, {TechFastTransits, 0.75, ""}, {TechSolarReturn, 0.75, ""},
				{TechTemperament, 0.7, ""}, {TechProgressions, 0.7, ""}, {TechMelothesia, 0.6, ""},
			},
			Keywords: []string{"crisis", "perdida", "duelo", "accidente", "emergencia personal"},
		},
		{
			ID: "deporte", Name: "Deporte y Rendimiento", Parent: "salud",
			TechniquesRequired: []string{TechFastTransits},
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechFastTransits, 0.9, ""}, {TechProfections, 0.85, ""},
				{TechSolarReturn, 0.8, ""}, {TechFirdaria, 0.8, ""}, {TechPrimaryDir, 0.75, ""},
				{TechLunarReturn, 0.7, ""}, {TechConvergence, 0.7, ""}, {TechChiron, 0.7, ""},
				{TechTemperament, 0.7, ""}, {TechEclipses, 0.7, ""},
			},
			Keywords: []string{"deporte", "competicion deportiva", "rendimiento fisico", "entrenamiento", "carrera deportiva"},
		},
		{
			ID: "cronico", Name: "Enfermedad Crónica", Parent: "salud",
			TechniquesRequired: []string{TechFirdaria, TechZR},
			TechniquesBrief: []TechniqueWeight{
				{TechFirdaria, 0.95, ""}, {TechZR, 0.9, ""}, {TechSolarArc, 0.9, ""},
				{TechHyleg, 0.9, ""}, {TechTriplicity, 0.85, ""}, {TechTransits, 0.85, ""},
				{TechProfections, 0.85, ""}, {TechPrimaryDir, 0.8, ""}, {TechDivisor, 0.8, ""},
				{TechChiron, 0.85, ""}, {TechDecumbitura, 0.8, ""}, {TechConvergence, 0.7, ""},
				{TechSolarReturn, 0.7, ""}, {TechEclipses, 0.7, ""}, {TechTemperament, 0.7, ""},
				{TechMelothesia, 0.6, ""},
			},
			Keywords: []string{"cronico", "enfermedad larga", "tratamiento largo", "cancer", "diabetes"},
		},
		{
			ID: "recuperacion", Name: "Recuperación", Parent: "salud",
			TechniquesRequired: []string{TechLunarReturn},
			TechniquesBrief: []TechniqueWeight{
				{TechLunarReturn, 0.9, ""}, {TechFastTransits, 0.85, ""}, {TechSolarReturn, 0.85, ""},
				{TechConvergence, 0.85, ""}, {TechSolarArc, 0.8, ""}, {TechProgressions, 0.8, ""},
				{TechLunations, 0.8, ""}, {TechHyleg, 0.8, ""}, {TechChiron, 0.85, ""},
				{TechProfections, 0.7, ""}, {TechTemperament, 0.7, ""}, {TechEclipses, 0.6, ""},
			},
			Keywords: []string{"recuperacion", "post operatorio", "rehabilitacion", "convalecencia"},
		},
		{
			ID: "emergencia", Name: "Emergencia Médica", Parent: "salud",
			TechniquesRequired: []string{TechStations, TechFastTransits},
			TechniquesBrief: []TechniqueWeight{
				{TechHyleg, 1.0, ""}, {TechSolarArc, 1.0, ""}, {TechPrimaryDir, 0.95, ""},
				{TechFastTransits, 0.95, ""}, {TechStations, 0.95, ""}, {TechConvergence, 0.95, ""},
				{TechEclipses, 0.9, ""}, {TechProfections, 0.85, ""}, {TechAntiscia, 0.8, ""},
				{TechDeclinations, 0.75, ""}, {TechDecumbitura, 0.8, ""}, {TechChiron, 0.7, ""},
				{TechTemperament, 0.7, ""}, {TechMelothesia, 0.6, ""},
			},
			Keywords: []string{"infarto", "acv", "emergencia medica", "urgencia", "internacion"},
		},

		// ══════════════════════════════════════════════════════════
		// SUBDOMAINS — Amor (3 new + 1 existing)
		// ══════════════════════════════════════════════════════════

		{
			ID: "amor_sinastria", Name: "Compatibilidad de Pareja", Parent: "amor",
			TechniquesRequired: []string{TechSynastry},
			TechniquesBrief: []TechniqueWeight{
				{TechSynastry, 1.0, ""}, {TechComposite, 0.8, ""},
				{TechTransits, 0.7, "casa_7,venus"}, {TechConvergence, 0.6, ""},
			},
			Keywords: []string{"compatibilidad", "nos llevamos", "somos compatibles"},
		},
		{
			ID: "pareja", Name: "Buscar Pareja", Parent: "amor",
			TechniquesRequired: []string{TechZR},
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechProgressions, 0.95, ""}, {TechZR, 0.85, ""},
				{TechVertex, 0.85, ""}, {TechProfections, 0.85, ""}, {TechNodes, 0.8, ""},
				{TechFirdaria, 0.8, ""}, {TechSolarReturn, 0.8, ""}, {TechLunarReturn, 0.8, ""},
				{TechEclipses, 0.75, ""}, {TechPrimaryDir, 0.7, ""}, {TechConvergence, 0.7, ""},
				{TechSynastry, 0.7, ""},
			},
			Precautions: []string{"Mercury Rx = no empezar relaciones nuevas"},
			Keywords:    []string{"buscar pareja", "encontrar el amor", "conocer a alguien", "soltero"},
		},
		{
			ID: "relacion", Name: "Estado de la Relación", Parent: "amor",
			TechniquesRequired: []string{TechSynastry},
			TechniquesBrief: []TechniqueWeight{
				{TechSynastry, 1.0, ""}, {TechSolarArc, 0.9, ""}, {TechComposite, 0.9, ""},
				{TechNodes, 0.85, ""}, {TechEclipses, 0.85, ""}, {TechProgressions, 0.85, ""},
				{TechProfections, 0.8, ""}, {TechLilith, 0.8, ""}, {TechFirdaria, 0.75, ""},
				{TechSolarReturn, 0.75, ""}, {TechPrimaryDir, 0.7, ""}, {TechLunarReturn, 0.7, ""},
				{TechConvergence, 0.75, ""},
			},
			Keywords: []string{"estado de la relacion", "crisis de pareja", "como va mi relacion", "pelea con mi pareja"},
		},
		{
			ID: "matrimonio", Name: "Matrimonio", Parent: "amor",
			TechniquesRequired: []string{TechSynastry, TechElectional},
			TechniquesBrief: []TechniqueWeight{
				{TechElectional, 1.0, ""}, {TechSynastry, 0.95, ""}, {TechLunations, 0.9, ""},
				{TechMercuryRx, 0.9, ""}, {TechSolarArc, 0.85, ""}, {TechLunarReturn, 0.85, ""},
				{TechConvergence, 0.8, ""}, {TechProfections, 0.8, ""}, {TechSolarReturn, 0.7, ""},
				{TechNodes, 0.7, ""}, {TechProgressions, 0.7, ""}, {TechFirdaria, 0.7, ""},
				{TechVertex, 0.7, ""},
			},
			Precautions: []string{"Mercury Rx = no casarse", "Luna VOC = ceremonia sin efecto"},
			Keywords:    []string{"casarme", "boda", "casamiento", "fecha para casarse", "matrimonio"},
		},

		// ══════════════════════════════════════════════════════════
		// SUBDOMAINS — Dinero (4 new)
		// ══════════════════════════════════════════════════════════

		{
			ID: "inmobiliario", Name: "Inmobiliario", Parent: "dinero",
			TechniquesRequired: []string{TechEclipses},
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechPrimaryDir, 0.9, ""}, {TechTimingWindows, 0.85, ""},
				{TechProfections, 0.85, ""}, {TechElectional, 0.8, ""}, {TechEclipses, 0.8, ""},
				{TechSolarReturn, 0.8, ""}, {TechFirdaria, 0.8, ""}, {TechConvergence, 0.7, ""},
				{TechMercuryRx, 0.7, ""}, {TechZR, 0.6, ""},
			},
			Precautions: []string{"Mercury Rx = no firmar escrituras"},
			Keywords:    []string{"inmobiliario", "propiedad", "departamento", "comprar casa", "vender casa", "alquiler", "escritura inmueble"},
		},
		{
			ID: "inversiones", Name: "Inversiones", Parent: "dinero",
			TechniquesRequired: []string{TechEclipses},
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechEclipses, 0.9, ""}, {TechPrimaryDir, 0.9, ""},
				{TechZR, 0.9, ""}, {TechMercuryRx, 0.9, ""}, {TechProfections, 0.85, ""},
				{TechFastTransits, 0.85, ""}, {TechLots, 0.85, ""}, {TechFirdaria, 0.8, ""},
				{TechSolarReturn, 0.8, ""}, {TechStations, 0.8, ""}, {TechConvergence, 0.8, ""},
			},
			Precautions: []string{"Mercury Rx = no invertir", "Estaciones = alta volatilidad"},
			Keywords:    []string{"inversion", "invertir", "acciones", "bolsa", "cripto", "oportunidad financiera"},
		},
		{
			ID: "deudas", Name: "Deudas Personales", Parent: "dinero",
			TechniquesRequired: []string{TechFastTransits},
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechPrimaryDir, 0.85, ""}, {TechProfections, 0.85, ""},
				{TechFastTransits, 0.8, ""}, {TechEclipses, 0.8, ""}, {TechFirdaria, 0.75, ""},
				{TechSolarReturn, 0.7, ""}, {TechConvergence, 0.7, ""}, {TechStations, 0.6, ""},
			},
			Keywords: []string{"deuda", "debo plata", "presion financiera", "salir de deudas", "pagar deuda"},
		},
		{
			ID: "emprendimiento", Name: "Emprendimiento", Parent: "dinero",
			TechniquesRequired: []string{TechMercuryRx, TechElectional},
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechMercuryRx, 0.9, ""}, {TechZR, 0.85, ""},
				{TechElectional, 0.85, ""}, {TechPrimaryDir, 0.85, ""}, {TechProfections, 0.85, ""},
				{TechLots, 0.8, ""}, {TechFirdaria, 0.8, ""}, {TechSolarReturn, 0.8, ""},
				{TechLunations, 0.8, ""}, {TechConvergence, 0.8, ""}, {TechEclipses, 0.75, ""},
			},
			Precautions: []string{"Mercury Rx = no arrancar", "Luna VOC = no empezar"},
			Keywords:    []string{"emprender", "negocio propio", "arrancar negocio", "startup", "independizarme"},
		},

		// ══════════════════════════════════════════════════════════
		// SUBDOMAINS — Familia (5 new)
		// ══════════════════════════════════════════════════════════

		{
			ID: "herencia", Name: "Herencia y Sucesión Familiar", Parent: "familia",
			TechniquesRequired: []string{TechPrimaryDir},
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechPrimaryDir, 0.9, ""}, {TechProfections, 0.85, ""},
				{TechFirdaria, 0.8, ""}, {TechSolarReturn, 0.8, ""}, {TechEclipses, 0.75, ""},
				{TechLots, 0.7, ""}, {TechConvergence, 0.7, ""}, {TechZR, 0.6, ""},
				{TechDivisor, 0.6, ""},
			},
			Precautions: []string{"Mercury Rx = no firmar testamentos"},
			Keywords:    []string{"herencia", "testamento", "sucesion familiar", "bienes"},
		},
		{
			ID: "fertilidad", Name: "Fertilidad y Embarazo", Parent: "familia",
			TechniquesRequired: []string{TechProgressions, TechLunarReturn},
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechProgressions, 0.95, ""}, {TechLunarReturn, 0.9, ""},
				{TechProfections, 0.85, ""}, {TechEclipses, 0.85, ""}, {TechFirdaria, 0.8, ""},
				{TechSolarReturn, 0.8, ""}, {TechPrimaryDir, 0.75, ""}, {TechFastTransits, 0.7, ""},
				{TechConvergence, 0.7, ""}, {TechLunations, 0.65, ""}, {TechZR, 0.6, ""},
			},
			Precautions: []string{"Mercury Rx = precaución con tratamientos"},
			Keywords:    []string{"fertilidad", "embarazo", "concepcion", "hijo", "quedar embarazada", "tratamiento fertilidad"},
		},
		{
			ID: "hijos", Name: "Vínculo con Hijos", Parent: "familia",
			TechniquesRequired: []string{TechProgressions},
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechNodes, 0.85, ""}, {TechProgressions, 0.85, ""},
				{TechProfections, 0.85, ""}, {TechSynastry, 0.8, ""}, {TechFirdaria, 0.8, ""},
				{TechSolarReturn, 0.8, ""}, {TechEclipses, 0.8, ""}, {TechPrimaryDir, 0.7, ""},
				{TechLunarReturn, 0.65, ""}, {TechConvergence, 0.65, ""},
			},
			Keywords: []string{"hijo", "hija", "mis hijos", "relacion con mi hijo", "colegio"},
		},
		{
			ID: "padres", Name: "Vínculo con Padres", Parent: "familia",
			TechniquesRequired: []string{TechNodes},
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechNodes, 0.9, ""}, {TechEclipses, 0.9, ""},
				{TechSynastry, 0.85, ""}, {TechProfections, 0.85, ""}, {TechPrimaryDir, 0.8, ""},
				{TechFirdaria, 0.8, ""}, {TechSolarReturn, 0.75, ""}, {TechProgressions, 0.75, ""},
				{TechConvergence, 0.7, ""},
			},
			Keywords: []string{"padre", "madre", "mis padres", "relacion con padres", "cuidar a mis padres"},
		},
		{
			ID: "divorcio", Name: "Divorcio y Separación", Parent: "familia",
			TechniquesRequired: []string{TechSynastry, TechPrimaryDir},
			TechniquesBrief: []TechniqueWeight{
				{TechSynastry, 1.0, ""}, {TechSolarArc, 1.0, ""}, {TechPrimaryDir, 0.9, ""},
				{TechEclipses, 0.9, ""}, {TechConvergence, 0.85, ""}, {TechProfections, 0.85, ""},
				{TechProgressions, 0.8, ""}, {TechSolarReturn, 0.8, ""}, {TechLilith, 0.75, ""},
				{TechFastTransits, 0.75, ""}, {TechFirdaria, 0.75, ""}, {TechStations, 0.7, ""},
				{TechNodes, 0.7, ""},
			},
			Precautions: []string{"Estaciones = emociones intensas"},
			Keywords:    []string{"separacion", "custodia", "separar", "dejar a mi pareja", "divorciarme"},
		},

		// ══════════════════════════════════════════════════════════
		// SUBDOMAINS — Competitivo (1 new)
		// ══════════════════════════════════════════════════════════

		{
			ID: "competencia", Name: "Competencia y Rivales", Parent: "competitivo",
			TechniquesBrief: []TechniqueWeight{
				{TechSolarArc, 1.0, ""}, {TechSynastry, 0.9, ""}, {TechPrimaryDir, 0.9, ""},
				{TechProfections, 0.85, ""}, {TechFirdaria, 0.8, ""}, {TechSolarReturn, 0.8, ""},
				{TechEclipses, 0.75, ""}, {TechFastTransits, 0.65, ""}, {TechZR, 0.6, ""},
				{TechConvergence, 0.7, ""},
			},
			Precautions: []string{"Mercury Rx = no iniciar conflictos"},
			Keywords:    []string{"competidor personal", "conflicto con alguien", "pelea con vecino", "enemistad"},
		},
	}
}
