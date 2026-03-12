package devtools

import (
	"github.com/spcent/plumego/core"
	"github.com/spcent/plumego/metrics"
)

// NewAppComponent builds a devtools component wired to a core.App instance.
func NewAppComponent(app *core.App) *DevToolsComponent {
	if app == nil {
		return NewComponent(Options{})
	}

	return NewComponent(Options{
		Debug:   app.DebugEnabled(),
		Logger:  app.Logger(),
		EnvFile: app.EnvPath(),
		Hooks: Hooks{
			ConfigSnapshot: app.ConfigSnapshot,
			MiddlewareList: app.MiddlewareNames,
			AttachDevMetrics: func(dev *metrics.DevCollector) {
				app.AttachHTTPObserver(dev)
			},
		},
	})
}
