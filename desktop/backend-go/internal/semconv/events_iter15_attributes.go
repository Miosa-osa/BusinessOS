package semconv

import "go.opentelemetry.io/otel/attribute"

// Event Routing attributes (iter15)
const (
	EventRoutingStrategyKey   = attribute.Key("event.routing.strategy")
	EventRoutingFilterCountKey = attribute.Key("event.routing.filter_count")
	EventSubscriberCountKey   = attribute.Key("event.subscriber.count")
)

func EventRoutingStrategy(val string) attribute.KeyValue {
	return EventRoutingStrategyKey.String(val)
}

func EventRoutingFilterCount(val int) attribute.KeyValue {
	return EventRoutingFilterCountKey.Int(val)
}

func EventSubscriberCount(val int) attribute.KeyValue {
	return EventSubscriberCountKey.Int(val)
}
