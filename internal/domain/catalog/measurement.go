package catalog

import (
	"github.com/Alps86/shop/internal/domain/value"
	"math/big"
)

var unitDimensions = map[string]string{"mm": "length", "cm": "length", "m": "length", "g": "mass", "kg": "mass", "ml": "volume", "l": "volume", "GB": "storage", "TB": "storage", "Wh": "energy", "h": "time", "rpm": "speed"}
var unitFactors = map[string]int64{"mm": 1, "cm": 10, "m": 1000, "g": 1, "kg": 1000, "ml": 1, "l": 1000, "GB": 1, "TB": 1000, "Wh": 1, "h": 1, "rpm": 1}

func (measurement Measurement) Convert(target string) (Measurement, error) {
	if err := measurement.compatible(target); err != nil {
		return Measurement{}, err
	}

	number, err := measurement.Value.Rational()
	if err != nil {
		return Measurement{}, err
	}

	if number.Sign() <= 0 {
		return Measurement{}, value.NewError("InvalidAttribute", "Measurement")
	}

	number.Mul(&number, big.NewRat(unitFactors[measurement.Unit], unitFactors[target]))

	return newMeasurement(&number, target)
}

func (measurement Measurement) compatible(target string) error {
	dimension, exists := unitDimensions[measurement.Unit]
	if !exists {
		return value.NewError("InvalidUnit", "Unit")
	}

	if unitDimensions[target] != dimension {
		return value.NewError("InvalidUnit", "TargetUnit")
	}

	return nil
}

func (measurement Measurement) UnitPrice(price value.Money, basis Measurement) (value.Money, error) {
	converted, err := measurement.Convert(basis.Unit)
	if err != nil {
		return value.Money{}, err
	}

	amount, err := basis.Convert(basis.Unit)
	if err != nil {
		return value.Money{}, err
	}

	return converted.priceRatio(price, amount)
}

func (measurement Measurement) priceRatio(price value.Money, basis Measurement) (value.Money, error) {
	if err := price.Validate(); err != nil {
		return value.Money{}, err
	}

	denominator := measurement.validatedValue()
	numerator := basis.validatedValue()
	numerator.Quo(numerator, denominator)
	numerator.Mul(numerator, big.NewRat(price.MinorUnits, 1))
	numerator.Add(numerator, big.NewRat(1, 2))
	rounded := new(big.Int).Quo(numerator.Num(), numerator.Denom())
	if !rounded.IsInt64() {
		return value.Money{}, value.NewError("AmountOverflow", "UnitPrice")
	}

	return value.NewMoney(rounded.Int64(), price.Currency)
}

// Callers have validated both values through ParseDecimal before exact arithmetic.
func (measurement Measurement) validatedValue() *big.Rat {
	result, _ := new(big.Rat).SetString(string(measurement.Value))
	return result
}

func newMeasurement(number *big.Rat, target string) (Measurement, error) {
	canonical, err := value.ParseDecimal(number.FloatString(6))
	if err != nil {
		return Measurement{}, err
	}

	if number.Cmp((Measurement{Value: canonical}).validatedValue()) != 0 {
		return Measurement{}, value.NewError("InvalidAttribute", "Measurement.Precision")
	}

	return Measurement{Value: canonical, Unit: target}, nil
}
