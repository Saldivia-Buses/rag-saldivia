package intelligence

import (
	"fmt"
	"math"
	"time"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/technique"
)

// Alert represents an urgent astrological activation for a contact.
type Alert struct {
	ContactName string  `json:"contact_name"`
	ContactID   string  `json:"contact_id"`
	Type        string  `json:"type"`    // "sa_urgent", "sa_critical", "eclipse"
	Detail      string  `json:"detail"`
	Urgency     string  `json:"urgency"` // "urgent" (< 0.15°), "critical" (< 0.05°)
	Orb         float64 `json:"orb"`
}

// ContactForAlert is the minimal contact data needed for alert scanning.
type ContactForAlert struct {
	ID        string
	Name      string
	BirthYear int
	BirthMonth int
	BirthDay  int
	BirthHour float64
	Lat, Lon, Alt float64
	UTCOffset int
}

// ScanAlerts checks contacts for SA/DP/Eclipse activations that are exact
// or near-exact this period. Max maxContacts processed per scan.
// SA < 0.15° = urgent (~8 weeks), SA < 0.05° = critical (~1 week).
func ScanAlerts(contacts []ContactForAlert, year int, maxContacts int) []Alert {
	if maxContacts <= 0 {
		maxContacts = 10
	}
	if len(contacts) > maxContacts {
		contacts = contacts[:maxContacts]
	}

	var alerts []Alert
	now := time.Now()
	jdNow := ephemeris.JulDay(now.Year(), int(now.Month()), now.Day(), float64(now.Hour()))

	for _, c := range contacts {
		chart, err := natal.BuildNatal(c.BirthYear, c.BirthMonth, c.BirthDay,
			c.BirthHour, c.Lat, c.Lon, c.Alt, c.UTCOffset)
		if err != nil {
			continue
		}

		// Check Solar Arcs
		saResults := technique.FindSolarArcActivations(chart, jdNow)
		for _, sa := range saResults {
			orb := math.Abs(sa.Orb)
			if orb < 0.05 {
				alerts = append(alerts, Alert{
					ContactName: c.Name, ContactID: c.ID,
					Type: "sa_critical", Urgency: "critical", Orb: orb,
					Detail: fmt.Sprintf("SA %s %s %s (orbe %.3f°) — esta semana", sa.SAplanet, sa.Aspect, sa.NatPlanet, orb),
				})
			} else if orb < 0.15 {
				alerts = append(alerts, Alert{
					ContactName: c.Name, ContactID: c.ID,
					Type: "sa_urgent", Urgency: "urgent", Orb: orb,
					Detail: fmt.Sprintf("SA %s %s %s (orbe %.3f°) — próximas 8 semanas", sa.SAplanet, sa.Aspect, sa.NatPlanet, orb),
				})
			}
		}

		// Check eclipses near natal points
		eclActivations, err := technique.FindEclipseActivations(chart, year)
		if err == nil {
			for _, ea := range eclActivations {
				if ea.Orb < 2.0 {
					alerts = append(alerts, Alert{
						ContactName: c.Name, ContactID: c.ID,
						Type: "eclipse", Urgency: "urgent", Orb: ea.Orb,
						Detail: fmt.Sprintf("Eclipse %s sobre %s (orbe %.2f°)", ea.Eclipse.Type, ea.NatPoint, ea.Orb),
					})
				}
			}
		}
	}

	return alerts
}
