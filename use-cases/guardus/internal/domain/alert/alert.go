package alert

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrAlertWithInvalidDescription is the error with which Gatus will panic if an alert has an invalid character
	ErrAlertWithInvalidDescription = errors.New("alert description must not have \" or \\")

	ErrAlertWithInvalidMinimumReminderInterval = errors.New("minimum-reminder-interval must be either omitted or be at least 5m")
)

// Alert is endpoint.Endpoint's alert configuration
type Alert struct {
	// Type of alert (required)
	Type Type `json:"type"`

	// Enabled defines whether the alert is enabled. Pointer so callers can
	// distinguish "unset" from "explicitly false" in ParseWithDefaultAlert.
	// Use Alert.IsEnabled() to read the effective value.
	Enabled *bool `json:"enabled,omitempty"`

	// FailureThreshold is the number of failures in a row needed before triggering the alert
	FailureThreshold int `json:"failure-threshold,omitempty"`

	// SuccessThreshold defines how many successful executions must happen in a row before an ongoing incident is marked as resolved
	SuccessThreshold int `json:"success-threshold,omitempty"`

	// MinimumReminderInterval is the interval between reminders
	MinimumReminderInterval time.Duration `json:"minimum-reminder-interval,omitempty"`

	// Description of the alert. Pointer for the same unset/empty distinction as Enabled.
	Description *string `json:"description,omitempty"`

	// SendOnResolved defines whether to send a second notification when the issue has been resolved.
	// Pointer for the same unset/empty distinction as Enabled.
	SendOnResolved *bool `json:"send-on-resolved,omitempty"`

	// ProviderOverride is an optional field that can be used to override the provider's configuration
	// It is freeform so that it can be used for any provider-specific configuration.
	ProviderOverride map[string]any `json:"provider-override,omitempty"`

	// ResolveKey is an optional field that is used by some providers (i.e. PagerDuty's dedup_key) to resolve
	// ongoing/triggered incidents
	ResolveKey string `json:"-"`

	// Triggered is used to determine whether an alert has been triggered. When an alert is resolved, this value
	// should be set back to false. It is used to prevent the same alert from going out twice.
	Triggered bool `json:"-"`
}

// ValidateAndSetDefaults validates the alert's configuration and sets the default value of fields that have one
func (alert *Alert) ValidateAndSetDefaults() error {
	if alert.FailureThreshold <= 0 {
		alert.FailureThreshold = 3
	}
	if alert.SuccessThreshold <= 0 {
		alert.SuccessThreshold = 2
	}
	if alert.MinimumReminderInterval != 0 && alert.MinimumReminderInterval < 5*time.Minute {
		return ErrAlertWithInvalidMinimumReminderInterval
	}
	if strings.ContainsAny(alert.GetDescription(), "\"\\") {
		return ErrAlertWithInvalidDescription
	}
	return nil
}

// GetDescription retrieves the description of the alert
func (alert *Alert) GetDescription() string {
	if alert.Description == nil {
		return ""
	}
	return *alert.Description
}

// IsEnabled returns whether an alert is enabled or not
// Returns true if not set
func (alert *Alert) IsEnabled() bool {
	if alert.Enabled == nil {
		return true
	}
	return *alert.Enabled
}

// IsSendingOnResolved returns whether an alert is sending on resolve or not
func (alert *Alert) IsSendingOnResolved() bool {
	if alert.SendOnResolved == nil {
		return false
	}
	return *alert.SendOnResolved
}

// Checksum returns a checksum of the alert
// Used to determine which persisted triggered alert should be deleted on application start
func (alert *Alert) Checksum() string {
	hash := sha256.New()
	hash.Write([]byte(string(alert.Type) + "_" +
		strconv.FormatBool(alert.IsEnabled()) + "_" +
		strconv.FormatBool(alert.IsSendingOnResolved()) + "_" +
		strconv.Itoa(alert.SuccessThreshold) + "_" +
		strconv.Itoa(alert.FailureThreshold) + "_" +
		alert.GetDescription()),
	)
	return hex.EncodeToString(hash.Sum(nil))
}

func (alert *Alert) ProviderOverrideAsBytes() []byte {
	b, _ := json.Marshal(alert.ProviderOverride)
	return b
}
