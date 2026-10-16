package demo_test

import (
	"golang.org/x/net/html"
	"strings"
)

func (response response) diagnostic() string {
	tokenizer := html.NewTokenizer(strings.NewReader(response.body))
	var output strings.Builder
	for {
		kind := tokenizer.Next()
		if kind == html.ErrorToken {
			return strings.Join(strings.Fields(output.String()), " ")
		}

		if kind == html.TextToken {
			output.Write(tokenizer.Text())
			output.WriteByte(' ')
		}
	}
}
