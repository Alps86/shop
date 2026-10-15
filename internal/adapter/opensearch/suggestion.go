package opensearch

import (
	"github.com/Alps86/shop/internal/application"
	"slices"
	"strings"
	"unicode/utf8"
)

func (adapter Adapter) suggestBody(body JSON, query application.SearchQuery) {
	text := strings.TrimSpace(query.Text)
	if utf8.RuneCountInString(text) < 2 {
		return
	}

	fuzziness := 1
	if utf8.RuneCountInString(text) >= 8 {
		fuzziness = 2
	}

	completion := JSON{"field": "suggest", "size": 25, "skip_duplicates": true, "contexts": JSON{"visibility": []string{"public"}}, "fuzzy": JSON{"fuzziness": fuzziness, "min_length": 4, "prefix_length": 2, "transpositions": true}}
	body["suggest"] = JSON{"catalog": JSON{"prefix": text, "completion": completion}}
}

func (response SearchResponse) suggestions(query application.SearchQuery) []application.Suggestion {
	set := SuggestionSet{Values: []application.Suggestion{}, Seen: map[string]bool{}}
	for _, group := range response.Suggest["catalog"] {
		set.appendGroup(group, query.Text)
	}

	slices.SortFunc(set.Values, response.compareSuggestions)

	return set.Values[:min(5, len(set.Values))]
}

func (set *SuggestionSet) appendGroup(group SuggestionGroup, query string) {
	for _, option := range group.Options {
		if set.Seen[option.Text] {
			continue
		}

		set.Seen[option.Text] = true
		corrected := !strings.HasPrefix(strings.ToLower(option.Text), strings.ToLower(query))
		set.Values = append(set.Values, application.Suggestion{Text: option.Text, Correction: corrected})
	}
}

func (response SearchResponse) compareSuggestions(left, right application.Suggestion) int {
	if left.Correction != right.Correction {
		if left.Correction {
			return 1
		}

		return -1
	}

	return strings.Compare(left.Text, right.Text)
}
