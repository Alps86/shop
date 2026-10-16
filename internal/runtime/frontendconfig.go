package runtime

import (
	"github.com/Alps86/shop/internal/domain/value"
	"net"
	"net/url"
)

func (config FrontendConfig) Validate() error {
	origin, err := url.ParseRequestURI(config.PublicOrigin)
	if err != nil {
		return value.NewError("InvalidInput", "PublicOrigin")
	}

	if !config.validOrigin(origin) {
		return value.NewError("InvalidInput", "PublicOrigin")
	}

	return config.validateDirectories()
}

func (config FrontendConfig) validateDirectories() error {
	if config.ContentDirectory == "" {
		return value.NewError("InvalidInput", "ContentDirectory")
	}

	_, _, err := net.SplitHostPort(config.BackendListen)
	if err != nil {
		return value.NewError("InvalidInput", "BackendListen")
	}

	return nil
}

func (config FrontendConfig) validOrigin(origin *url.URL) bool {
	return config.originEndpoint(origin) && config.originMetadata(origin)
}

func (config FrontendConfig) originEndpoint(origin *url.URL) bool {
	return origin.Host != "" && origin.Path == "" && (origin.Scheme == "http" || origin.Scheme == "https")
}

func (config FrontendConfig) originMetadata(origin *url.URL) bool {
	return origin.RawQuery == "" && origin.Fragment == "" && origin.User == nil
}
