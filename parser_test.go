package main

import (
	"fmt"
	"testing"
)

func TestParseAndExpandPlainString(t *testing.T) {
	var cases = []struct {
		Template  string
		Expansion map[string]string
		Want      string
	}{
		{"one", map[string]string{}, "one"},
		{"{x}", map[string]string{"x": "one"}, "one"},
		{"foo{x}bar", map[string]string{"x": "one"}, "fooonebar"},
		{"{x}{x}", map[string]string{"x": "one"}, "oneone"},
		{"{x:00d}", map[string]string{"x": "1"}, "1"},
		{"{x:01d}", map[string]string{"x": "1"}, "1"},
		{"{x:02d}", map[string]string{"x": "1"}, "01"},
		{"{x:02d}.{y:02d}", map[string]string{"x": "1", "y": "2"}, "01.02"},
		{"{foo}", map[string]string{"foo": "FOO"}, "FOO"},
		{"{f1}{f2}", map[string]string{"f1": "3", "f2": "4"}, "34"},
	}
	for _, tc := range cases {
		tmpl, err := ParseString(tc.Template)
		failWhenErr(t, err)
		ctx := Context{
			State: tc.Expansion,
		}
		x, err := tmpl.Expand(&ctx)
		failWhenErr(t, err)
		if x != tc.Want {
			fmt.Printf("Wanted '%s' but got '%s'\n", tc.Want, x)
		}
		failWhen(t, x != tc.Want)
	}
}
