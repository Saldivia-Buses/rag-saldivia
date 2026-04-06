package astromath

import "math"

// Aspect represents a detected aspect between two points.
type Aspect struct {
	Name     string
	Angle    float64
	Orb      float64
	Exact    bool // within 0.1° of exact
	Applying bool // closing orb
}

// FindAspect checks if two longitudes form a major aspect within the given orb.
func FindAspect(lon1, lon2, maxOrb float64) *Aspect {
	diff := AngDiff(lon1, lon2)
	for name, angle := range AspectAngles {
		orb := math.Abs(diff - angle)
		if orb <= maxOrb {
			return &Aspect{
				Name:  name,
				Angle: angle,
				Orb:   orb,
				Exact: orb < 0.1,
			}
		}
	}
	return nil
}

// FindAspectWithMotion checks aspect and determines if applying/separating.
// Uses shortest-arc signed difference to correctly handle all configurations.
func FindAspectWithMotion(lon1, lon2, speed1, speed2, maxOrb float64) *Aspect {
	asp := FindAspect(lon1, lon2, maxOrb)
	if asp == nil {
		return nil
	}
	// Signed shortest-arc difference: positive = lon1 ahead of lon2
	signed := SignedDiff(lon2, lon1)
	currentSep := math.Abs(signed)
	relSpeed := speed1 - speed2
	// Applying: the separation is decreasing toward the exact aspect angle
	if signed >= 0 {
		asp.Applying = relSpeed < 0 || currentSep > asp.Angle
	} else {
		asp.Applying = relSpeed > 0 || currentSep > asp.Angle
	}
	return asp
}
