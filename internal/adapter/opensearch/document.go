package opensearch

import (
	"encoding/json"
	"slices"
)

func (document Document) MarshalJSON() ([]byte, error) {
	document.Variants = slices.Clone(document.Variants)
	for index := range document.Variants {
		payload, err := json.Marshal(document.Variants[index].Card)
		if err != nil {
			return nil, err
		}

		document.Variants[index].CardPayload = payload
	}

	return json.Marshal(documentJSON(document))
}
