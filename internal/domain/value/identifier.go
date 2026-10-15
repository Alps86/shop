package value

import "regexp"

var identifierPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`)

func (input IdentifierSyntax) Validate() error {
	if !identifierPattern.MatchString(string(input)) {
		return NewError("InvalidInput", "ID")
	}

	return nil
}
