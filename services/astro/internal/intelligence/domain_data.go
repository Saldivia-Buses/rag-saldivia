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
		// SUBDOMAINS
		// ══════════════════════════════════════════════════════════

		// Empresa subdomains
		{
			ID: "empresa_ventas", Name: "Ventas y Comercio", Parent: "empresa",
			TechniquesBrief: []TechniqueWeight{
				{TechTransits, 1.0, "casa_2,casa_7,mercurio"}, {TechTimingWindows, 1.0, ""},
				{TechFastTransits, 0.7, "mercurio,venus"}, {TechConvergence, 1.0, ""},
			},
			Keywords: []string{"venta", "cliente", "cotizacion", "propuesta", "licitacion"},
		},
		{
			ID: "empresa_rrhh", Name: "RRHH y Equipo", Parent: "empresa",
			TechniquesBrief: []TechniqueWeight{
				{TechSynastry, 1.0, ""}, {TechTransits, 0.8, "casa_6,casa_11"},
				{TechTimingWindows, 0.8, ""}, {TechConvergence, 0.9, ""},
			},
			Keywords: []string{"empleado", "contratar", "despedir", "equipo", "rrhh", "personal"},
		},
		{
			ID: "empresa_finanzas", Name: "Finanzas Empresariales", Parent: "empresa",
			TechniquesBrief: []TechniqueWeight{
				{TechTransits, 1.0, "casa_2,casa_8,jupiter,saturno"},
				{TechPlanetCycles, 0.9, "jupiter,saturno"}, {TechLots, 0.8, "prosperidad"},
				{TechConvergence, 1.0, ""},
			},
			Keywords: []string{"flujo de caja", "cash flow", "cobranza", "deuda empresarial", "inversion empresarial"},
		},

		// Carrera subdomains
		{
			ID: "carrera_vocacion", Name: "Vocación", Parent: "carrera",
			TechniquesBrief: []TechniqueWeight{
				{TechAlmuten, 1.0, ""}, {TechDisposition, 0.9, ""},
				{TechSect, 0.8, ""}, {TechProfections, 0.7, ""},
			},
			Keywords: []string{"vocacion", "que estudiar", "para que soy bueno", "talento"},
		},

		// Salud subdomains
		{
			ID: "salud_cirugia", Name: "Cirugía", Parent: "salud",
			Precautions: []string{"Evitar cirugía con Luna en signo del órgano afectado", "Mercury Rx = revisar diagnóstico"},
			Keywords:    []string{"cirugia", "operacion", "operar"},
		},
		{
			ID: "salud_mental", Name: "Salud Mental", Parent: "salud",
			TechniquesBrief: []TechniqueWeight{
				{TechTransits, 1.0, "luna,neptuno,casa_12"}, {TechLunations, 0.8, ""},
				{TechTemperament, 0.7, ""}, {TechConvergence, 0.9, ""},
			},
			Precautions: []string{"Siempre recomendar terapeuta profesional"},
			Keywords:    []string{"ansiedad", "depresion", "estres", "mental", "emocional", "psicologico"},
		},

		// Amor subdomains
		{
			ID: "amor_sinastria", Name: "Compatibilidad de Pareja", Parent: "amor",
			TechniquesRequired: []string{TechSynastry},
			TechniquesBrief: []TechniqueWeight{
				{TechSynastry, 1.0, ""}, {TechComposite, 0.8, ""},
				{TechTransits, 0.7, "casa_7,venus"}, {TechConvergence, 0.6, ""},
			},
			Keywords: []string{"compatibilidad", "sinastria", "nos llevamos", "somos compatibles"},
		},
	}
}
