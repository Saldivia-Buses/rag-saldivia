package astromath

// Bound represents an Egyptian bound (term) within a zodiac sign.
type Bound struct {
	Lord    string
	FromDeg float64
	ToDeg   float64
}

// EgyptianBounds contains the Egyptian/Ptolemaic bounds for each sign.
// Index 0 = Aries, 11 = Pisces. Each sign has 5 bounds totaling 30°.
var EgyptianBounds = [12][]Bound{
	// Aries
	{{"Júpiter", 0, 6}, {"Venus", 6, 12}, {"Mercurio", 12, 20}, {"Marte", 20, 25}, {"Saturno", 25, 30}},
	// Taurus
	{{"Venus", 0, 8}, {"Mercurio", 8, 14}, {"Júpiter", 14, 22}, {"Saturno", 22, 27}, {"Marte", 27, 30}},
	// Gemini
	{{"Mercurio", 0, 6}, {"Júpiter", 6, 12}, {"Venus", 12, 17}, {"Marte", 17, 24}, {"Saturno", 24, 30}},
	// Cancer
	{{"Marte", 0, 7}, {"Venus", 7, 13}, {"Mercurio", 13, 19}, {"Júpiter", 19, 26}, {"Saturno", 26, 30}},
	// Leo
	{{"Júpiter", 0, 6}, {"Venus", 6, 11}, {"Saturno", 11, 18}, {"Mercurio", 18, 24}, {"Marte", 24, 30}},
	// Virgo
	{{"Mercurio", 0, 7}, {"Venus", 7, 17}, {"Júpiter", 17, 21}, {"Marte", 21, 28}, {"Saturno", 28, 30}},
	// Libra
	{{"Saturno", 0, 6}, {"Mercurio", 6, 14}, {"Júpiter", 14, 21}, {"Venus", 21, 28}, {"Marte", 28, 30}},
	// Scorpio
	{{"Marte", 0, 7}, {"Venus", 7, 11}, {"Mercurio", 11, 19}, {"Júpiter", 19, 24}, {"Saturno", 24, 30}},
	// Sagittarius
	{{"Júpiter", 0, 12}, {"Venus", 12, 17}, {"Mercurio", 17, 21}, {"Saturno", 21, 26}, {"Marte", 26, 30}},
	// Capricorn
	{{"Mercurio", 0, 7}, {"Júpiter", 7, 14}, {"Venus", 14, 22}, {"Saturno", 22, 26}, {"Marte", 26, 30}},
	// Aquarius
	{{"Mercurio", 0, 7}, {"Venus", 7, 13}, {"Júpiter", 13, 20}, {"Marte", 20, 25}, {"Saturno", 25, 30}},
	// Pisces
	{{"Venus", 0, 12}, {"Júpiter", 12, 16}, {"Mercurio", 16, 19}, {"Marte", 19, 28}, {"Saturno", 28, 30}},
}

// BoundLord returns the bound lord for a given ecliptic longitude.
func BoundLord(lon float64) string {
	lon = Normalize360(lon)
	sign := int(lon / 30)
	degInSign := lon - float64(sign*30)
	for _, b := range EgyptianBounds[sign] {
		if degInSign >= b.FromDeg && degInSign < b.ToDeg {
			return b.Lord
		}
	}
	return EgyptianBounds[sign][4].Lord
}

// ZRSignYears returns the minor years for Zodiacal Releasing per sign.
var ZRSignYears = [12]float64{
	15, 8, 20, 25, 19, 20, 8, 15, 12, 27, 30, 12,
}
