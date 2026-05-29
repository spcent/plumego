// Package provider defines the AlertProvider contract that every alerting
// backend implements.
//
// Unlike upstream gatus this package does not use reflection to discover
// providers, and unlike a typical registry pattern it does not rely on
// init() side effects either. Dispatch is a typed-field switch on
// alerting.Config (see internal/alerting/config.go): each supported provider
// is a named struct field, and GetAlertingProviderByAlertType resolves by
// alert.Type. Adding a provider means adding a field and a switch case —
// nothing implicit, no hidden globals.
package provider

import (
	"fmt"

	"guardus/internal/domain/alert"
	"guardus/internal/domain/endpoint"
)

// AlertProvider is the contract every alerting backend implements.
type AlertProvider interface {
	// Validate reports whether the configured provider is valid.
	Validate() error
	// Send dispatches a notification. resolved indicates whether the alert
	// transitioned from triggered → resolved.
	Send(ep *endpoint.Endpoint, a *alert.Alert, r *endpoint.Result, resolved bool) error
	// GetDefaultAlert returns a template alert that endpoint-level alerts
	// merge into. May return nil when no default is configured.
	GetDefaultAlert() *alert.Alert
	// GetAlertType returns the alert.Type this provider handles.
	GetAlertType() alert.Type
	// ValidateOverrides checks group/per-alert override compatibility.
	ValidateOverrides(group string, a *alert.Alert) error
}

// MergeProviderDefaultAlertIntoEndpointAlert applies any unset fields from
// providerDefaultAlert onto endpointAlert. Mirrors the upstream behaviour.
func MergeProviderDefaultAlertIntoEndpointAlert(providerDefaultAlert, endpointAlert *alert.Alert) {
	if providerDefaultAlert == nil || endpointAlert == nil {
		return
	}
	if endpointAlert.Enabled == nil {
		endpointAlert.Enabled = providerDefaultAlert.Enabled
	}
	if endpointAlert.SendOnResolved == nil {
		endpointAlert.SendOnResolved = providerDefaultAlert.SendOnResolved
	}
	if endpointAlert.Description == nil {
		endpointAlert.Description = providerDefaultAlert.Description
	}
	if endpointAlert.FailureThreshold == 0 {
		endpointAlert.FailureThreshold = providerDefaultAlert.FailureThreshold
	}
	if endpointAlert.SuccessThreshold == 0 {
		endpointAlert.SuccessThreshold = providerDefaultAlert.SuccessThreshold
	}
}

// ErrUnknownProvider is returned by Get when no provider is registered for a type.
var ErrUnknownProvider = fmt.Errorf("no alert provider registered for type")
