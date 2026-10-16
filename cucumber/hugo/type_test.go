package hugo_test

import "github.com/Alps86/shop/internal/adapter/hugo"

type world struct {
	page hugo.Page
	err  error
}

type interruptedWriter struct {
	writes     int
	successful int
}
