package heartbeat

import (
	"encoding/json"
	"fmt"
	"time"
)

// Config used to check if the external endpoint has received new results when it should have.
// This configuration is used to trigger alerts when an external endpoint has no new results for a defined period of time
type Config struct {
	// Interval is the time interval at which Gatus verifies whether the external endpoint has received new results
	// If no new result is received within the interval, the endpoint is marked as failed and alerts are triggered
	Interval time.Duration `json:"interval"`
}

// UnmarshalJSON accepts Interval as either a duration string ("60s") or a
// numeric nanosecond value, mirroring the YAML ergonomics this struct used to provide.
func (c *Config) UnmarshalJSON(data []byte) error {
	var aux struct {
		Interval any `json:"interval"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	switch v := aux.Interval.(type) {
	case nil:
		c.Interval = 0
	case string:
		if v == "" {
			c.Interval = 0
			return nil
		}
		d, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("heartbeat interval: %w", err)
		}
		c.Interval = d
	case float64:
		c.Interval = time.Duration(v)
	default:
		return fmt.Errorf("heartbeat interval: unsupported type %T", v)
	}
	return nil
}
