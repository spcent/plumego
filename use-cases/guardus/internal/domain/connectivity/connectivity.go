// Package connectivity is the configuration and logic to verify whether the
// guardus process can reach the internet.
package connectivity

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"guardus/internal/client"
)

var (
	ErrInvalidInterval  = errors.New("connectivity.checker.interval must be 5s or higher")
	ErrInvalidDNSTarget = errors.New("connectivity.checker.target must be suffixed with :53")
)

// Config is the configuration for the connectivity checker.
type Config struct {
	Checker *Checker `json:"checker,omitempty"`
}

func (c *Config) ValidateAndSetDefaults() error {
	if c.Checker != nil {
		if c.Checker.Interval == 0 {
			c.Checker.Interval = 60 * time.Second
		} else if c.Checker.Interval < 5*time.Second {
			return ErrInvalidInterval
		}
		if !strings.HasSuffix(c.Checker.Target, ":53") {
			return ErrInvalidDNSTarget
		}
	}
	return nil
}

// Checker probes a TCP DNS endpoint to detect connectivity loss.
type Checker struct {
	Target   string        `json:"target"`
	Interval time.Duration `json:"interval,omitempty"`

	isConnected bool
	lastCheck   time.Time
}

// UnmarshalJSON accepts Interval as either a duration string ("60s") or numeric
// nanoseconds, mirroring the YAML ergonomics this struct used to provide.
func (c *Checker) UnmarshalJSON(data []byte) error {
	type checkerAlias Checker
	aux := struct {
		Interval any `json:"interval,omitempty"`
		*checkerAlias
	}{checkerAlias: (*checkerAlias)(c)}
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
			return fmt.Errorf("connectivity checker interval: %w", err)
		}
		c.Interval = d
	case float64:
		c.Interval = time.Duration(v)
	default:
		return fmt.Errorf("connectivity checker interval: unsupported type %T", v)
	}
	return nil
}

func (c *Checker) Check() bool {
	connected, _ := client.CanCreateNetworkConnection("tcp", c.Target, "", &client.Config{Timeout: 5 * time.Second})
	return connected
}

func (c *Checker) IsConnected() bool {
	if now := time.Now(); now.After(c.lastCheck.Add(c.Interval)) {
		c.lastCheck, c.isConnected = now, c.Check()
	}
	return c.isConnected
}
