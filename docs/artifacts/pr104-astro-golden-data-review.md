# Gateway Review -- PR #104 Astro Service Golden Test Data (Plan 11, Phase 0)

**Fecha:** 2026-04-06
**Resultado:** CAMBIOS REQUERIDOS

---

## Bloqueantes

Ninguno. Los datos son generados y el servicio Go aun no existe, por lo que no hay riesgo de runtime. Los problemas a continuacion deben corregirse antes de que Go consuma estos archivos, pero no bloquean el merge del PR de datos.

---

## Debe corregirse

### M-01 [MEDIUM] Transits: episodes y ep_details son tuples posicionales, no dicts

**Archivo:** `services/astro/testdata/golden/transits_adrian_2026.json`

Cada transit entry tiene `ep_details` como `[[3, 4, false], ...]` y `episodes` como `[[[2.574, 3.677, false, 2461116.5], ...]]`. Estos son tuples posicionales de Python sin nombres de campo.

En Go, esto se deserializaria como `[][]interface{}` o requeriria un unmarshaller custom. Go no puede saber que el primer float es "orb", el segundo "position", el tercero "retrograde", el cuarto "jd".

**Fix:** Convertir en la funcion `generate_transits()` cada sub-array a un dict con keys explicitas. Ejemplo:

```python
# ep_details: [month_start, month_end, retrograde]
# episodes: [orb, transit_pos, retrograde, jd]
```

Debe ser:

```json
{"month_start": 3, "month_end": 4, "retrograde": false}
{"orb": 2.574, "transit_lon": 3.677, "retrograde": false, "jd": 2461116.5}
```

### M-02 [MEDIUM] Firdaria: sequence es array de tuples, no dicts

**Archivo:** `services/astro/testdata/golden/firdaria_adrian_2026.json`, lineas 24-61

`sequence` es `[["Sol", 10], ["Venus", 8], ...]`. Mismo problema que M-01 -- tuples posicionales sin keys.

**Fix:** Convertir a `[{"lord": "Sol", "years": 10}, ...]`.

### M-03 [MEDIUM] Firdaria: Unicode emoji prefixes en nombres de planetas

**Archivo:** `services/astro/testdata/golden/firdaria_adrian_2026.json`, lineas 54, 58

Los nodos aparecen como `"U+2604 Nodo Norte"` y `"U+2605 Nodo Sur"` (con simbolos Unicode astrologicos). Sin embargo, en TODOS los demas archivos (natal, primary_dir, transits, profections) los mismos puntos aparecen como `"Nodo Norte"` y `"Nodo Sur"` sin prefijo.

Esto generara una mismatch en Go cuando se compare por nombre: `"Nodo Norte" != "U+2604 Nodo Norte"`.

**Fix:** Normalizar nombres en `generate_firdaria()` para eliminar los prefijos Unicode, o documentar explicitamente que Go debe hacer strip de prefijos Unicode en la sequence de firdaria.

### M-04 [MEDIUM] Solar return: todos los planetas tienen speed=0.0

**Archivo:** `services/astro/testdata/golden/solar_return_adrian_2026.json`

Cada planeta tiene `"speed": 0.0`. Para el Sol en la solar return es llamativo (deberia tener ~1.01 deg/day como en la natal). El Sol en la SR tiene `lon: 275.409` que coincide con la natal -- esto es correcto (misma posicion solar), pero la velocidad deberia estar presente.

Esto sugiere que `calculate_solar_return()` no devuelve speed en sus tuples, y `planet_tuple_to_dict()` recibio tuples de 4 elementos (linea 52 del generator: `len(t) >= 4` branch asigna `speed: 0.0`).

**Fix:** Verificar si `calculate_solar_return()` realmente no devuelve speed. Si lo devuelve, ajustar el tuple handling. Si no, documentar en el golden file que `speed=0.0` es un placeholder y Go tests no deben verificar speed en SR charts. Agregar un campo `"_note"` al output o un comentario en el generator.

### M-05 [MEDIUM] Solar arc: faltan puntos extra (Nodo Norte, Nodo Sur, Lilith, Fortuna, Vertice, Ecl.Prenatal)

**Archivo:** `services/astro/testdata/golden/solar_arc_adrian_2026.json`

Solo tiene los 10 planetas principales (Sol a Pluton). El natal chart tiene 17 puntos (incluyendo MC, AS, Nodo Norte, Nodo Sur, Lilith, Fortuna, Vertice, Ecl.Prenatal, Ecl.Prenatal.L). El generator itera `PLANET_IDS` (de primary_directions.py) que aparentemente solo contiene los 10 planetas clasicos.

Esto es potencialmente intencional (SA no se aplica a MC/AS/Fortuna/Vertice), pero Nodo Norte, Nodo Sur y Lilith SI suelen recibir arco solar.

**Fix:** Evaluar si Go necesitara SA para nodos y Lilith. Si si, agregar esos puntos al calculo. Si no, documentar la exclusion.

### M-06 [MEDIUM] Profections: month_lords y month_houses usan string keys ("1" a "12")

**Archivo:** `services/astro/testdata/golden/profections_adrian_2026.json`

Las keys del cascade son strings: `"1"`, `"2"`, etc. En Go, JSON decodifica esto como `map[string]string` o `map[string]int`. Esto funciona, pero es semanticamente incorrecto -- son numeros de mes, no strings.

Esto NO es bloqueante (Go maneja `map[string]` bien), pero sera mas limpio para los tests si se documentan como int-keyed o se ajusta el generator para emitir un array de 12 posiciones en vez de un dict con string keys.

---

## Sugerencias

### S-01 [LOW] Transits file: 268KB / 12932 lineas / 86 activaciones es grande pero aceptable

Para golden test data, 268KB no es excesivo -- Go's `os.ReadFile()` + `json.Unmarshal()` lo maneja sin problemas. Sin embargo, si los Go tests solo verifican un subset (top 5-10 activaciones), se podria agregar un golden file reducido `transits_adrian_2026_top10.json` con las primeras 10 activaciones para tests rapidos, manteniendo el full file para tests exhaustivos.

### S-02 [LOW] Generator: `sys.path.insert(0, os.path.expanduser("~/astro-v2"))`

La linea 16 hardcodea `~/astro-v2` como path al codebase Python. Esto funciona en la maquina de desarrollo pero falla en CI.

Dado que este script solo se corre una vez para generar los golden files (no en CI), esto es aceptable. Pero seria bueno agregar un `ASTRO_V2_PATH` env var override:

```python
astro_path = os.environ.get("ASTRO_V2_PATH", os.path.expanduser("~/astro-v2"))
sys.path.insert(0, astro_path)
```

### S-03 [LOW] Generator: error handling silencioso en funciones con try/except

Las funciones `generate_transits`, `generate_profections`, `generate_firdaria`, `generate_solar_return`, `generate_primary_dir` imprimen un WARNING pero continuan la ejecucion. Esto es correcto para robustez, pero significa que un golden file podria no generarse y pasar desapercibido.

**Sugerencia:** Al final del `__main__` block, verificar que los 7 archivos esperados existen y tienen contenido no vacio:

```python
expected = ["natal_adrian", "solar_arc_adrian_2026", "primary_dir_adrian_2026",
            "transits_adrian_2026", "profections_adrian_2026",
            "firdaria_adrian_2026", "solar_return_adrian_2026"]
for name in expected:
    path = os.path.join(GOLDEN_DIR, f"{name}.json")
    if not os.path.exists(path) or os.path.getsize(path) < 10:
        print(f"FAIL: {name}.json missing or empty!")
        sys.exit(1)
```

### S-04 [LOW] Natal chart: Ecl.Prenatal.L y Ecl.Prenatal tienen longitudes redondeadas

En `natal_adrian.json`, el eclipse prenatal tiene `"lon": 220.4949` y `"lon": 55.9195` -- estos parecen valores redondeados a 4 decimales comparados con las coordenadas RA/Dec que tienen 15+ decimales. Esto podria ser intencional (precision original del dato) pero vale verificar.

### S-05 [INFO] Security: no hay datos sensibles

Los datos son calculos astronomicos para un sujeto de prueba identificado solo por nombre y fecha de nacimiento. Esto es informacion publica/calculable. No hay passwords, tokens, IPs, ni PII sensible mas alla del nombre. OK para el repo.

### S-06 [INFO] Consistencia de input: todos los archivos usan Adrian 27/12/1975 Rosario

Verificado en todos los golden files:
- natal: `year=1975, month=12, day=27, hour=16, minute=14, lat=-32.9468, lon=-60.6393`
- solar_arc: natal_jd=2442774.301388889 (consistent con natal)
- primary_dir: age=50.4668 (consistent con birth 1975-12-27 -> ref 2026-06-15)
- transits: year=2026
- profections: birth_date="27/12/1975", year=2026
- firdaria: birth_date="27/12/1975", year=2026
- solar_return: natal input matches, year=2026

Todo consistente.

---

## Lo que esta bien

1. **Estructura input/output clara.** Cada golden file separa `input` y `output`, lo que facilita testear en Go: leer input, pasar a la funcion Go, comparar output. Pattern excelente para table-driven tests.

2. **`planet_tuple_to_dict()` es defensivo.** Maneja tuples de 5, 4, y valores escalares. El fallback `{"raw": str(t)}` asegura que nunca se pierde data silenciosamente.

3. **Solar arc incluye constantes.** El archivo contiene `naibod_rate: 0.985626`, lo que permite verificar que Go usa la misma constante.

4. **Primary directions: datos ricos y completos.** Incluye 55 activaciones con promissor, significator, aspect, arc, age_exact, orb, tipo (directa/conversa), sistema (polich-page), y applying. Excelente para validar precision numerica.

5. **Date serialization manejada.** El generator convierte `date` objects a string via `default=str` en `json.dump()` y tambien via explicit checks en transits/primary_dir. No hay `datetime` objects fugados.

6. **GOLDEN_DIR relativo al script.** Usa `os.path.dirname(__file__)`, no un path absoluto. Esto funciona desde cualquier directorio.

7. **Moshier fallback documentado.** El docstring y el warning en `__main__` dejan claro que sin ephemeris files se usa Moshier (~1 arcsecond accuracy). Los Go tests deben usar la misma precision.
