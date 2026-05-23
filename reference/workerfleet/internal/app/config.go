package app

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"workerfleet/internal/domain"
	"workerfleet/internal/platform/store"
)

const (
	StoreBackendMemory = "memory"
	StoreBackendMongo  = "mongo"
	ProfileDevelopment = "dev"
	ProfileProduction  = "prod"

	nanosecondsPerRetentionDay = uint64((24 * time.Hour) / time.Nanosecond)
	maxRetentionDays           = uint64(1<<63-1) / nanosecondsPerRetentionDay
	defaultLoopTimeout         = 25 * time.Second
	defaultLoopFailureBackoff  = 5 * time.Second
	defaultLoopMaxBackoff      = 1 * time.Minute
	defaultLoopLeaseTTL        = 90 * time.Second
	minStatusStaleAfter        = 5 * time.Second
	minStatusOfflineAfter      = 10 * time.Second
	minStageStuckAfter         = 30 * time.Second
)

type Config struct {
	StoreBackend string
	Profile      string
	Policy       PolicyConfig
	Metrics      MetricsConfig
	Mongo        MongoConfig
	Retention    time.Duration
	Runtime      RuntimeConfig
	Kube         KubeConfig
	Notifier     NotifierConfig
	WorkerAuth   WorkerIngressAuthConfig
	AdminAuth    AdminAuthConfig
}

type PolicyConfig struct {
	Status domain.StatusPolicy
	Alert  domain.AlertPolicy
}

type MetricsConfig struct {
	ExperimentalSeriesEnabled bool
}

type MongoConfig struct {
	URI              string
	Database         string
	ConnectTimeout   time.Duration
	OperationTimeout time.Duration
	MaxPoolSize      uint64
}

type RuntimeConfig struct {
	KubeSyncEnabled              bool
	StatusSweepEnabled           bool
	AlertEvaluationEnabled       bool
	NotificationEnabled          bool
	KubeSyncInterval             time.Duration
	StatusSweepInterval          time.Duration
	AlertEvaluationInterval      time.Duration
	NotificationDeliveryInterval time.Duration
	NotifierDeliveryTimeout      time.Duration
	LoopLeaseTTL                 time.Duration
	LoopLeaseOwner               string
}

type KubeConfig struct {
	APIHost         string
	BearerToken     string
	Namespace       string
	LabelSelector   string
	WorkerContainer string
}

type NotifierConfig struct {
	FeishuWebhookURL string
	WebhookURL       string
	WebhookHeaders   map[string]string
}

type WorkerIngressAuthConfig struct {
	Token string
}

type AdminAuthConfig struct {
	Token    string
	Required bool
}

func DefaultConfig() Config {
	return DefaultConfigForProfile(ProfileDevelopment)
}

func DefaultConfigForProfile(profile string) Config {
	normalized := normalizeProfile(profile)
	statusPolicy := defaultStatusPolicyForProfile(normalized)
	return Config{
		StoreBackend: StoreBackendMemory,
		Profile:      normalized,
		Policy: PolicyConfig{
			Status: statusPolicy,
			Alert:  statusPolicy.AlertPolicy(),
		},
		Metrics: MetricsConfig{
			ExperimentalSeriesEnabled: normalized == ProfileDevelopment,
		},
		Mongo: MongoConfig{
			ConnectTimeout:   10 * time.Second,
			OperationTimeout: 10 * time.Second,
		},
		Retention: store.DefaultRetention,
		Runtime: RuntimeConfig{
			KubeSyncInterval:             30 * time.Second,
			StatusSweepInterval:          30 * time.Second,
			AlertEvaluationInterval:      30 * time.Second,
			NotificationDeliveryInterval: 30 * time.Second,
			NotifierDeliveryTimeout:      5 * time.Second,
			LoopLeaseTTL:                 defaultLoopLeaseTTL,
			LoopLeaseOwner:               defaultLoopLeaseOwner(),
		},
		Kube: KubeConfig{
			WorkerContainer: "worker",
		},
		AdminAuth: AdminAuthConfig{
			Required: normalized == ProfileProduction,
		},
	}
}

func LoadConfigFromEnv() (Config, error) {
	return LoadConfig(os.LookupEnv)
}

func LoadConfig(lookup func(string) (string, bool)) (Config, error) {
	if lookup == nil {
		lookup = os.LookupEnv
	}

	cfg := DefaultConfig()
	if value, ok := lookup("WORKERFLEET_PROFILE"); ok {
		profile, err := parseProfileEnv(value)
		if err != nil {
			return Config{}, err
		}
		cfg = DefaultConfigForProfile(profile)
	}
	if value, ok := lookup("WORKERFLEET_STORE_BACKEND"); ok {
		cfg.StoreBackend = strings.ToLower(strings.TrimSpace(value))
	}
	if value, ok := lookup("WORKERFLEET_MONGO_URI"); ok {
		cfg.Mongo.URI = strings.TrimSpace(value)
	}
	if value, ok := lookup("WORKERFLEET_MONGO_DATABASE"); ok {
		cfg.Mongo.Database = strings.TrimSpace(value)
	}
	if value, ok := lookup("WORKERFLEET_MONGO_CONNECT_TIMEOUT"); ok {
		timeout, err := parseDurationEnv("WORKERFLEET_MONGO_CONNECT_TIMEOUT", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Mongo.ConnectTimeout = timeout
	}
	if value, ok := lookup("WORKERFLEET_MONGO_OPERATION_TIMEOUT"); ok {
		timeout, err := parseDurationEnv("WORKERFLEET_MONGO_OPERATION_TIMEOUT", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Mongo.OperationTimeout = timeout
	}
	if value, ok := lookup("WORKERFLEET_MONGO_MAX_POOL_SIZE"); ok {
		poolSize, err := parseUintEnv("WORKERFLEET_MONGO_MAX_POOL_SIZE", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Mongo.MaxPoolSize = poolSize
	}
	if value, ok := lookup("WORKERFLEET_RETENTION_DAYS"); ok {
		days, err := parseUintEnv("WORKERFLEET_RETENTION_DAYS", value)
		if err != nil {
			return Config{}, err
		}
		if days == 0 {
			return Config{}, errors.New("WORKERFLEET_RETENTION_DAYS must be greater than zero")
		}
		if days > maxRetentionDays {
			return Config{}, errors.New("WORKERFLEET_RETENTION_DAYS is too large")
		}
		cfg.Retention = time.Duration(days) * 24 * time.Hour
	}
	if value, ok := lookup("WORKERFLEET_KUBE_SYNC_ENABLED"); ok {
		enabled, err := parseBoolEnv("WORKERFLEET_KUBE_SYNC_ENABLED", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Runtime.KubeSyncEnabled = enabled
	}
	if value, ok := lookup("WORKERFLEET_STATUS_SWEEP_ENABLED"); ok {
		enabled, err := parseBoolEnv("WORKERFLEET_STATUS_SWEEP_ENABLED", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Runtime.StatusSweepEnabled = enabled
	}
	if value, ok := lookup("WORKERFLEET_ALERT_EVALUATION_ENABLED"); ok {
		enabled, err := parseBoolEnv("WORKERFLEET_ALERT_EVALUATION_ENABLED", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Runtime.AlertEvaluationEnabled = enabled
	}
	if value, ok := lookup("WORKERFLEET_NOTIFICATION_ENABLED"); ok {
		enabled, err := parseBoolEnv("WORKERFLEET_NOTIFICATION_ENABLED", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Runtime.NotificationEnabled = enabled
	}
	if value, ok := lookup("WORKERFLEET_KUBE_SYNC_INTERVAL"); ok {
		interval, err := parseDurationEnv("WORKERFLEET_KUBE_SYNC_INTERVAL", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Runtime.KubeSyncInterval = interval
	}
	if value, ok := lookup("WORKERFLEET_STATUS_SWEEP_INTERVAL"); ok {
		interval, err := parseDurationEnv("WORKERFLEET_STATUS_SWEEP_INTERVAL", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Runtime.StatusSweepInterval = interval
	}
	if value, ok := lookup("WORKERFLEET_ALERT_EVALUATION_INTERVAL"); ok {
		interval, err := parseDurationEnv("WORKERFLEET_ALERT_EVALUATION_INTERVAL", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Runtime.AlertEvaluationInterval = interval
	}
	if value, ok := lookup("WORKERFLEET_NOTIFICATION_DELIVERY_INTERVAL"); ok {
		interval, err := parseDurationEnv("WORKERFLEET_NOTIFICATION_DELIVERY_INTERVAL", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Runtime.NotificationDeliveryInterval = interval
	}
	if value, ok := lookup("WORKERFLEET_NOTIFIER_DELIVERY_TIMEOUT"); ok {
		timeout, err := parseDurationEnv("WORKERFLEET_NOTIFIER_DELIVERY_TIMEOUT", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Runtime.NotifierDeliveryTimeout = timeout
	}
	if value, ok := lookup("WORKERFLEET_LOOP_LEASE_TTL"); ok {
		ttl, err := parseDurationEnv("WORKERFLEET_LOOP_LEASE_TTL", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Runtime.LoopLeaseTTL = ttl
	}
	if value, ok := lookup("WORKERFLEET_LOOP_LEASE_OWNER"); ok {
		owner := strings.TrimSpace(value)
		if owner == "" {
			return Config{}, errors.New("WORKERFLEET_LOOP_LEASE_OWNER must not be empty when set")
		}
		cfg.Runtime.LoopLeaseOwner = owner
	}
	if value, ok := lookup("WORKERFLEET_STATUS_STALE_AFTER"); ok {
		duration, err := parseDurationEnv("WORKERFLEET_STATUS_STALE_AFTER", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Policy.Status.StaleAfter = duration
	}
	if value, ok := lookup("WORKERFLEET_STATUS_OFFLINE_AFTER"); ok {
		duration, err := parseDurationEnv("WORKERFLEET_STATUS_OFFLINE_AFTER", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Policy.Status.OfflineAfter = duration
	}
	if value, ok := lookup("WORKERFLEET_STATUS_STAGE_STUCK_AFTER"); ok {
		duration, err := parseDurationEnv("WORKERFLEET_STATUS_STAGE_STUCK_AFTER", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Policy.Status.StageStuckAfter = duration
	}
	if value, ok := lookup("WORKERFLEET_STATUS_RESTART_BURST_THRESHOLD"); ok {
		threshold, err := parseInt32Env("WORKERFLEET_STATUS_RESTART_BURST_THRESHOLD", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Policy.Status.RestartBurstThreshold = threshold
	}
	cfg.Policy.Alert = cfg.Policy.Status.AlertPolicy()
	if value, ok := lookup("WORKERFLEET_ALERT_STAGE_STUCK_AFTER"); ok {
		duration, err := parseDurationEnv("WORKERFLEET_ALERT_STAGE_STUCK_AFTER", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Policy.Alert.StageStuckAfter = duration
	}
	if value, ok := lookup("WORKERFLEET_ALERT_RESTART_BURST_THRESHOLD"); ok {
		threshold, err := parseInt32Env("WORKERFLEET_ALERT_RESTART_BURST_THRESHOLD", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Policy.Alert.RestartBurstThreshold = threshold
	}
	if value, ok := lookup("WORKERFLEET_EXPERIMENTAL_METRICS_ENABLED"); ok {
		enabled, err := parseBoolEnv("WORKERFLEET_EXPERIMENTAL_METRICS_ENABLED", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Metrics.ExperimentalSeriesEnabled = enabled
	}
	if value, ok := lookup("WORKERFLEET_KUBE_API_HOST"); ok {
		cfg.Kube.APIHost = strings.TrimSpace(value)
	}
	if value, ok := lookup("WORKERFLEET_KUBE_BEARER_TOKEN"); ok {
		cfg.Kube.BearerToken = strings.TrimSpace(value)
	}
	if value, ok := lookup("WORKERFLEET_KUBE_NAMESPACE"); ok {
		cfg.Kube.Namespace = strings.TrimSpace(value)
	}
	if value, ok := lookup("WORKERFLEET_KUBE_LABEL_SELECTOR"); ok {
		cfg.Kube.LabelSelector = strings.TrimSpace(value)
	}
	if value, ok := lookup("WORKERFLEET_KUBE_WORKER_CONTAINER"); ok {
		cfg.Kube.WorkerContainer = strings.TrimSpace(value)
	}
	if value, ok := lookup("WORKERFLEET_FEISHU_WEBHOOK_URL"); ok {
		cfg.Notifier.FeishuWebhookURL = strings.TrimSpace(value)
	}
	if value, ok := lookup("WORKERFLEET_WEBHOOK_URL"); ok {
		cfg.Notifier.WebhookURL = strings.TrimSpace(value)
	}
	if value, ok := lookup("WORKERFLEET_WEBHOOK_HEADERS"); ok {
		headers, err := parseHeadersEnv("WORKERFLEET_WEBHOOK_HEADERS", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Notifier.WebhookHeaders = headers
	}
	if value, ok := lookup("WORKERFLEET_WORKER_AUTH_TOKEN"); ok {
		token := strings.TrimSpace(value)
		if token == "" {
			return Config{}, errors.New("WORKERFLEET_WORKER_AUTH_TOKEN must not be empty when set")
		}
		cfg.WorkerAuth.Token = token
	}
	if value, ok := lookup("WORKERFLEET_QUERY_AUTH_REQUIRED"); ok {
		required, err := parseBoolEnv("WORKERFLEET_QUERY_AUTH_REQUIRED", value)
		if err != nil {
			return Config{}, err
		}
		cfg.AdminAuth.Required = required
	}
	if value, ok := lookup("WORKERFLEET_ADMIN_AUTH_TOKEN"); ok {
		token := strings.TrimSpace(value)
		if token == "" {
			return Config{}, errors.New("WORKERFLEET_ADMIN_AUTH_TOKEN must not be empty when set")
		}
		cfg.AdminAuth.Token = token
	}

	return cfg, ValidateConfig(cfg)
}

func ValidateConfig(cfg Config) error {
	if _, err := validateProfile(cfg.Profile); err != nil {
		return err
	}
	switch cfg.StoreBackend {
	case "", StoreBackendMemory:
	case StoreBackendMongo:
		if strings.TrimSpace(cfg.Mongo.URI) == "" {
			return errors.New("WORKERFLEET_MONGO_URI is required when WORKERFLEET_STORE_BACKEND=mongo")
		}
		if strings.TrimSpace(cfg.Mongo.Database) == "" {
			return errors.New("WORKERFLEET_MONGO_DATABASE is required when WORKERFLEET_STORE_BACKEND=mongo")
		}
	default:
		return fmt.Errorf("unsupported WORKERFLEET_STORE_BACKEND %q", cfg.StoreBackend)
	}
	if cfg.Runtime.KubeSyncEnabled && strings.TrimSpace(cfg.Kube.WorkerContainer) == "" {
		return errors.New("WORKERFLEET_KUBE_WORKER_CONTAINER is required when WORKERFLEET_KUBE_SYNC_ENABLED=true")
	}
	if cfg.Runtime.LoopLeaseTTL < defaultLoopTimeout {
		return fmt.Errorf("WORKERFLEET_LOOP_LEASE_TTL must be at least %s", defaultLoopTimeout)
	}
	if strings.TrimSpace(cfg.Runtime.LoopLeaseOwner) == "" {
		return errors.New("WORKERFLEET_LOOP_LEASE_OWNER is required")
	}
	if cfg.Profile == ProfileProduction && strings.TrimSpace(cfg.WorkerAuth.Token) == "" {
		return errors.New("WORKERFLEET_WORKER_AUTH_TOKEN is required when WORKERFLEET_PROFILE=prod")
	}
	if cfg.Profile == ProfileProduction && strings.TrimSpace(cfg.AdminAuth.Token) == "" {
		return errors.New("WORKERFLEET_ADMIN_AUTH_TOKEN is required when WORKERFLEET_PROFILE=prod")
	}
	if cfg.AdminAuth.Required && strings.TrimSpace(cfg.AdminAuth.Token) == "" {
		return errors.New("WORKERFLEET_ADMIN_AUTH_TOKEN is required when WORKERFLEET_QUERY_AUTH_REQUIRED=true")
	}
	if cfg.Runtime.NotificationEnabled &&
		strings.TrimSpace(cfg.Notifier.FeishuWebhookURL) == "" &&
		strings.TrimSpace(cfg.Notifier.WebhookURL) == "" {
		return errors.New("WORKERFLEET_FEISHU_WEBHOOK_URL or WORKERFLEET_WEBHOOK_URL is required when WORKERFLEET_NOTIFICATION_ENABLED=true")
	}
	if err := validateStatusPolicyConfig(cfg.Policy.Status); err != nil {
		return err
	}
	if err := validateAlertPolicyConfig(cfg.Policy.Alert); err != nil {
		return err
	}
	return nil
}

func configuredPolicies(cfg Config) (domain.StatusPolicy, domain.AlertPolicy) {
	statusPolicy := cfg.Policy.Status
	if statusPolicy == (domain.StatusPolicy{}) {
		statusPolicy = defaultStatusPolicyForProfile(cfg.Profile)
	}
	alertPolicy := cfg.Policy.Alert
	if alertPolicy == (domain.AlertPolicy{}) {
		alertPolicy = statusPolicy.AlertPolicy()
	}
	return statusPolicy, alertPolicy
}

func (cfg RuntimeConfig) kubeSyncLoopSettings() loopExecutionSettings {
	return cfg.loopSettings("kube_sync", cfg.KubeSyncInterval)
}

func (cfg RuntimeConfig) statusSweepLoopSettings() loopExecutionSettings {
	return cfg.loopSettings("status_sweep", cfg.StatusSweepInterval)
}

func (cfg RuntimeConfig) alertEvaluationLoopSettings() loopExecutionSettings {
	return cfg.loopSettings("alert_evaluate", cfg.AlertEvaluationInterval)
}

func (cfg RuntimeConfig) notificationDeliveryLoopSettings() loopExecutionSettings {
	return cfg.loopSettings("notification_deliver", cfg.NotificationDeliveryInterval)
}

func (cfg RuntimeConfig) loopSettings(name string, interval time.Duration) loopExecutionSettings {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return loopExecutionSettings{
		Name:              name,
		Interval:          interval,
		Timeout:           defaultLoopTimeout,
		FailureBackoff:    defaultLoopFailureBackoff,
		MaxFailureBackoff: defaultLoopMaxBackoff,
	}
}

func parseDurationEnv(name string, value string) (time.Duration, error) {
	duration, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", name)
	}
	return duration, nil
}

func parseUintEnv(name string, value string) (uint64, error) {
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}
	return parsed, nil
}

func parseInt32Env(name string, value string) (int32, error) {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 32)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}
	return int32(parsed), nil
}

func parseBoolEnv(name string, value string) (bool, error) {
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false, fmt.Errorf("parse %s: %w", name, err)
	}
	return parsed, nil
}

func parseHeadersEnv(name string, value string) (map[string]string, error) {
	headers := map[string]string{}
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, headerValue, ok := strings.Cut(part, "=")
		key = strings.TrimSpace(key)
		headerValue = strings.TrimSpace(headerValue)
		if !ok || key == "" {
			return nil, fmt.Errorf("parse %s: header entries must use key=value", name)
		}
		headers[key] = headerValue
	}
	return headers, nil
}

func parseProfileEnv(value string) (string, error) {
	return validateProfile(strings.TrimSpace(value))
}

func validateProfile(profile string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(profile)) {
	case "", ProfileDevelopment:
		return ProfileDevelopment, nil
	case ProfileProduction, "production":
		return ProfileProduction, nil
	default:
		return "", fmt.Errorf("unsupported WORKERFLEET_PROFILE %q", profile)
	}
}

func normalizeProfile(profile string) string {
	normalized, err := validateProfile(profile)
	if err != nil {
		return ProfileDevelopment
	}
	return normalized
}

func defaultLoopLeaseOwner() string {
	hostname, err := os.Hostname()
	if err == nil && strings.TrimSpace(hostname) != "" {
		return strings.TrimSpace(hostname)
	}
	return "workerfleet"
}

func defaultStatusPolicyForProfile(profile string) domain.StatusPolicy {
	switch normalizeProfile(profile) {
	case ProfileProduction:
		return domain.StatusPolicy{
			StaleAfter:            60 * time.Second,
			OfflineAfter:          2 * time.Minute,
			StageStuckAfter:       15 * time.Minute,
			RestartBurstThreshold: 5,
		}
	default:
		return domain.DefaultStatusPolicy()
	}
}

func validateStatusPolicyConfig(policy domain.StatusPolicy) error {
	if err := policy.Validate(); err != nil {
		return fmt.Errorf("invalid worker status policy: %w", err)
	}
	if policy.StaleAfter < minStatusStaleAfter {
		return fmt.Errorf("WORKERFLEET_STATUS_STALE_AFTER must be at least %s", minStatusStaleAfter)
	}
	if policy.OfflineAfter < minStatusOfflineAfter {
		return fmt.Errorf("WORKERFLEET_STATUS_OFFLINE_AFTER must be at least %s", minStatusOfflineAfter)
	}
	if policy.StageStuckAfter < minStageStuckAfter {
		return fmt.Errorf("WORKERFLEET_STATUS_STAGE_STUCK_AFTER must be at least %s", minStageStuckAfter)
	}
	return nil
}

func validateAlertPolicyConfig(policy domain.AlertPolicy) error {
	if err := policy.Validate(); err != nil {
		return fmt.Errorf("invalid alert policy: %w", err)
	}
	if policy.StageStuckAfter < minStageStuckAfter {
		return fmt.Errorf("WORKERFLEET_ALERT_STAGE_STUCK_AFTER must be at least %s", minStageStuckAfter)
	}
	return nil
}
