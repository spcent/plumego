package endpoint

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"guardus/internal/client"
	"guardus/internal/domain/alert"
	"guardus/internal/domain/endpoint/dns"
	"guardus/internal/domain/endpoint/ui"
	"guardus/internal/domain/gontext"
	"guardus/internal/domain/key"
	"guardus/internal/domain/maintenance"
)

type Type string

const (
	HostHeader        = "Host"
	ContentTypeHeader = "Content-Type"
	UserAgentHeader   = "User-Agent"
	GatusUserAgent    = "Guardus/1.0"

	TypeDNS      Type = "DNS"
	TypeTCP      Type = "TCP"
	TypeICMP     Type = "ICMP"
	TypeSTARTTLS Type = "STARTTLS"
	TypeTLS      Type = "TLS"
	TypeHTTP     Type = "HTTP"
	TypeGRPC     Type = "GRPC"
	TypeWS       Type = "WEBSOCKET"
	TypeUNKNOWN  Type = "UNKNOWN"
)

var (
	ErrEndpointWithNoCondition = errors.New("you must specify at least one condition per endpoint")
	ErrEndpointWithNoURL       = errors.New("you must specify an url for each endpoint")
	ErrUnknownEndpointType     = errors.New("unknown endpoint type")
	ErrInvalidConditionFormat  = errors.New("invalid condition format: does not match '<VALUE> <COMPARATOR> <VALUE>'")

	// ErrInvalidEndpointIntervalForDomainExpirationPlaceholder is retained to keep
	// validation parity with upstream gatus, even though guardus v1 does not
	// implement domain expiration probing.
	ErrInvalidEndpointIntervalForDomainExpirationPlaceholder = errors.New("the minimum interval for an endpoint with a condition using the " + DomainExpirationPlaceholder + " placeholder is 300s (5m)")
)

// Endpoint is the configuration of a service to be monitored.
type Endpoint struct {
	Enabled                 *bool                 `json:"enabled,omitempty"`
	Name                    string                `json:"name"`
	Group                   string                `json:"group,omitempty"`
	URL                     string                `json:"url"`
	Method                  string                `json:"method,omitempty"`
	Body                    string                `json:"body,omitempty"`
	GraphQL                 bool                  `json:"graphql,omitempty"`
	Headers                 map[string]string     `json:"headers,omitempty"`
	ExtraLabels             map[string]string     `json:"extra-labels,omitempty"`
	Interval                time.Duration         `json:"interval,omitempty"`
	Conditions              []Condition           `json:"conditions"`
	Alerts                  []*alert.Alert        `json:"alerts,omitempty"`
	MaintenanceWindows      []*maintenance.Config `json:"maintenance-windows,omitempty"`
	DNSConfig               *dns.Config           `json:"dns,omitempty"`
	ClientConfig            *client.Config        `json:"client,omitempty"`
	UIConfig                *ui.Config            `json:"ui,omitempty"`
	NumberOfFailuresInARow  int                   `json:"-"`
	NumberOfSuccessesInARow int                   `json:"-"`
	LastReminderSent        time.Time             `json:"-"`
}

// UnmarshalJSON accepts Interval as either a duration string ("60s") or a
// numeric nanosecond value, so JSON configs preserve the YAML ergonomics.
func (e *Endpoint) UnmarshalJSON(data []byte) error {
	type endpointAlias Endpoint
	aux := struct {
		Interval any `json:"interval,omitempty"`
		*endpointAlias
	}{endpointAlias: (*endpointAlias)(e)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	switch v := aux.Interval.(type) {
	case nil:
		e.Interval = 0
	case string:
		if v == "" {
			e.Interval = 0
			return nil
		}
		d, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("endpoint interval: %w", err)
		}
		e.Interval = d
	case float64:
		e.Interval = time.Duration(v)
	default:
		return fmt.Errorf("endpoint interval: unsupported type %T", v)
	}
	return nil
}

func (e *Endpoint) IsEnabled() bool {
	if e.Enabled == nil {
		return true
	}
	return *e.Enabled
}

func (e *Endpoint) Type() Type {
	switch {
	case e.DNSConfig != nil:
		return TypeDNS
	case strings.HasPrefix(e.URL, "tcp://"):
		return TypeTCP
	case strings.HasPrefix(e.URL, "icmp://"):
		return TypeICMP
	case strings.HasPrefix(e.URL, "starttls://"):
		return TypeSTARTTLS
	case strings.HasPrefix(e.URL, "tls://"):
		return TypeTLS
	case strings.HasPrefix(e.URL, "http://") || strings.HasPrefix(e.URL, "https://"):
		return TypeHTTP
	case strings.HasPrefix(e.URL, "grpc://") || strings.HasPrefix(e.URL, "grpcs://"):
		return TypeGRPC
	case strings.HasPrefix(e.URL, "ws://") || strings.HasPrefix(e.URL, "wss://"):
		return TypeWS
	default:
		return TypeUNKNOWN
	}
}

func (e *Endpoint) ValidateAndSetDefaults() error {
	if err := validateEndpointNameGroupAndAlerts(e.Name, e.Group, e.Alerts); err != nil {
		return err
	}
	if len(e.URL) == 0 {
		return ErrEndpointWithNoURL
	}
	if e.ClientConfig == nil {
		e.ClientConfig = client.GetDefaultConfig()
	} else if err := e.ClientConfig.ValidateAndSetDefaults(); err != nil {
		return err
	}
	if e.UIConfig == nil {
		e.UIConfig = ui.GetDefaultConfig()
	} else if err := e.UIConfig.ValidateAndSetDefaults(); err != nil {
		return err
	}
	if e.Interval == 0 {
		e.Interval = 1 * time.Minute
	}
	if len(e.Method) == 0 {
		e.Method = http.MethodGet
	}
	if len(e.Headers) == 0 {
		e.Headers = make(map[string]string)
	}
	if !hasHeader(e.Headers, UserAgentHeader) {
		e.Headers[UserAgentHeader] = GatusUserAgent
	}
	if !hasHeader(e.Headers, ContentTypeHeader) && e.GraphQL {
		e.Headers[ContentTypeHeader] = "application/json"
	}
	if len(e.Conditions) == 0 {
		return ErrEndpointWithNoCondition
	}
	for _, c := range e.Conditions {
		if e.Interval < 5*time.Minute && c.hasDomainExpirationPlaceholder() {
			return ErrInvalidEndpointIntervalForDomainExpirationPlaceholder
		}
		if err := c.Validate(); err != nil {
			return fmt.Errorf("%v: %w", ErrInvalidConditionFormat, err)
		}
	}
	if e.DNSConfig != nil {
		return e.DNSConfig.ValidateAndSetDefault()
	}
	if e.Type() == TypeUNKNOWN {
		return ErrUnknownEndpointType
	}
	for _, mw := range e.MaintenanceWindows {
		if err := mw.ValidateAndSetDefaults(); err != nil {
			return err
		}
	}
	if _, err := http.NewRequest(e.Method, e.URL, bytes.NewBuffer([]byte(e.getParsedBody()))); err != nil {
		return err
	}
	return nil
}

func (e *Endpoint) DisplayName() string {
	if len(e.Group) > 0 {
		return e.Group + "/" + e.Name
	}
	return e.Name
}

func (e *Endpoint) Key() string { return key.ConvertGroupAndNameToKey(e.Group, e.Name) }

// Close releases idle HTTP connections so config reloads don't leak sockets.
func (e *Endpoint) Close() {
	if e.Type() == TypeHTTP {
		client.GetHTTPClient(e.ClientConfig).CloseIdleConnections()
	}
}

func (e *Endpoint) EvaluateHealth() *Result {
	return e.EvaluateHealthWithContext(nil)
}

func (e *Endpoint) EvaluateHealthWithContext(ctx *gontext.Gontext) *Result {
	result := &Result{Success: true, Errors: []string{}}
	processed := e
	if ctx != nil {
		processed = e.preprocessWithContext(result, ctx)
	}
	if processed.DNSConfig != nil {
		result.Hostname = strings.TrimSuffix(processed.URL, ":53")
	} else if processed.Type() == TypeICMP {
		result.Hostname = strings.TrimPrefix(processed.URL, "icmp://")
	} else {
		urlObject, err := url.Parse(processed.URL)
		if err != nil {
			result.AddError(err.Error())
		} else {
			result.Hostname = urlObject.Hostname()
			result.port = urlObject.Port()
		}
	}
	if processed.needsToRetrieveIP() {
		processed.getIP(result)
	}
	if processed.needsToRetrieveDomainExpiration() && len(result.Hostname) > 0 {
		// guardus v1 does not implement domain expiration probing; record a
		// non-fatal error so users see the placeholder is unsupported.
		result.AddError("domain expiration probing is not supported in guardus v1")
	}
	if len(result.Errors) == 0 {
		processed.call(result)
	} else {
		result.Success = false
	}
	for _, condition := range processed.Conditions {
		success := condition.evaluate(result, processed.UIConfig.DontResolveFailedConditions, processed.UIConfig.ResolveSuccessfulConditions, ctx)
		if !success {
			result.Success = false
		}
	}
	result.Timestamp = time.Now()
	if processed.UIConfig.HideURL {
		for i, errStr := range result.Errors {
			result.Errors[i] = strings.ReplaceAll(errStr, processed.URL, "<redacted>")
		}
	}
	if processed.UIConfig.HideHostname {
		for i, errStr := range result.Errors {
			result.Errors[i] = strings.ReplaceAll(errStr, result.Hostname, "<redacted>")
		}
		result.Hostname = ""
	}
	if processed.UIConfig.HidePort && len(result.port) > 0 {
		for i, errStr := range result.Errors {
			result.Errors[i] = strings.ReplaceAll(errStr, result.port, "<redacted>")
		}
		result.port = ""
	}
	if processed.UIConfig.HideErrors {
		result.Errors = nil
	}
	if processed.UIConfig.HideConditions {
		result.ConditionResults = nil
	}
	return result
}

func (e *Endpoint) preprocessWithContext(result *Result, ctx *gontext.Gontext) *Endpoint {
	processed := &Endpoint{}
	*processed = *e
	var err error
	if processed.URL, err = replaceContextPlaceholders(e.URL, ctx); err != nil {
		result.AddError(err.Error())
	}
	if processed.Body, err = replaceContextPlaceholders(e.Body, ctx); err != nil {
		result.AddError(err.Error())
	}
	if e.Headers != nil {
		processed.Headers = make(map[string]string)
		for k, v := range e.Headers {
			if processed.Headers[k], err = replaceContextPlaceholders(v, ctx); err != nil {
				result.AddError(err.Error())
			}
		}
	}
	return processed
}

func replaceContextPlaceholders(input string, ctx *gontext.Gontext) (string, error) {
	if ctx == nil {
		return input, nil
	}
	var contextErrors []string
	contextRegex := regexp.MustCompile(`\[CONTEXT\]\.[\w\.\-]+`)
	out := contextRegex.ReplaceAllStringFunc(input, func(match string) string {
		path := strings.TrimPrefix(match, "[CONTEXT].")
		v, err := ctx.Get(path)
		if err != nil {
			contextErrors = append(contextErrors, fmt.Sprintf("path '%s' not found", path))
			return match
		}
		return fmt.Sprintf("%v", v)
	})
	if len(contextErrors) > 0 {
		return out, fmt.Errorf("context placeholder resolution failed: %s", strings.Join(contextErrors, ", "))
	}
	return out, nil
}

func (e *Endpoint) getParsedBody() string {
	body := e.Body
	body = strings.ReplaceAll(body, "[ENDPOINT_NAME]", e.Name)
	body = strings.ReplaceAll(body, "[ENDPOINT_GROUP]", e.Group)
	body = strings.ReplaceAll(body, "[ENDPOINT_URL]", e.URL)
	randRegex, err := regexp.Compile(`\[RANDOM_STRING_\d+\]`)
	if err == nil {
		body = randRegex.ReplaceAllStringFunc(body, func(match string) string {
			n, _ := strconv.Atoi(match[15 : len(match)-1])
			if n > 8192 {
				n = 8192
			}
			const availableCharacterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
			b := make([]byte, n)
			for i := range b {
				b[i] = availableCharacterBytes[rand.Intn(len(availableCharacterBytes))]
			}
			return string(b)
		})
	}
	return body
}

func (e *Endpoint) getIP(result *Result) {
	ips, err := net.LookupIP(result.Hostname)
	if err != nil {
		result.AddError(err.Error())
		return
	}
	result.IP = ips[0].String()
}

func (e *Endpoint) call(result *Result) {
	var (
		request     *http.Request
		response    *http.Response
		err         error
		certificate *x509.Certificate
	)
	t := e.Type()
	if t == TypeHTTP {
		request = e.buildHTTPRequest()
	}
	startTime := time.Now()
	switch t {
	case TypeDNS:
		result.Connected, result.DNSRCode, result.Body, err = client.QueryDNS(e.DNSConfig.QueryType, e.DNSConfig.QueryName, e.URL)
		if err != nil {
			result.AddError(err.Error())
			return
		}
		result.Duration = time.Since(startTime)
	case TypeSTARTTLS, TypeTLS:
		if t == TypeSTARTTLS {
			result.Connected, certificate, err = client.CanPerformStartTLS(strings.TrimPrefix(e.URL, "starttls://"), e.ClientConfig)
		} else {
			result.Connected, result.Body, certificate, err = client.CanPerformTLS(strings.TrimPrefix(e.URL, "tls://"), e.getParsedBody(), e.ClientConfig)
		}
		if err != nil {
			result.AddError(err.Error())
			return
		}
		result.Duration = time.Since(startTime)
		result.CertificateExpiration = time.Until(certificate.NotAfter)
	case TypeTCP:
		result.Connected, result.Body = client.CanCreateNetworkConnection("tcp", strings.TrimPrefix(e.URL, "tcp://"), e.getParsedBody(), e.ClientConfig)
		result.Duration = time.Since(startTime)
	case TypeICMP:
		result.Connected, result.Duration = client.Ping(strings.TrimPrefix(e.URL, "icmp://"), e.ClientConfig)
	case TypeWS:
		wsHeaders := map[string]string{}
		if e.Headers != nil {
			maps.Copy(wsHeaders, e.Headers)
		}
		if !hasHeader(wsHeaders, UserAgentHeader) {
			wsHeaders[UserAgentHeader] = GatusUserAgent
		}
		result.Connected, result.Body, err = client.QueryWebSocket(e.URL, e.getParsedBody(), wsHeaders, e.ClientConfig)
		if err != nil {
			result.AddError(err.Error())
			return
		}
		result.Duration = time.Since(startTime)
	case TypeGRPC:
		useTLS := strings.HasPrefix(e.URL, "grpcs://")
		address := strings.TrimPrefix(strings.TrimPrefix(e.URL, "grpcs://"), "grpc://")
		connected, status, gerr, duration := client.PerformGRPCHealthCheck(address, useTLS, e.ClientConfig)
		if gerr != nil {
			result.AddError(gerr.Error())
			return
		}
		result.Connected = connected
		result.Duration = duration
		if e.needsToReadBody() {
			result.Body = []byte(fmt.Sprintf("{\"status\":\"%s\"}", status))
		}
	default:
		response, err = client.GetHTTPClient(e.ClientConfig).Do(request)
		result.Duration = time.Since(startTime)
		if err != nil {
			result.AddError(err.Error())
			return
		}
		defer response.Body.Close()
		if response.TLS != nil && len(response.TLS.PeerCertificates) > 0 {
			certificate = response.TLS.PeerCertificates[0]
			result.CertificateExpiration = time.Until(certificate.NotAfter)
		}
		result.HTTPStatus = response.StatusCode
		result.Connected = response.StatusCode > 0
		if e.needsToReadBody() {
			result.Body, err = io.ReadAll(response.Body)
			if err != nil {
				result.AddError("error reading response body:" + err.Error())
			}
		}
	}
}

func (e *Endpoint) buildHTTPRequest() *http.Request {
	var bodyBuffer *bytes.Buffer
	if e.GraphQL {
		graphQlBody := map[string]string{"query": e.getParsedBody()}
		body, _ := json.Marshal(graphQlBody)
		bodyBuffer = bytes.NewBuffer(body)
	} else {
		bodyBuffer = bytes.NewBuffer([]byte(e.getParsedBody()))
	}
	request, _ := http.NewRequest(e.Method, e.URL, bodyBuffer)
	for k, v := range e.Headers {
		request.Header.Set(k, v)
		if strings.EqualFold(k, HostHeader) {
			request.Host = v
		}
	}
	return request
}

func (e *Endpoint) needsToReadBody() bool {
	for _, condition := range e.Conditions {
		if condition.hasBodyPlaceholder() {
			return true
		}
	}
	return false
}

func (e *Endpoint) needsToRetrieveDomainExpiration() bool {
	for _, condition := range e.Conditions {
		if condition.hasDomainExpirationPlaceholder() {
			return true
		}
	}
	return false
}

func (e *Endpoint) needsToRetrieveIP() bool {
	for _, condition := range e.Conditions {
		if condition.hasIPPlaceholder() {
			return true
		}
	}
	return false
}

func hasHeader(headers map[string]string, name string) bool {
	for k := range headers {
		if strings.EqualFold(k, name) {
			return true
		}
	}
	return false
}
