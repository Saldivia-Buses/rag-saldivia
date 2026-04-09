package inventory

// Differ detects changes between scan cycles.
// Used to generate change events and update documentation.

// Change represents a detected difference between scans.
type Change struct {
	DeviceID string `json:"device_id"`
	Field    string `json:"field"`    // "ip", "status", "hostname", "ports", etc.
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
}

// DiffPorts compares two port lists and returns changes.
func DiffPorts(old, new []PortDoc) []Change {
	oldMap := make(map[int]PortDoc)
	for _, p := range old {
		oldMap[p.Port] = p
	}

	var changes []Change
	for _, p := range new {
		if prev, ok := oldMap[p.Port]; !ok {
			changes = append(changes, Change{
				Field:    "port_added",
				NewValue: portKey(p),
			})
		} else if prev.State != p.State {
			changes = append(changes, Change{
				Field:    "port_state",
				OldValue: portKey(prev),
				NewValue: portKey(p),
			})
		}
		delete(oldMap, p.Port)
	}

	// Ports that disappeared
	for _, p := range oldMap {
		changes = append(changes, Change{
			Field:    "port_removed",
			OldValue: portKey(p),
		})
	}

	return changes
}

func portKey(p PortDoc) string {
	return p.Protocol + ":" + string(rune(p.Port)) + "/" + p.State
}
