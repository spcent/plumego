package config

import "errors"

const (
	defaultTitle               = "Health Dashboard | Gatus"
	defaultDescription         = "Gatus is an advanced automated status page that lets you monitor your applications and configure alerts to notify you if there's an issue"
	defaultHeader              = "Gatus"
	defaultDashboardHeading    = "Health Dashboard"
	defaultDashboardSubheading = "Monitor the health of your endpoints in real-time"
	defaultLogo                = ""
	defaultLink                = ""
	defaultFavicon             = "/favicon.ico"
	defaultFavicon16           = "/favicon-16x16.png"
	defaultFavicon32           = "/favicon-32x32.png"
	defaultSortBy              = "name"
	defaultFilterBy            = "none"
	defaultLoginSubtitle       = "System Monitoring Dashboard"
)

var (
	defaultDarkMode = true

	ErrButtonValidationFailed = errors.New("invalid button configuration: missing required name or link")
	ErrInvalidDefaultSortBy   = errors.New("invalid default-sort-by value: must be 'name', 'group', or 'health'")
	ErrInvalidDefaultFilterBy = errors.New("invalid default-filter-by value: must be 'none', 'failing', or 'unstable'")
)

// UIConfig describes user-visible labels and assets.
type UIConfig struct {
	Title               string   `json:"title,omitempty"`
	Description         string   `json:"description,omitempty"`
	DashboardHeading    string   `json:"dashboard-heading,omitempty"`
	DashboardSubheading string   `json:"dashboard-subheading,omitempty"`
	Header              string   `json:"header,omitempty"`
	Logo                string   `json:"logo,omitempty"`
	Link                string   `json:"link,omitempty"`
	Favicon             Favicon  `json:"favicon,omitempty"`
	Buttons             []Button `json:"buttons,omitempty"`
	DarkMode            *bool    `json:"dark-mode,omitempty"`
	DefaultSortBy       string   `json:"default-sort-by,omitempty"`
	DefaultFilterBy     string   `json:"default-filter-by,omitempty"`
	LoginSubtitle       string   `json:"login-subtitle,omitempty"`

	MaximumNumberOfResults int `json:"-"`
}

func (cfg *UIConfig) IsDarkMode() bool {
	if cfg.DarkMode != nil {
		return *cfg.DarkMode
	}
	return defaultDarkMode
}

type Button struct {
	Name string `json:"name,omitempty"`
	Link string `json:"link,omitempty"`
}

func (b *Button) Validate() error {
	if len(b.Name) == 0 || len(b.Link) == 0 {
		return ErrButtonValidationFailed
	}
	return nil
}

type Favicon struct {
	Default   string `json:"default,omitempty"`
	Size16x16 string `json:"size16x16,omitempty"`
	Size32x32 string `json:"size32x32,omitempty"`
}

func defaultUIConfig() *UIConfig {
	return &UIConfig{
		Title:                  defaultTitle,
		Description:            defaultDescription,
		DashboardHeading:       defaultDashboardHeading,
		DashboardSubheading:    defaultDashboardSubheading,
		Header:                 defaultHeader,
		Logo:                   defaultLogo,
		Link:                   defaultLink,
		DarkMode:               &defaultDarkMode,
		DefaultSortBy:          defaultSortBy,
		DefaultFilterBy:        defaultFilterBy,
		LoginSubtitle:          defaultLoginSubtitle,
		MaximumNumberOfResults: 100,
		Favicon: Favicon{
			Default:   defaultFavicon,
			Size16x16: defaultFavicon16,
			Size32x32: defaultFavicon32,
		},
	}
}

// ValidateAndSetDefaults fills missing fields and rejects invalid enums.
func (cfg *UIConfig) ValidateAndSetDefaults() error {
	if len(cfg.Title) == 0 {
		cfg.Title = defaultTitle
	}
	if len(cfg.Description) == 0 {
		cfg.Description = defaultDescription
	}
	if len(cfg.DashboardHeading) == 0 {
		cfg.DashboardHeading = defaultDashboardHeading
	}
	if len(cfg.DashboardSubheading) == 0 {
		cfg.DashboardSubheading = defaultDashboardSubheading
	}
	if len(cfg.Header) == 0 {
		cfg.Header = defaultHeader
	}
	if cfg.DarkMode == nil {
		cfg.DarkMode = &defaultDarkMode
	}
	if len(cfg.DefaultSortBy) == 0 {
		cfg.DefaultSortBy = defaultSortBy
	} else if cfg.DefaultSortBy != "name" && cfg.DefaultSortBy != "group" && cfg.DefaultSortBy != "health" {
		return ErrInvalidDefaultSortBy
	}
	if len(cfg.DefaultFilterBy) == 0 {
		cfg.DefaultFilterBy = defaultFilterBy
	} else if cfg.DefaultFilterBy != "none" && cfg.DefaultFilterBy != "failing" && cfg.DefaultFilterBy != "unstable" {
		return ErrInvalidDefaultFilterBy
	}
	if len(cfg.LoginSubtitle) == 0 {
		cfg.LoginSubtitle = defaultLoginSubtitle
	}
	if len(cfg.Favicon.Default) == 0 {
		cfg.Favicon.Default = defaultFavicon
	}
	if len(cfg.Favicon.Size16x16) == 0 {
		cfg.Favicon.Size16x16 = defaultFavicon16
	}
	if len(cfg.Favicon.Size32x32) == 0 {
		cfg.Favicon.Size32x32 = defaultFavicon32
	}
	for _, btn := range cfg.Buttons {
		if err := btn.Validate(); err != nil {
			return err
		}
	}
	return nil
}
