package value

import (
	"math/big"
	"regexp"
	"strings"
)

var decimalPattern = regexp.MustCompile(`^-?(0|[1-9][0-9]{0,11})(\.[0-9]{1,6})?$`)

func ParseDecimal(input string) (Decimal, error) {
	if !decimalPattern.MatchString(input) {
		return "", NewError("InvalidInput", "Decimal")
	}

	canonical := input
	if strings.Contains(canonical, ".") {
		canonical = strings.TrimRight(strings.TrimRight(canonical, "0"), ".")
	}

	if canonical == "-0" {
		canonical = "0"
	}

	return Decimal(canonical), nil
}

func (number Decimal) Rational() (big.Rat, error) {
	canonical, err := ParseDecimal(string(number))
	if err != nil {
		return big.Rat{}, err
	}

	result, _ := new(big.Rat).SetString(string(canonical))

	return *result, nil
}

func (number Decimal) Scaled() (int64, error) {
	rational, err := number.Rational()
	if err != nil {
		return 0, err
	}

	rational.Mul(&rational, big.NewRat(1000000, 1))

	return new(big.Int).Quo(rational.Num(), rational.Denom()).Int64(), nil
}

func (number Decimal) Compare(other Decimal) (int, error) {
	left, err := number.Rational()
	if err != nil {
		return 0, err
	}

	right, err := other.Rational()
	if err != nil {
		return 0, err
	}

	return left.Cmp(&right), nil
}
