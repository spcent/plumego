package core

import "reflect"

func (a *App) builtInComponents() []Component {
	var comps []Component

	if a.config.PubSub.Enabled && !a.hasComponentType((*pubSubDebugComponent)(nil)) {
		comps = append(comps, newPubSubDebugComponent(a.config.PubSub, a.pub))
	}

	if a.config.WebhookOut.Enabled && !a.hasComponentType((*webhookOutComponent)(nil)) {
		comps = append(comps, newWebhookOutComponent(a.config.WebhookOut))
	}

	if a.config.WebhookIn.Enabled && !a.hasComponentType((*webhookInComponent)(nil)) {
		comps = append(comps, newWebhookInComponent(a.config.WebhookIn, a.pub, a.logger))
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
