package refcheck

import (
	"reflect"
	"testing"
)

func TestCollectHTML(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []Reference
	}{
		{
			name:     "url in text content is not a reference",
			content:  `<p>see https://example.com/page for details</p>`,
			expected: []Reference{},
		},
		{
			name:     "url in code element is not a reference",
			content:  `<code>&lt;a href="https://example.com"&gt;x&lt;/a&gt;</code>`,
			expected: []Reference{},
		},
		{
			name:     "url in comment is not a reference",
			content:  `<nav><!-- <a href="/reading-notes/">Reading Notes</a> --></nav>`,
			expected: []Reference{},
		},
		{
			name:    "script src can carry integrity",
			content: `<script src="https://cdn.example/x.js"></script>`,
			expected: []Reference{{
				File: "index.html", URL: "https://cdn.example/x.js", Role: "script src",
				Origin: OriginRemote, CanCarryIntegrity: true,
			}},
		},
		{
			name:    "stylesheet link can carry integrity and records both attributes",
			content: `<link rel="stylesheet" href="https://cdn.example/a.css" integrity="sha384-x" crossorigin="anonymous">`,
			expected: []Reference{{
				File: "index.html", URL: "https://cdn.example/a.css", Role: "link stylesheet href",
				Origin: OriginRemote, CanCarryIntegrity: true, HasIntegrity: true, HasCrossOrigin: true,
			}},
		},
		{
			name:    "icon link cannot carry integrity",
			content: `<link rel="icon" href="/favicon.ico">`,
			expected: []Reference{{
				File: "index.html", URL: "/favicon.ico", Role: "link icon href",
				Origin: OriginInternal,
			}},
		},
		{
			name:    "img src cannot carry integrity",
			content: `<img src="images/a.jpg">`,
			expected: []Reference{{
				File: "index.html", URL: "images/a.jpg", Role: "img src", Origin: OriginInternal,
			}},
		},
		{
			name:    "srcset yields one reference per candidate without descriptors",
			content: `<img srcset="a.jpg 1x, b.jpg 2x">`,
			expected: []Reference{
				{File: "index.html", URL: "a.jpg", Role: "img srcset", Origin: OriginInternal},
				{File: "index.html", URL: "b.jpg", Role: "img srcset", Origin: OriginInternal},
			},
		},
		{
			name:     "non-fetchable schemes and fragments are ignored",
			content:  `<a href="mailto:x@example.com">m</a><a href="tel:+1">t</a><a href="#top">f</a>`,
			expected: []Reference{},
		},
		{
			name:    "empty crossorigin is recorded as present",
			content: `<script src="https://cdn.example/x.js" integrity="sha384-x" crossorigin></script>`,
			expected: []Reference{{
				File: "index.html", URL: "https://cdn.example/x.js", Role: "script src",
				Origin: OriginRemote, CanCarryIntegrity: true, HasIntegrity: true, HasCrossOrigin: true,
			}},
		},
		{
			name:     "form action is not a reference, because it addresses a server route not a file",
			content:  `<form action="/subscribe"><input type="submit"></form>`,
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
