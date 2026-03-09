package core

import (
	"reflect"

	"github.com/spcent/plumego/core/components/devtools"
)

func (a *App) builtInComponents() []Component {
	a.mu.RLock()
	debug := a.config.Debug
	logger := a.logger
	a.mu.RUnlock()

	var comps []Component

	if debug && !a.hasComponentType((*devtools.DevToolsComponent)(nil)) {
		comps = append(comps, newDevToolsComponent(a))
	}

	_ = logger // available for future built-in components
	return comps
}

func (a *App) hasComponentType(target any) bool {
	if target == nil {
		return false
	}

	typeOfTarget := reflect.TypeOf(target)
	if typeOfTarget == nil {
		return false
	}

	a.mu.RLock()
	comps := append([]Component{}, a.components...)
	a.mu.RUnlock()

	for _, c := range comps {
		if reflect.TypeOf(c) == typeOfTarget {
			return true
		}
	}

	return false
}
