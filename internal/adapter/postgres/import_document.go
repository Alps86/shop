package postgres

import (
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
	"net/url"
)

func (manifest ImportManifest) validateDocuments(product catalog.Product) error {
	if err := manifest.validateComplianceDocuments(product.Compliance); err != nil {
		return err
	}

	for _, variant := range product.Variants {
		if !variant.Compliance.Present {
			continue
		}

		if err := manifest.validateComplianceDocuments(variant.Compliance.Value); err != nil {
			return err
		}
	}

	return nil
}

func (manifest ImportManifest) validateComplianceDocuments(compliance catalog.ProductCompliance) error {
	documents := append([]catalog.ProductDocument{}, compliance.Documents...)
	documents = append(documents, compliance.Safety.Instructions...)
	documents = append(documents, compliance.EnergyLabel.Value.LabelDocuments...)
	documents = append(documents, compliance.EnergyLabel.Value.InformationSheets...)
	for _, document := range documents {
		if err := manifest.validateDocumentURL(document.URL); err != nil {
			return err
		}
	}

	return nil
}

func (manifest ImportManifest) validateDocumentURL(address string) error {
	parsed, err := url.Parse(address)
	if err != nil {
		return value.NewError("InvalidInput", "Document.URL")
	}

	if parsed.Scheme != "https" {
		return value.NewError("InvalidInput", "Document.URL")
	}

	if parsed.Hostname() == "" {
		return value.NewError("InvalidInput", "Document.URL")
	}

	if parsed.User != nil {
		return value.NewError("InvalidInput", "Document.URL")
	}

	return nil
}
