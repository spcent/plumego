package core

import "reflect"

func (a *App) declaredComponents() []Component {
	a.mu.RLock()
	comps := append([]Component{}, a.components...)
	a.mu.RUnlock()

	return filterNilComponents(comps)
}

func (a *App) hasComponentType(target any) bool {
	a.mu.RLock()
	comps := append([]Component{}, a.components...)
	a.mu.RUnlock()
	return hasComponentType(comps, target)
}

func hasComponentType(components []Component, target any) bool {
	if target == nil {
		return false
	}

	typeOfTarget := reflect.TypeOf(target)
	if typeOfTarget == nil {
		return false
	}

	for _, c := range components {
		if reflect.TypeOf(c) == typeOfTarget {
			return true
		}
	}

	return false
}
