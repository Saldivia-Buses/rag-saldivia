package natal

import (
	"fmt"
	"math"
	"strings"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
)

// Aspect grid dimensions.
const (
	gridCellSize = 36
	gridFontSize = 12
	gridPadding  = 40
)

// Aspect glyphs for the grid.
var aspectGridGlyphs = map[string]string{
	"conjunction": "☌", "sextile": "⚹", "square": "□",
	"trine": "△", "opposition": "☍",
}

var aspectGridColors = map[string]string{
	"conjunction": "#d47828", "sextile": "#3878c8", "square": "#c43030",
	"trine": "#2a8858", "opposition": "#983028",
}

// gridPlanets in display order.
var gridPlanets = []string{
	"Sol", "Luna", "Mercurio", "Venus", "Marte",
	"Júpiter", "Saturno", "Urano", "Neptuno", "Plutón",
}

// RenderAspectGrid generates an SVG aspect grid for a single chart.
func RenderAspectGrid(chart *Chart) string {
	// Collect planets that exist in chart
	var planets []string
	var lons []float64
	for _, name := range gridPlanets {
		if p, ok := chart.Planets[name]; ok {
			planets = append(planets, name)
			lons = append(lons, p.Lon)
		}
	}

	n := len(planets)
	if n < 2 {
		return ""
	}

	size := gridPadding + n*gridCellSize
	var b strings.Builder

	b.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" font-family="'Segoe UI',system-ui,sans-serif">`,
		size, size, size, size))
	b.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="#fafafa"/>`, size, size))

	// Planet labels along diagonal
	for i, name := range planets {
		glyph := planetGlyphs[name]
		if glyph == "" {
			glyph = name[:2]
		}
		col := planetColors[name]
		if col == "" {
			col = "#404040"
		}
		x := gridPadding + i*gridCellSize + gridCellSize/2
		y := gridPadding + i*gridCellSize + gridCellSize/2
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" dominant-baseline="central" font-size="%d" fill="%s" font-weight="600">%s</text>`,
			x, y, gridFontSize+2, col, glyph))
	}

	// Grid cells — lower triangle only
	for i := 1; i < n; i++ {
		for j := 0; j < i; j++ {
			asp := astromath.FindAspect(lons[i], lons[j], 8.0)
			if asp == nil {
				continue
			}

			x := gridPadding + j*gridCellSize + gridCellSize/2
			y := gridPadding + i*gridCellSize + gridCellSize/2

			glyph := aspectGridGlyphs[asp.Name]
			color := aspectGridColors[asp.Name]
			if glyph == "" || color == "" {
				continue
			}

			// Cell background
			opacity := 0.15 + (1-asp.Orb/8)*0.15
			b.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="%s" opacity="%.2f" rx="3"/>`,
				gridPadding+j*gridCellSize+2, gridPadding+i*gridCellSize+2,
				gridCellSize-4, gridCellSize-4, color, opacity))

			// Aspect glyph
			b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" dominant-baseline="central" font-size="%d" fill="%s">%s</text>`,
				x, y-5, gridFontSize+4, color, glyph))

			// Orb label
			orbStr := fmt.Sprintf("%.1f°", math.Round(asp.Orb*10)/10)
			b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" dominant-baseline="central" font-size="7" fill="#888">%s</text>`,
				x, y+8, orbStr))
		}
	}

	// Grid lines
	for i := 0; i <= n; i++ {
		y := gridPadding + i*gridCellSize
		x := gridPadding + i*gridCellSize
		b.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#ddd" stroke-width="0.5"/>`,
			gridPadding, y, size, y))
		b.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#ddd" stroke-width="0.5"/>`,
			x, gridPadding, x, size))
	}

	b.WriteString(`</svg>`)
	return b.String()
}

// RenderSynastryGrid generates an SVG aspect grid between two charts.
// Chart A planets on rows, Chart B planets on columns.
func RenderSynastryGrid(chartA, chartB *Chart, nameA, nameB string) string {
	var planetsA, planetsB []string
	var lonsA, lonsB []float64

	for _, name := range gridPlanets {
		if p, ok := chartA.Planets[name]; ok {
			planetsA = append(planetsA, name)
			lonsA = append(lonsA, p.Lon)
		}
	}
	for _, name := range gridPlanets {
		if p, ok := chartB.Planets[name]; ok {
			planetsB = append(planetsB, name)
			lonsB = append(lonsB, p.Lon)
		}
	}

	nA := len(planetsA)
	nB := len(planetsB)
	if nA < 1 || nB < 1 {
		return ""
	}

	width := gridPadding + nB*gridCellSize
	height := gridPadding + nA*gridCellSize

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" font-family="'Segoe UI',system-ui,sans-serif">`,
		width, height, width, height))
	b.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="#fafafa"/>`, width, height))

	// Column labels (Chart B)
	for j, name := range planetsB {
		glyph := planetGlyphs[name]
		if glyph == "" {
			glyph = name[:2]
		}
		x := gridPadding + j*gridCellSize + gridCellSize/2
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" font-size="%d" fill="#404040" font-weight="600">%s</text>`,
			x, gridPadding-8, gridFontSize, glyph))
	}

	// Row labels (Chart A)
	for i, name := range planetsA {
		glyph := planetGlyphs[name]
		if glyph == "" {
			glyph = name[:2]
		}
		y := gridPadding + i*gridCellSize + gridCellSize/2
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="end" dominant-baseline="central" font-size="%d" fill="#404040" font-weight="600">%s</text>`,
			gridPadding-8, y, gridFontSize, glyph))
	}

	// Aspect cells — full matrix (not triangular)
	for i := 0; i < nA; i++ {
		for j := 0; j < nB; j++ {
			asp := astromath.FindAspect(lonsA[i], lonsB[j], 8.0)
			if asp == nil {
				continue
			}

			x := gridPadding + j*gridCellSize + gridCellSize/2
			y := gridPadding + i*gridCellSize + gridCellSize/2

			glyph := aspectGridGlyphs[asp.Name]
			color := aspectGridColors[asp.Name]
			if glyph == "" || color == "" {
				continue
			}

			opacity := 0.15 + (1-asp.Orb/8)*0.15
			b.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="%s" opacity="%.2f" rx="3"/>`,
				gridPadding+j*gridCellSize+2, gridPadding+i*gridCellSize+2,
				gridCellSize-4, gridCellSize-4, color, opacity))

			b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" dominant-baseline="central" font-size="%d" fill="%s">%s</text>`,
				x, y-5, gridFontSize+4, color, glyph))

			orbStr := fmt.Sprintf("%.1f°", math.Round(asp.Orb*10)/10)
			b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" dominant-baseline="central" font-size="7" fill="#888">%s</text>`,
				x, y+8, orbStr))
		}
	}

	// Grid lines
	for i := 0; i <= nA; i++ {
		y := gridPadding + i*gridCellSize
		b.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#ddd" stroke-width="0.5"/>`,
			gridPadding, y, width, y))
	}
	for j := 0; j <= nB; j++ {
		x := gridPadding + j*gridCellSize
		b.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#ddd" stroke-width="0.5"/>`,
			x, gridPadding, x, height))
	}

	// Name labels
	if nameA != "" {
		b.WriteString(fmt.Sprintf(`<text x="5" y="%d" font-size="9" fill="#666" transform="rotate(-90,5,%d)">%s</text>`,
			height/2, height/2, nameA))
	}
	if nameB != "" {
		b.WriteString(fmt.Sprintf(`<text x="%d" y="12" text-anchor="middle" font-size="9" fill="#666">%s</text>`,
			gridPadding+nB*gridCellSize/2, nameB))
	}

	b.WriteString(`</svg>`)
	return b.String()
}
