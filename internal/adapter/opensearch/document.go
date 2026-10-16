package opensearch

import (
	"encoding/json"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"slices"
)

func (document Document) MarshalJSON() ([]byte, error) {
	document.Variants = slices.Clone(document.Variants)
	for index := range document.Variants {
		payload, err := json.Marshal(newStoredCard(document.Variants[index].Card))
		if err != nil {
			return nil, err
		}

		document.Variants[index].CardPayload = payload
	}

	return json.Marshal(documentJSON(document))
}

func newStoredCard(card application.ProductCard) storedSearchCard {
	return storedSearchCard{ProductCard: card, Compliance: newStoredCompliance(card.Compliance)}
}

func newStoredCompliance(compliance catalog.ProductCompliance) storedCardCompliance {
	return storedCardCompliance{ProductCompliance: compliance, Safety: compliance.Safety, ChemicalLabel: compliance.ChemicalLabel, EnergyLabel: compliance.EnergyLabel, NetContent: compliance.NetContent, UnitPriceBasis: compliance.UnitPriceBasis, Documents: compliance.Documents, Notices: compliance.Notices, RequiredFields: compliance.RequiredFields, NonApplicableReason: compliance.NonApplicableReason}
}
