package opensearch

import (
	"github.com/Alps86/shop/internal/application"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func completionGroup(field, text string) JSON {
	fuzziness := 1
	if utf8.RuneCountInString(text) >= 8 {
		fuzziness = 2
	}

	completion := JSON{"field": field, "size": 25, "skip_duplicates": true, "contexts": JSON{"visibility": []string{"public"}}, "fuzzy": JSON{"fuzziness": fuzziness, "min_length": 4, "prefix_length": 2, "transpositions": true}}

	return JSON{"prefix": text, "completion": completion}
}

func wordTokens(text string) []string {
	return strings.FieldsFunc(text, func(character rune) bool {
		return !unicode.IsLetter(character) && !unicode.IsNumber(character)
	})
}

func suggestionWords(names []string) []string {
	words := []string{}
	for _, name := range names {
		words = append(words, wordTokens(name)...)
	}

	slices.Sort(words)

	return slices.Compact(words)
}

func addWordGroups(groups JSON, text string) {
	for index, word := range wordTokens(text) {
		if utf8.RuneCountInString(word) < 4 {
			continue
		}

		groups["word"+strconv.Itoa(index)] = completionGroup("words", word)
	}
}

func (response SearchResponse) wordCorrection(text string) []application.Suggestion {
	corrected := response.correctedQuery(text)
	if corrected == text {
		return []application.Suggestion{}
	}

	return []application.Suggestion{{Text: corrected, Correction: true}}
}

func (response SearchResponse) correctedQuery(text string) string {
	var corrected strings.Builder
	offset := 0
	for index, word := range wordTokens(text) {
		position := offset + strings.Index(text[offset:], word)
		corrected.WriteString(text[offset:position])
		corrected.WriteString(response.correctedWord(word, index))
		offset = position + len(word)
	}

	corrected.WriteString(text[offset:])

	return corrected.String()
}

func (response SearchResponse) correctedWord(word string, index int) string {
	set := SuggestionSet{Values: []application.Suggestion{}, Seen: map[string]bool{}}
	for _, group := range response.Suggest["word"+strconv.Itoa(index)] {
		set.appendGroup(group, word)
	}

	if set.containsWord(word) {
		return word
	}

	if len(set.Values) == 0 {
		return word
	}

	slices.SortFunc(set.Values, response.compareSuggestions)

	return set.Values[0].Text
}

func (set SuggestionSet) containsWord(word string) bool {
	for _, item := range set.Values {
		if strings.EqualFold(item.Text, word) {
			return true
		}
	}

	return false
}
