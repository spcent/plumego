package pubsubdebug

import "github.com/spcent/plumego/pubsub"

// PubSubConfig configures the optional pubsub snapshot endpoint.
type PubSubConfig struct {
	Enabled bool          // Whether to enable the debug endpoint
	Path    string        // Path for the debug endpoint
	Pub     pubsub.PubSub // PubSub instance for snapshotting
}
