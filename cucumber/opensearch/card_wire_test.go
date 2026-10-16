package opensearch_test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/Alps86/shop/internal/adapter/opensearch"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"slices"
	"strconv"
	"strings"
)

func (current *world) cardWire(table *godog.Table) error {
	data := read(table)[0]
	payload, err := current.codec.remainderPayload(data["Remainder"])
	if err != nil {
		return err
	}

	encoding, exists := map[string]*base64.Encoding{"standard": base64.StdEncoding, "raw": base64.RawStdEncoding}[data["Format"]]
	if !exists {
		return fmt.Errorf("unbekanntes Base64format")
	}

	encoded, err := mutateWire(encoding.EncodeToString(payload), data["Mutation"])
	current.codec.wirePayload = value.Optional[string]{Present: true, Value: encoded}

	return err
}

func (probe cardCodecProbe) remainderPayload(raw string) ([]byte, error) {
	remainder, err := strconv.Atoi(raw)
	if err != nil {
		return nil, err
	}

	if !slices.Contains([]int{0, 1, 2}, remainder) {
		return nil, fmt.Errorf("ungültige Payloadrestklasse")
	}

	payload := bytes.Clone(probe.document.Variants[0].CardPayload)
	for len(payload)%3 != remainder {
		payload = append(payload, ' ')
	}

	return payload, nil
}

func mutateWire(encoded, mutation string) (string, error) {
	if mutation == "none" {
		return encoded, nil
	}

	if mutation == "restbits" {
		return noncanonicalWire(encoded), nil
	}

	control, exists := map[string]string{"cr": "\r", "lf": "\n", "crlf": "\r\n"}[mutation]
	if exists {
		return encoded[:4] + control + encoded[4:], nil
	}

	return "", fmt.Errorf("unbekannte Base64mutation")
}

func noncanonicalWire(encoded string) string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	raw := strings.TrimRight(encoded, "=")
	index := strings.IndexByte(alphabet, raw[len(raw)-1])

	return raw[:len(raw)-1] + string(alphabet[index+1]) + encoded[len(raw):]
}

func (current *world) searchWireCard(table *godog.Table) error {
	current.msearch.response = current.codec.response()
	current.adapter = opensearch.New(current.config)

	return current.runSearch(read(table)[0])
}
