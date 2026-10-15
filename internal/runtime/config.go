package runtime

import (
	"github.com/Alps86/shop/internal/domain/value"
	"net/url"
	"os"
)

func NewEnvironmentConfig() Config {
	return Config{
		DSN:               os.Getenv("SHOP_POSTGRES_DSN"),
		Schema:            os.Getenv("SHOP_SCHEMA"),
		TemporalAddress:   os.Getenv("SHOP_TEMPORAL_ADDRESS"),
		TemporalNamespace: os.Getenv("SHOP_TEMPORAL_NAMESPACE"),
		TaskQueue:         os.Getenv("SHOP_TASK_QUEUE"),
		SearchEndpoint:    os.Getenv("SHOP_OPENSEARCH_URL"),
		SearchIndex:       os.Getenv("SHOP_SEARCH_INDEX"),
		Publication:       os.Getenv("SHOP_CDC_PUBLICATION"),
		Slot:              os.Getenv("SHOP_CDC_SLOT"),
		Listen:            os.Getenv("SHOP_LISTEN"),
		SeedFile:          os.Getenv("SHOP_SEED_FILE"),
		Context:           value.RequestContext{Locale: "de-DE", Currency: "EUR", Market: "DE"},
	}
}

func (config Config) Validate() error {
	if err := config.validateIdentifiers(); err != nil {
		return err
	}

	if err := config.validateConnections(); err != nil {
		return err
	}

	return config.Context.Validate()
}

func (config Config) validateIdentifiers() error {
	for _, input := range []string{config.Schema, config.SearchIndex, config.Publication, config.Slot} {
		if err := value.IdentifierSyntax(input).Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (config Config) validateConnections() error {
	for _, input := range []string{config.DSN, config.TemporalAddress, config.TemporalNamespace, config.TaskQueue, config.Listen} {
		if input == "" {
			return value.NewError("InvalidInput", "RuntimeConfiguration")
		}
	}

	return config.validateSearchEndpoint()
}

func (config Config) validateSearchEndpoint() error {
	endpoint, err := url.ParseRequestURI(config.SearchEndpoint)
	if err != nil {
		return value.NewError("InvalidInput", "SearchEndpoint")
	}

	if !config.httpEndpoint(endpoint) {
		return value.NewError("InvalidInput", "SearchEndpoint")
	}

	return nil
}

func (config Config) httpEndpoint(endpoint *url.URL) bool {
	return endpoint.Host != "" && (endpoint.Scheme == "http" || endpoint.Scheme == "https")
}
