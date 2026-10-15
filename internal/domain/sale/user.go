package sale

import (
	"github.com/Alps86/shop/internal/domain/value"
	"math/big"
)

func (user User) Validate(categories []value.CategoryID) error {
	if value.IdentifierSyntax(string(user.ID)).Validate() != nil {
		return value.NewError("InvalidInput", "UserID")
	}

	seen := map[value.CategoryID]bool{}
	for _, interest := range user.CategoryInterests {
		if err := interest.validate(categories, seen); err != nil {
			return err
		}

		seen[interest.CategoryID] = true
	}

	return nil
}

func (interest CategoryInterest) validate(categories []value.CategoryID, seen map[value.CategoryID]bool) error {
	weight, err := interest.Weight.Rational()
	if err != nil {
		return value.NewError("InvalidInput", "Weight")
	}

	if interest.invalidWeight(weight) {
		return value.NewError("InvalidInput", "Weight")
	}

	if seen[interest.CategoryID] {
		return value.NewError("InvalidInput", "CategoryID")
	}

	return interest.checkReference(categories)
}

func (interest CategoryInterest) invalidWeight(weight big.Rat) bool {
	return weight.Sign() < 0 || weight.Cmp(big.NewRat(1, 1)) > 0
}

func (interest CategoryInterest) checkReference(categories []value.CategoryID) error {
	for _, category := range categories {
		if interest.CategoryID == category {
			return nil
		}
	}

	return value.NewError("InvalidReference", "CategoryID")
}
