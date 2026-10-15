package value_test

import "github.com/Alps86/shop/internal/domain/value"

type world struct {
	err       error
	canonical string
	context   value.RequestContext
}
