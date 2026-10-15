package value

import (
	"bytes"
	"encoding/json"
	"reflect"
)

func (optional Optional[T]) Validate() error {
	if optional.Present {
		return nil
	}

	var zero T
	if !reflect.DeepEqual(optional.Value, zero) {
		return NewError("InvalidInput", "Optional.Value")
	}

	return nil
}

func (input RequiredJSON) Validate() error {
	if bytes.Equal(bytes.TrimSpace(input), []byte("null")) {
		return NewError("InvalidInput", "RequiredField")
	}

	if !json.Valid(input) {
		return NewError("InvalidInput", "JSON")
	}

	return nil
}

func ParseOptionalReplacement[T any](input []byte) (Optional[T], error) {
	if bytes.Equal(bytes.TrimSpace(input), []byte("null")) {
		return Optional[T]{}, nil
	}

	var result T
	if err := json.Unmarshal(input, &result); err != nil {
		return Optional[T]{}, NewError("InvalidInput", "JSON")
	}

	return Optional[T]{Value: result, Present: true}, nil
}
