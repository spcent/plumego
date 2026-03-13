package core

func (a *App) declaredComponents() []Component {
	return nil
}

func (a *App) hasComponentType(target any) bool {
	return hasComponentType(nil, target)
}

func hasComponentType(components []Component, target any) bool {
	return false
}
