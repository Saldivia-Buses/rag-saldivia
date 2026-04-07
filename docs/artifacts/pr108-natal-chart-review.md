# Gateway Review -- PR #108 Natal Chart (Astro Phase 4)

**Fecha:** 2026-04-06
**Resultado:** CAMBIOS REQUERIDOS

## Archivos revisados

- `services/astro/internal/natal/chart.go`
- `services/astro/internal/natal/chart_test.go`
- `services/astro/internal/ephemeris/sweph.go`
- `services/astro/internal/astromath/convert.go`
- `services/astro/internal/astromath/houses.go`
- `services/astro/internal/astromath/angles.go`
- `services/astro/testdata/golden/natal_adrian.json`

---

## Bloqueantes

### B1. Lilith y Nodo Norte/Sur no estan en el skip de combustion -- falsos positivos [MEDIA]

**Archivo:** `services/astro/internal/natal/chart.go:107`

```go
skip := map[string]bool{"Sol": true, "Fortuna": true, "Espíritu": true, "Nodo Norte": true, "Nodo Sur": true}
```

"Lilith" y "Quiron" no estan en skip. Lilith (Mean Apogee) es un punto matematico, no un cuerpo -- aplicarle combustion no tiene sentido astronomico. Quiron como asteroid body SI puede ser combusto, asi que esta bien.

**Fix:** Agregar `"Lilith": true` al skip map. Quiron dejarlo afuera.

---

## Debe corregirse

### D1. EclNut y CalcHouses se llaman ANTES de adquirir CalcMu [ALTA]

**Archivo:** `services/astro/internal/natal/chart.go:44-54`

```go
ephemeris.CalcMu.Lock()
defer ephemeris.CalcMu.Unlock()

ephemeris.SetTopo(lon, lat, alt)
epsilon, err := ephemeris.EclNut(jd)       // OK, inside lock
cusps, ascmc, err := ephemeris.CalcHouses(...)  // OK, inside lock
```

Despues de releer: el Lock esta en linea 44, SetTopo en 48, EclNut en 49, CalcHouses en 54. Todo esta dentro del lock. **No hay bug aqui.** El orden es correcto: Lock -> SetTopo -> EclNut -> CalcHouses -> CalcPlanetFullLocked -> Unlock.

*(Auto-corregido durante review -- no requiere accion.)*

### D2. CalcHouses usa CalcHouses (no CalcHousesEx con flags topocentric) [MEDIA]

**Archivo:** `services/astro/internal/natal/chart.go:54`

```go
cusps, ascmc, err := ephemeris.CalcHouses(jd, lat, lon, ephemeris.HouseTopocentric)
```

BuildNatal usa posiciones topocentricas para planetas (`FlagTopoctr`) pero `CalcHouses` llama a `swe.Houses` que NO acepta flags -- por lo que las cusps se calculan geocentricamente. Esto es inconsistente.

Existe `CalcHousesEx` que acepta flags y llama a `swe.HousesEx`. Para consistencia topocentric deberia usarse:

```go
cusps, ascmc, err := ephemeris.CalcHousesEx(jd, ephemeris.FlagTopoctr, lat, lon, ephemeris.HouseTopocentric)
```

**Nota:** En la practica la diferencia geocentric vs topocentric en cusps es sub-arcsecond para altitudes normales. Pero el principio de consistencia importa, y las tecnicas de primary directions van a depender de cusps precisas.

### D3. Golden test no verifica RA/Dec ni cusps [MEDIA]

**Archivo:** `services/astro/internal/natal/chart_test.go:82-92`

El golden file tiene RA, Dec, cusps, y ARMC. El test solo verifica `Lon`. Si una regresion rompe RA/Dec (que usa CalcPlanetFullLocked), el test no lo detecta. Dado que RA/Dec son criticos para primary directions, deberian compararse.

**Fix:** Agregar assertions para RA y Dec en el loop de planetas (misma tolerancia 0.01). Agregar verificacion de ARMC contra `chart.ARMC`.

### D4. South Node RA es solo +180, deberia ser recalculado [BAJA]

**Archivo:** `services/astro/internal/natal/chart.go:78-84`

```go
planets["Nodo Sur"] = &ephemeris.PlanetPos{
    Lon:   astromath.Normalize360(pos.Lon + 180),
    Lat:   -pos.Lat,
    Speed: pos.Speed,
    RA:    astromath.Normalize360(pos.RA + 180),
    Dec:   -pos.Dec,
}
```

Para la ecliptica (Lon) negar Lat y sumar 180 a Lon es correcto por definicion. Para equatorial, `RA + 180` y `-Dec` es la transformacion correcta para un punto antipodal en la ecliptica cuando Lat = 0 (que es el caso del nodo). Sin embargo, `Speed` deberia tener signo invertido: si NN retrograda (speed negativo), SN tambien retrograda con la misma magnitud. El golden file confirma:

- NN speed: `0.02195`
- SN speed: `-0.02195`

Pero el Go code copia `pos.Speed` sin negar, asi que SN.Speed sera positivo cuando deberia ser negativo. Esto afecta el `IsRetrograde` check: SN se reportara como directo cuando deberia ser retrogrado.

**Fix:** `Speed: -pos.Speed` (o dejarlo igual si se argumenta que SN siempre retrograda por convencion, pero el golden file dice negativo).

### D5. Map iteration order no determinista -- no-issue funcional pero afecta debugging [INFORMACIONAL]

**Archivo:** `services/astro/internal/natal/chart.go:62`

`mainPlanets` es un map, asi que el orden de CalcPlanetFullLocked es no determinista. Esto no causa bugs porque los resultados son independientes entre planetas, pero si un calculo falla, los logs no mostraran siempre el mismo orden. No requiere fix a menos que se quiera reproducibilidad exacta en logs.

---

## Sugerencias

### S1. Tolerance de 0.01 grados es adecuada para Moshier

El comentario del test es correcto: Moshier (que usa swephgo por default sin archivos .se1) difiere de Swiss Ephemeris pleno en ~0.002-0.01 grados para planetas clasicos, hasta 0.05 para Pluton en fechas extremas. 0.01 es un buen balance para 1975.

Para Pluton especificamente, si alguna vez la CI falla, considerar subir la tolerancia a 0.02 solo para Pluton.

### S2. Missing Quiron en golden file

La golden file no tiene "Quiron". Si Chiron falla (sin archivos asteroid ephemeris), se skipea silenciosamente. Esto esta bien por ahora, pero en CI deberia haber un test que verifica que Quiron se calcula al menos para el caso basico (con Moshier, que SI soporta Chiron).

### S3. "Espiritu" y "Fortuna" no validados contra Python

La golden file tiene "Fortuna" (lon 342.659) pero NO "Espiritu". El test solo chequea existencia. Deberia comparar `chart.Planets["Fortuna"].Lon` contra `golden.Output.Planets["Fortuna"].Lon` como hace con los demas planetas. Actualmente esto funciona por accidente: el loop en linea 82 SI itera Fortuna del golden file y lo compara contra Go.

### S4. Coarse lock es correcto para este caso

BuildNatal con un lock para toda la funcion es la estrategia correcta. SetTopo es global state en swephgo; intercalar dos BuildNatal con diferentes lat/lon sin lock seria un data race. El costo es ~10-15 CalcUt calls * ~50us = ~1ms total hold time, que es irrelevante.

Si en el futuro se necesita paralelismo (e.g., batch de 100 charts), considerar un worker pool de 1 o usar swephgo instances separadas.

---

## Lo que esta bien

- **Locked/Unlocked pattern:** `CalcPlanetFull` vs `CalcPlanetFullLocked` es un patron limpio. La funcion privada `calcPlanetFullLocked` evita duplicacion. Los docs son claros sobre cuando usar cada una.
- **No re-entrant deadlock:** Verificado -- `EclNut`, `CalcHouses`, `CalcPlanet` ninguno toca `CalcMu`. Solo `CalcPlanetFull` (el wrapper publico) lo adquiere, y `BuildNatal` usa `CalcPlanetFullLocked` correctamente.
- **Lilith geocentric flag:** Correctamente omite `FlagTopoctr` para Mean Apogee (linea 88). Lilith es un punto orbital, no topocentric.
- **Chiron graceful degradation:** Skip con log en vez de fail si no hay archivos de asteroides.
- **Golden file approach:** Comparar contra Python pyswisseph output es la forma correcta de validar un port.
- **Part of Fortune/Spirit formulas:** Correctas. Diurnal = ASC + Moon - Sun, Nocturnal invierte. Spirit es el inverso exacto.
- **South Node ecliptic coords:** Lon + 180, -Lat es correcto para un punto antipodal en la ecliptica con Lat = 0.
- **HouseForLon:** Maneja wrap-around correctamente (casa que cruza 0 Aries).

---

## Resumen de acciones

| ID | Severidad | Accion |
|----|-----------|--------|
| D2 | MEDIA | Usar `CalcHousesEx` con `FlagTopoctr` para cusps |
| D3 | MEDIA | Agregar RA/Dec/ARMC assertions al golden test |
| D4 | BAJA | Negar Speed en South Node: `Speed: -pos.Speed` |
| B1 | MEDIA | Agregar `"Lilith": true` al combustion skip map |
| S2 | BAJA | Agregar Quiron al golden file o test dedicado |
