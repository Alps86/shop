package opensearch

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

func (payload *cardPayload) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		return errors.New("card payload must be a JSON string")
	}

	var encoded string
	if err := json.Unmarshal(data, &encoded); err != nil {
		return err
	}

	decoded, err := decodeCardPayload(encoded)
	if err != nil {
		return err
	}

	*payload = decoded

	return nil
}

func decodeCardPayload(encoded string) ([]byte, error) {
	encoding := base64.RawStdEncoding.Strict()
	if strings.Contains(encoded, "=") {
		encoding = base64.StdEncoding.Strict()
	}

	decoded, err := encoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	if encoding.EncodeToString(decoded) != encoded {
		return nil, errors.New("card payload base64 is not canonical")
	}

	return decoded, nil
}
