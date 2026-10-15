package event

import "github.com/Alps86/shop/internal/domain/value"

func (envelope ProjectionEnvelope) Validate() error {
	if envelope.SchemaVersion != 1 {
		return value.NewError("InvalidInput", "SchemaVersion")
	}

	if err := value.IdentifierSyntax(envelope.EventID).Validate(); err != nil {
		return err
	}

	if envelope.OccurredAt.IsZero() {
		return value.NewError("InvalidInput", "OccurredAt")
	}

	if err := envelope.Context.Validate(); err != nil {
		return err
	}

	return envelope.validateProjection()
}

func (envelope ProjectionEnvelope) validateProjection() error {
	if envelope.Projection.Revision == 0 {
		return value.NewError("InvalidInput", "Revision")
	}

	if err := value.IdentifierSyntax(string(envelope.Projection.ProductID)).Validate(); err != nil {
		return err
	}

	if err := envelope.Projection.Product.Validate(); err != nil {
		return err
	}

	return envelope.validateKind()
}

func (envelope ProjectionEnvelope) validateKind() error {
	if envelope.Kind == "upsert" {
		return envelope.validateUpsert()
	}

	if envelope.Kind == "delete" {
		return envelope.validateDelete()
	}

	return value.NewError("InvalidInput", "Kind")
}

func (envelope ProjectionEnvelope) validateUpsert() error {
	if !envelope.Projection.Product.Present {
		return value.NewError("InvalidInput", "Product")
	}

	if envelope.Projection.Product.Value.ID != envelope.Projection.ProductID {
		return value.NewError("InvalidInput", "ProductID")
	}

	return nil
}

func (envelope ProjectionEnvelope) validateDelete() error {
	if envelope.Projection.Product.Present {
		return value.NewError("InvalidInput", "Product")
	}

	if len(envelope.Projection.Inventory) > 0 {
		return value.NewError("InvalidInput", "Inventory")
	}

	return nil
}
