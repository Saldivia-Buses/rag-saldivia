package technique

// TimingWindow represents an optimal date range based on technique convergence.
type TimingWindow struct {
	Month       int      `json:"month"`
	Score       int      `json:"score"`       // convergence score (higher = more active)
	Techniques  []string `json:"techniques"`  // which techniques converge here
	Nature      string   `json:"nature"`      // "favorable", "challenging", "mixed"
	Description string   `json:"description"` // human-readable summary
}

// CalcTimingWindows analyzes the convergence matrix and identifies
// high-activity and favorable windows.
// This is a post-processing step that requires other techniques to be computed first.
func CalcTimingWindows(
	transits []TransitActivation,
	eclipses []EclipseActivation,
	stations []Station,
	directions []PrimaryDirection,
) []TimingWindow {
	// Score each month
	scores := make([]TimingWindow, 12)
	for i := range scores {
		scores[i].Month = i + 1
	}

	// Transit episodes
	for _, tr := range transits {
		for _, ep := range tr.EpDetails {
			for m := ep.MonthStart; m <= ep.MonthEnd; m++ {
				if m >= 1 && m <= 12 {
					scores[m-1].Score += 2
					if m == ep.MonthStart {
						scores[m-1].Techniques = append(scores[m-1].Techniques,
							"TR_"+tr.Transit+"_"+tr.Aspect+"_"+tr.Natal)
					}
				}
			}
		}
	}

	// Eclipse activations
	for _, e := range eclipses {
		m := e.Eclipse.Month
		if m >= 1 && m <= 12 {
			scores[m-1].Score += 3
			scores[m-1].Techniques = append(scores[m-1].Techniques,
				"ECL_"+e.Eclipse.Type+"_"+e.NatPoint)
		}
	}

	// Stations near natal points
	for _, st := range stations {
		if st.NatPoint != "" && st.Month >= 1 && st.Month <= 12 {
			scores[st.Month-1].Score += 4
			scores[st.Month-1].Techniques = append(scores[st.Month-1].Techniques,
				"STAT_"+st.Planet+"_"+st.NatPoint)
		}
	}

	// Primary directions (approximate month from age)
	for _, d := range directions {
		if d.OrbDeg < 1.0 { // only tight directions
			// PDs don't have a month — score the birthday month as approximation
			// This is a simplified approach; a full implementation would convert age to calendar month
			scores[0].Score += 3
			scores[0].Techniques = append(scores[0].Techniques,
				"PD_"+d.Promissor+"_"+d.Aspect+"_"+d.Significator)
		}
	}

	// Classify nature based on technique mix
	for i := range scores {
		favorable := 0
		challenging := 0
		for _, t := range scores[i].Techniques {
			if len(t) > 3 {
				// Rough heuristic: eclipses and stations are challenging, trines favorable
				if t[:3] == "ECL" || t[:4] == "STAT" {
					challenging++
				} else {
					favorable++
				}
			}
		}
		switch {
		case favorable > challenging:
			scores[i].Nature = "favorable"
		case challenging > favorable:
			scores[i].Nature = "challenging"
		case scores[i].Score > 0:
			scores[i].Nature = "mixed"
		default:
			scores[i].Nature = "neutral"
		}
	}

	return scores
}
