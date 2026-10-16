package opensearch

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/Alps86/shop/internal/domain/value"
)

func (adapter Adapter) searchRequest(body JSON) searchRequest {
	suggest, exists := body["suggest"]
	if !exists {
		return searchRequest{Path: "/" + adapter.config.Index + "/_search?request_cache=true", Payload: body}
	}

	delete(body, "suggest")
	completion := JSON{"size": 0, "_source": false, "stored_fields": []string{"_none_"}, "query": JSON{"match_all": JSON{}}, "suggest": suggest}
	lines := multiSearchLines{{"request_cache": true}, body, {"request_cache": true}, completion}

	return searchRequest{Path: "/" + adapter.config.Index + "/_msearch", Payload: lines, Multi: true}
}

func encodePayload(payload any) ([]byte, error) {
	lines, ok := payload.(multiSearchLines)
	if ok {
		return lines.encode()
	}

	return json.Marshal(payload)
}

func (lines multiSearchLines) encode() ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	failures := []error{}
	for _, line := range lines {
		failures = append(failures, encoder.Encode(line))
	}

	return buffer.Bytes(), errors.Join(failures...)
}

func decodeSearchResponse(body []byte, multi bool) (SearchResponse, error) {
	if multi {
		return decodeMultiSearch(body)
	}

	var response SearchResponse
	err := json.Unmarshal(body, &response)

	return response, err
}

func decodeMultiSearch(body []byte) (SearchResponse, error) {
	var response multiSearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return SearchResponse{}, err
	}

	if len(response.Responses) != 2 {
		return SearchResponse{}, value.NewError("SearchUnavailable", "Responses")
	}

	return response.result()
}

func (response multiSearchResponse) result() (SearchResponse, error) {
	for _, part := range response.Responses {
		if err := part.validate(); err != nil {
			return SearchResponse{}, err
		}
	}

	main := response.Responses[0].SearchResponse
	main.Suggest = response.Responses[1].Suggest

	return main, nil
}

func (part multiSearchPart) validate() error {
	if part.Status != 200 {
		return value.NewError("SearchUnavailable", "PartStatus")
	}

	if len(part.Error) > 0 {
		return value.NewError("SearchUnavailable", "PartError")
	}

	if part.TimedOut {
		return value.NewError("SearchUnavailable", "PartTimeout")
	}

	if part.Shards.Failed > 0 {
		return value.NewError("SearchUnavailable", "PartShards")
	}

	return nil
}
