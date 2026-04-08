package technique

// TimelineEvent represents a single astrological event on a timeline.
type TimelineEvent struct {
	Month     int     `json:"month"`
	Day       int     `json:"day"`
	Technique string  `json:"technique"`
	Planet    string  `json:"planet"`
	Target    string  `json:"target"`     // natal point affected
	Aspect    string  `json:"aspect"`
	Severity  float64 `json:"severity"`   // 0-1 importance
}

// BuildActivationTimeline creates a chronological list of all activations in a year.
// Merges transits, eclipses, stations, and SA/PD activations into one sorted list.
func BuildActivationTimeline(
	transits []TransitActivation,
	fastTransits []FastTransitActivation,
	eclipses []EclipseActivation,
	stations []Station,
	solarArcs []SolarArcResult,
	eclipseTriggers []EclipseTrigger,
) []TimelineEvent {
	var events []TimelineEvent

	// Slow transits
	for _, tr := range transits {
		for _, ep := range tr.EpDetails {
			events = append(events, TimelineEvent{
				Month: ep.MonthStart, Technique: "tránsito",
				Planet: tr.Transit, Target: tr.Natal, Aspect: tr.Aspect,
				Severity: 0.7,
			})
		}
	}

	// Fast transits
	for _, ft := range fastTransits {
		events = append(events, TimelineEvent{
			Month: ft.Month, Day: ft.Day, Technique: "tránsito rápido",
			Planet: ft.Transit, Target: ft.Natal, Aspect: ft.Aspect,
			Severity: 0.3,
		})
	}

	// Eclipses
	for _, ecl := range eclipses {
		events = append(events, TimelineEvent{
			Month: ecl.Eclipse.Month, Day: ecl.Eclipse.Day,
			Technique: "eclipse " + ecl.Eclipse.Type,
			Target: ecl.NatPoint, Aspect: ecl.Aspect,
			Severity: 0.9,
		})
	}

	// Stations
	for _, st := range stations {
		sev := 0.6
		if st.NatPoint != "" {
			sev = 0.85
		}
		events = append(events, TimelineEvent{
			Month: st.Month, Day: st.Day, Technique: "estación",
			Planet: st.Planet, Target: st.NatPoint,
			Aspect: st.Type, Severity: sev,
		})
	}

	// Eclipse triggers
	for _, et := range eclipseTriggers {
		events = append(events, TimelineEvent{
			Month: et.Month, Technique: "trigger eclipse",
			Planet: et.TriggerPlanet, Target: et.NatPoint,
			Aspect: et.Aspect, Severity: 0.75,
		})
	}

	// Sort by month, then day, then severity descending
	for i := 0; i < len(events); i++ {
		for j := i + 1; j < len(events); j++ {
			if events[j].Month < events[i].Month ||
				(events[j].Month == events[i].Month && events[j].Day < events[i].Day) ||
				(events[j].Month == events[i].Month && events[j].Day == events[i].Day && events[j].Severity > events[i].Severity) {
				events[i], events[j] = events[j], events[i]
			}
		}
	}

	return events
}
