package registry

import "time"

// DefaultExpiry is a default node expiry duration.
const DefaultExpiry = time.Second

// Settings are setting use by the Registry service.
type Settings struct {
	// ExpiryDuration is how long a node can stay alive before being expired.
	ExpiryDuration time.Duration
}
