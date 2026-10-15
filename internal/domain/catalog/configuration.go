package catalog

import "github.com/Alps86/shop/internal/domain/value"

func (family ProductFamily) Validate() error {
	if err := value.IdentifierSyntax(string(family.ID)).Validate(); err != nil {
		return err
	}

	if family.Version == 0 {
		return value.NewError("InvalidAttribute", "Version")
	}

	return family.validateDefinitions()
}

func (family ProductFamily) validateDefinitions() error {
	seen := map[AttributeCode]bool{}
	for _, definition := range family.Attributes {
		if seen[definition.Code] {
			return value.NewError("InvalidAttribute", "Code")
		}

		if err := definition.Validate(); err != nil {
			return err
		}

		seen[definition.Code] = true
	}

	return nil
}

func (attributes Attributes) Validate(family ProductFamily, level AttributeLevel) error {
	definitions := map[AttributeCode]AttributeDefinition{}
	for _, definition := range family.Attributes {
		definitions[definition.Code] = definition
		if err := attributes.checkRequired(definition, level); err != nil {
			return err
		}
	}

	for code, attribute := range attributes {
		definition, exists := definitions[code]
		if !exists {
			return value.NewError("InvalidAttribute", "Code")
		}

		if _, err := definition.ValidateValue(attribute, level); err != nil {
			return err
		}
	}

	return nil
}

func (attributes Attributes) checkRequired(definition AttributeDefinition, level AttributeLevel) error {
	if definition.Level != level {
		return nil
	}

	if !definition.Required {
		return nil
	}

	if _, exists := attributes[definition.Code]; !exists {
		return value.NewError("InvalidAttribute", string(definition.Code))
	}

	return nil
}

func (product Product) referencesCategory(id value.CategoryID) bool {
	if product.PrimaryCategoryID == id {
		return true
	}

	for _, additional := range product.AdditionalCategoryIDs {
		if additional == id {
			return true
		}
	}

	return false
}
