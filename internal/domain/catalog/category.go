package catalog

import "github.com/Alps86/shop/internal/domain/value"

func (category Category) Validate(catalog Catalog) error {
	if err := category.validateDefinition(); err != nil {
		return err
	}

	if err := category.validateParent(catalog); err != nil {
		return err
	}

	return category.validateAttributeCodes(catalog)
}

func (category Category) validateDefinition() error {
	if err := value.IdentifierSyntax(string(category.ID)).Validate(); err != nil {
		return err
	}

	if category.Version == 0 {
		return value.NewError("InvalidAttribute", "Version")
	}

	if len(category.FilterAttributes) > 4 {
		return value.NewError("InvalidAttribute", "FilterAttributes")
	}

	if len(category.TopAttributes) > 3 {
		return value.NewError("InvalidAttribute", "TopAttributes")
	}

	return nil
}

func (category Category) validateParent(catalog Catalog) error {
	seen := map[value.CategoryID]bool{category.ID: true}
	parent := category.ParentID
	for parent.Present {
		if seen[parent.Value] {
			return value.NewError("InvalidReference", "ParentID")
		}

		next, exists := catalog.Categories[parent.Value]
		if !exists {
			return value.NewError("InvalidReference", "ParentID")
		}

		seen[parent.Value] = true
		parent = next.ParentID
	}

	return nil
}

func (catalog Catalog) CanDeleteCategory(id value.CategoryID) error {
	if catalog.hasChild(id) {
		return value.NewError("Conflict", "CategoryID")
	}

	for _, product := range catalog.Products {
		if product.referencesCategory(id) {
			return value.NewError("Conflict", "CategoryID")
		}
	}

	return nil
}

func (catalog Catalog) hasChild(id value.CategoryID) bool {
	for _, category := range catalog.Categories {
		if category.hasParent(id) {
			return true
		}
	}

	return false
}

func (category Category) validateAttributeCodes(catalog Catalog) error {
	allowed := catalog.attributeCodes()
	for _, codes := range [][]AttributeCode{category.FilterAttributes, category.TopAttributes} {
		if err := category.validateCodes(codes, allowed); err != nil {
			return err
		}
	}

	return nil
}

func (catalog Catalog) attributeCodes() map[AttributeCode]bool {
	result := map[AttributeCode]bool{}
	for _, family := range catalog.Families {
		for _, definition := range family.Attributes {
			result[definition.Code] = true
		}
	}

	return result
}

func (category Category) validateCodes(codes []AttributeCode, allowed map[AttributeCode]bool) error {
	seen := map[AttributeCode]bool{}
	for _, code := range codes {
		if !allowed[code] {
			return value.NewError("InvalidAttribute", "Category.Attributes")
		}

		if seen[code] {
			return value.NewError("InvalidAttribute", "Category.Attributes")
		}

		seen[code] = true
	}

	return nil
}

func (category Category) hasParent(id value.CategoryID) bool {
	if !category.ParentID.Present {
		return false
	}

	return category.ParentID.Value == id
}
