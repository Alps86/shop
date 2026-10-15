package event

import (
	"github.com/Alps86/shop/internal/domain/value"
	"reflect"
)

func (envelopes ProjectionEnvelopes) Unique() (ProjectionEnvelopes, error) {
	registry := projectionRegistry{}
	result := ProjectionEnvelopes{}
	for _, envelope := range envelopes {
		duplicate, err := registry.register(envelope)
		if err != nil {
			return nil, err
		}

		if duplicate {
			continue
		}

		result = append(result, envelope)
	}

	return result, nil
}

func (registry projectionRegistry) register(envelope ProjectionEnvelope) (bool, error) {
	previous, exists := registry[envelope.EventID]
	if !exists {
		registry[envelope.EventID] = envelope
		return false, nil
	}

	if !reflect.DeepEqual(previous, envelope) {
		return true, value.NewError("Conflict", "EventID")
	}

	return true, nil
}
