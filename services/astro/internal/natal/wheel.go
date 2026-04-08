package natal

import (
	"fmt"
	"html"
	"math"
	"strings"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
)

// SVG dimensions and radii.
const (
	wheelSize   = 800
	wheelCX     = 400
	wheelCY     = 400
	rOuter      = 360 // outer ring
	rZodOuter   = 340 // zodiac band outer
	rZodInner   = 280 // zodiac band inner
	rPlanetBand = 255 // planet glyph radius
	rHouseOuter = 240 // house cusp outer
	rInner      = 180 // inner circle (aspect zone)
	rCenter     = 40  // center medallion
	minSep      = 12  // min separation between glyphs (degrees)
)

// Sign glyphs (Unicode astro symbols).
var signGlyphs = [12]string{"♈", "♉", "♊", "♋", "♌", "♍", "♎", "♏", "♐", "♑", "♒", "♓"}

// Element colors: Fire, Earth, Air, Water (by sign_idx % 4).
var elemColors = [4][2]string{
	{"#fdf0ec", "#b83a2a"}, // Fire: bg, glyph
	{"#edf3e8", "#3a6628"}, // Earth
	{"#faf5e4", "#8a6818"}, // Air
	{"#e8edf6", "#2c4278"}, // Water
}

// Planet display glyphs.
var planetGlyphs = map[string]string{
	"Sol": "☉", "Luna": "☽", "Mercurio": "☿", "Venus": "♀", "Marte": "♂",
	"Júpiter": "♃", "Saturno": "♄", "Urano": "♅", "Neptuno": "♆", "Plutón": "♇",
	"Quirón": "⚷", "Nodo Norte": "☊", "Nodo Sur": "☋", "Lilith": "⚸",
	"Fortuna": "⊕", "Espíritu": "⊗",
}

// Planet colors.
var planetColors = map[string]string{
	"Sol": "#d4a017", "Luna": "#6080b0", "Mercurio": "#808040",
	"Venus": "#c06080", "Marte": "#c04030", "Júpiter": "#4060a0",
	"Saturno": "#606060", "Urano": "#2890a0", "Neptuno": "#4868a8",
	"Plutón": "#604040", "Quirón": "#a08040", "Nodo Norte": "#707070",
	"Nodo Sur": "#909090", "Lilith": "#505050",
}

// Aspect colors: angle → (color, strokeWidth).
var aspectStyle = map[float64][2]string{
	0:   {"#d47828", "1.4"}, // conjunction — amber
	60:  {"#3878c8", "0.8"}, // sextile — blue
	90:  {"#c43030", "1.2"}, // square — crimson
	120: {"#2a8858", "1.0"}, // trine — emerald
	180: {"#983028", "1.2"}, // opposition — garnet
}

// ang converts ecliptic longitude to SVG angle (ASC at 180°/left).
func ang(lon, ascLon float64) float64 {
	return math.Mod(-(lon-ascLon)+360+180, 360)
}

// xy converts polar to cartesian.
func xy(angleDeg, r float64) (float64, float64) {
	a := angleDeg * math.Pi / 180
	return float64(wheelCX) + r*math.Cos(a), float64(wheelCY) - r*math.Sin(a)
}

// RenderWheel generates an SVG natal wheel as a string.
// Returns SVG with fixed 800x800 dimensions.
func RenderWheel(chart *Chart, name string) string {
	var b strings.Builder
	ascLon := chart.ASC

	// SVG header
	b.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" font-family="'Segoe UI',system-ui,sans-serif">`, wheelSize, wheelSize, wheelSize, wheelSize))
	b.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="#fafafa"/>`, wheelSize, wheelSize))

	// Title
	if name != "" {
		b.WriteString(fmt.Sprintf(`<text x="%d" y="25" text-anchor="middle" font-size="14" fill="#404040" font-weight="600">%s</text>`, wheelCX, html.EscapeString(name)))
	}

	// Zodiac band — 12 colored sectors
	for i := 0; i < 12; i++ {
		a1 := ang(float64(i)*30, ascLon)
		a2 := ang(float64(i)*30+30, ascLon)
		ec := elemColors[i%4]
		b.WriteString(sector(a1, a2, rZodOuter, rZodInner, ec[0]))

		// Sign glyph at midpoint
		amid := ang(float64(i)*30+15, ascLon)
		gx, gy := xy(amid, float64(rZodOuter+rZodInner)/2)
		b.WriteString(fmt.Sprintf(`<text x="%.1f" y="%.1f" text-anchor="middle" dominant-baseline="central" font-size="18" fill="%s">%s</text>`,
			gx, gy, ec[1], signGlyphs[i]))

		// Sign divider line
		x1, y1 := xy(a1, float64(rZodOuter))
		x2, y2 := xy(a1, float64(rZodInner))
		b.WriteString(fmt.Sprintf(`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#999" stroke-width="0.5"/>`, x1, y1, x2, y2))
	}

	// Border circles
	b.WriteString(circle(float64(rZodOuter), "#666", "1"))
	b.WriteString(circle(float64(rZodInner), "#888", "0.7"))
	b.WriteString(circle(float64(rInner), "#aaa", "0.5"))
	b.WriteString(circle(float64(rCenter), "#ddd", "0.5"))

	// House cusps
	for i := 1; i <= 12; i++ {
		if len(chart.Cusps) <= i {
			continue
		}
		a := ang(chart.Cusps[i], ascLon)
		isAngular := i == 1 || i == 4 || i == 7 || i == 10
		outerR := float64(rZodInner)
		if isAngular {
			outerR = float64(rZodOuter)
		}
		x1, y1 := xy(a, outerR)
		x2, y2 := xy(a, float64(rInner))

		sw := "0.5"
		dash := ` stroke-dasharray="4,3"`
		color := "#aaa"
		if isAngular {
			sw = "1.2"
			dash = ""
			color = "#555"
		}
		b.WriteString(fmt.Sprintf(`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="%s"%s/>`,
			x1, y1, x2, y2, color, sw, dash))

		// House number
		nextCusp := chart.Cusps[i%12+1]
		if len(chart.Cusps) <= i%12+1 {
			nextCusp = chart.Cusps[i] + 30
		}
		midLon := (chart.Cusps[i] + nextCusp) / 2
		if math.Abs(chart.Cusps[i]-nextCusp) > 180 {
			midLon += 180
		}
		hx, hy := xy(ang(midLon, ascLon), float64(rInner+rHouseOuter)/2)
		b.WriteString(fmt.Sprintf(`<text x="%.1f" y="%.1f" text-anchor="middle" dominant-baseline="central" font-size="9" fill="#888">%d</text>`,
			hx, hy, i))
	}

	// Aspect lines (in inner circle)
	skip := map[string]bool{"Fortuna": true, "Espíritu": true, "Nodo Sur": true}
	type planetEntry struct {
		name string
		lon  float64
	}
	var plist []planetEntry
	for name, pos := range chart.Planets {
		if skip[name] {
			continue
		}
		plist = append(plist, planetEntry{name, pos.Lon})
	}

	seen := make(map[string]bool)
	for i, p1 := range plist {
		for _, p2 := range plist[i+1:] {
			for aspAngle, style := range aspectStyle {
				orb := astromath.AngDiff(p1.lon, p2.lon)
				diff := math.Abs(orb - aspAngle)
				maxOrb := 8.0
				if aspAngle == 60 || aspAngle == 120 {
					maxOrb = 6.0
				}
				if diff > maxOrb {
					continue
				}
				key := p1.name + p2.name + fmt.Sprintf("%.0f", aspAngle)
				if seen[key] {
					continue
				}
				seen[key] = true

				a1 := ang(p1.lon, ascLon)
				a2 := ang(p2.lon, ascLon)
				x1, y1 := xy(a1, float64(rInner)-5)
				x2, y2 := xy(a2, float64(rInner)-5)
				opacity := 0.3 + (1-diff/maxOrb)*0.5
				b.WriteString(fmt.Sprintf(`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="%s" opacity="%.2f" stroke-linecap="round"/>`,
					x1, y1, x2, y2, style[0], style[1], opacity))
			}
		}
	}

	// Planet glyphs
	type glyphItem struct {
		name  string
		lon   float64
		angle float64
	}
	var items []glyphItem
	for name, pos := range chart.Planets {
		if skip[name] {
			continue
		}
		items = append(items, glyphItem{name, pos.Lon, ang(pos.Lon, ascLon)})
	}
	// Sort by angle
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].angle < items[i].angle {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	// Separate overlapping glyphs
	for pass := 0; pass < 10; pass++ {
		for i := 0; i < len(items); i++ {
			j := (i + 1) % len(items)
			da := math.Mod(items[j].angle-items[i].angle+360, 360)
			if da < float64(minSep) && da > 0 {
				push := (float64(minSep) - da) / 2
				items[i].angle = math.Mod(items[i].angle-push+360, 360)
				items[j].angle = math.Mod(items[j].angle+push, 360)
			}
		}
	}

	for _, it := range items {
		glyph := planetGlyphs[it.name]
		if glyph == "" {
			glyph = it.name[:2]
		}
		col := planetColors[it.name]
		if col == "" {
			col = "#404040"
		}

		// Tick mark at exact position
		origA := ang(it.lon, ascLon)
		tx1, ty1 := xy(origA, float64(rZodInner))
		tx2, ty2 := xy(origA, float64(rZodInner)-8)
		b.WriteString(fmt.Sprintf(`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="1"/>`, tx1, ty1, tx2, ty2, col))

		// Glyph
		gx, gy := xy(it.angle, float64(rPlanetBand))
		b.WriteString(fmt.Sprintf(`<text x="%.1f" y="%.1f" text-anchor="middle" dominant-baseline="central" font-size="14" fill="%s" font-weight="600">%s</text>`,
			gx, gy, col, glyph))

		// Degree label
		deg := int(it.lon) % 30
		dx, dy := xy(it.angle, float64(rPlanetBand)-18)
		b.WriteString(fmt.Sprintf(`<text x="%.1f" y="%.1f" text-anchor="middle" dominant-baseline="central" font-size="8" fill="#888">%d°</text>`,
			dx, dy, deg))

		// Retrograde marker
		if chart.Retrograde[it.name] {
			rx, ry := xy(it.angle, float64(rPlanetBand)-28)
			b.WriteString(fmt.Sprintf(`<text x="%.1f" y="%.1f" text-anchor="middle" dominant-baseline="central" font-size="7" fill="#c02020" font-weight="700">Rx</text>`, rx, ry))
		}
	}

	// Angular labels
	angles := []struct {
		label string
		lon   float64
		r     float64
	}{
		{"ASC", chart.ASC, float64(rZodOuter) + 15},
		{"MC", chart.MC, float64(rZodOuter) + 15},
	}
	for _, a := range angles {
		ax, ay := xy(ang(a.lon, ascLon), a.r)
		b.WriteString(fmt.Sprintf(`<text x="%.1f" y="%.1f" text-anchor="middle" dominant-baseline="central" font-size="10" fill="#333" font-weight="700">%s</text>`,
			ax, ay, a.label))
	}

	b.WriteString(`</svg>`)
	return b.String()
}

// sector draws an SVG arc sector between two angles.
func sector(a1, a2, rOuter, rInner float64, fill string) string {
	// Convert to radians
	r1 := a1 * math.Pi / 180
	r2 := a2 * math.Pi / 180

	ox1, oy1 := float64(wheelCX)+rOuter*math.Cos(r1), float64(wheelCY)-rOuter*math.Sin(r1)
	ox2, oy2 := float64(wheelCX)+rOuter*math.Cos(r2), float64(wheelCY)-rOuter*math.Sin(r2)
	ix1, iy1 := float64(wheelCX)+rInner*math.Cos(r2), float64(wheelCY)-rInner*math.Sin(r2)
	ix2, iy2 := float64(wheelCX)+rInner*math.Cos(r1), float64(wheelCY)-rInner*math.Sin(r1)

	largeArc := 0
	sweep := 0 // counterclockwise in SVG space
	da := math.Mod(a2-a1+360, 360)
	if da > 180 {
		largeArc = 1
	}

	return fmt.Sprintf(`<path d="M%.1f,%.1f A%.0f,%.0f 0 %d,%d %.1f,%.1f L%.1f,%.1f A%.0f,%.0f 0 %d,%d %.1f,%.1f Z" fill="%s"/>`,
		ox1, oy1, rOuter, rOuter, largeArc, sweep, ox2, oy2,
		ix1, iy1, rInner, rInner, largeArc, 1-sweep, ix2, iy2,
		fill)
}

// circle draws an SVG circle.
func circle(r float64, stroke, sw string) string {
	return fmt.Sprintf(`<circle cx="%d" cy="%d" r="%.0f" fill="none" stroke="%s" stroke-width="%s"/>`,
		wheelCX, wheelCY, r, stroke, sw)
}
