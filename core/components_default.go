package core

import "reflect"

func (a *App) builtInComponents() []Component {
	var comps []Component

	a.mu.RLock()
	pubSubConfig := a.config.PubSub
	webhookOutConfig := a.config.WebhookOut
	webhookInConfig := a.config.WebhookIn
	pub := a.pub
	logger := a.logger
	a.mu.RUnlock()

	if pubSubConfig.Enabled && !a.hasComponentType((*pubSubDebugComponent)(nil)) {
		comps = append(comps, newPubSubDebugComponent(pubSubConfig, pub))
	}

	if webhookOutConfig.Enabled && !a.hasComponentType((*webhookOutComponent)(nil)) {
		comps = append(comps, newWebhookOutComponent(webhookOutConfig))
	}

	if webhookInConfig.Enabled && !a.hasComponentType((*webhookInComponent)(nil)) {
		comps = append(comps, newWebhookInComponent(webhookInConfig, pub, logger))
	}

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

	for _, c := range a.components {
		if reflect.TypeOf(c) == typeOfTarget {
			return true
		}
	}

	return false
}
