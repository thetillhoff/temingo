package refcheck

import (
	"reflect"
	"testing"
)

func TestCollectCSS(t *testing.T) {
	tests := []struct {
		name     string
		css      string
		expected []Reference
	}{
		{
			name:     "no urls",
			css:      `.a { color: red; }`,
			expected: []Reference{},
		},
		{
			name: "background image, quoted and unquoted",
			css:  `.a{background:url("images/a.jpg")}.b{background:url(images/b.jpg)}`,
			expected: []Reference{
				{File: "style.css", URL: "images/a.jpg", Role: "css url()", Origin: OriginInternal},
				{File: "style.css", URL: "images/b.jpg", Role: "css url()", Origin: OriginInternal},
			},
		},
		{
			name: "font face src is a url like any other",
			css:  `@font-face{src:url('https://cdn.example/f.woff2')}`,
			expected: []Reference{
				{File: "style.css", URL: "https://cdn.example/f.woff2", Role: "css url()", Origin: OriginRemote},
			},
		},
		{
			name: "import with and without url()",
			css:  `@import url("https://cdn.example/a.css");@import "b.css";`,
			expected: []Reference{
				{File: "style.css", URL: "https://cdn.example/a.css", Role: "css @import", Origin: OriginRemote},
				{File: "style.css", URL: "b.css", Role: "css @import", Origin: OriginInternal},
			},
		},
		{
			name:     "data uri is ignored",
			css:      `.a{background:url(data:image/gif;base64,R0lGOD)}`,
			expected: []Reference{},
		},
		{
			name:     "css references never carry integrity",
			css:      `@import url("https://cdn.example/a.css");`,
			expected: []Reference{{File: "style.css", URL: "https://cdn.example/a.css", Role: "css @import", Origin: OriginRemote}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := CollectCSS("style.css", []byte(test.css))
			if !reflect.DeepEqual(got, test.expected) {
				t.Errorf("CollectCSS() = %+v, want %+v", got, test.expected)
			}
		})
	}
}

func TestCollectHTMLIncludesStyles(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []Reference
	}{
		{
			name:    "style block",
			content: `<style>.a{background:url(images/a.jpg)}</style>`,
			expected: []Reference{
				{File: "index.html", URL: "images/a.jpg", Role: "css url()", Origin: OriginInternal},
			},
		},
		{
			name:    "style attribute",
			content: `<div style="background:url(images/b.jpg)"></div>`,
			expected: []Reference{
				{File: "index.html", URL: "images/b.jpg", Role: "css url()", Origin: OriginInternal},
			},
		},
		{
			name:     "css in a code element is still not a reference",
			content:  `<code>.a{background:url(images/c.jpg)}</code>`,
			expected: []Reference{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := CollectHTML("index.html", []byte(test.content))
			if !reflect.DeepEqual(got, test.expected) {
				t.Errorf("CollectHTML() = %+v, want %+v", got, test.expected)
			}
		})
	}
}
