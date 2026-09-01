package pylit

import (
	"reflect"
	"testing"
)

func TestParse_Scalars(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want any
	}{
		{"single quote", `'hi'`, "hi"},
		{"double quote", `"hi"`, "hi"},
		{"true", `True`, true},
		{"false", `False`, false},
		{"none", `None`, nil},
		{"int", `42`, int64(42)},
		{"negative int", `-7`, int64(-7)},
		{"float", `10.5`, 10.5},
		{"escaped newline", `'a\nb'`, "a\nb"},
		{"escaped tab", `'a\tb'`, "a\tb"},
		{"escaped quote", `'it\'s'`, "it's"},
		{"adjacent concat", `'a' 'b'`, "ab"},
		{"plus concat", `'a' + 'b'`, "ab"},
		{"unicode content", `'café'`, "café"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse([]byte(tt.src))
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tt.src, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse(%q) = %#v, want %#v", tt.src, got, tt.want)
			}
		})
	}
}

func TestParse_TripleQuoted(t *testing.T) {
	src := "\"\"\"multi\nline\n===\n\"\"\""
	got, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	want := "multi\nline\n===\n"
	if got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestParse_Collections(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want any
	}{
		{"list", `['a', 'b']`, []any{"a", "b"}},
		{"tuple", `('a', 'b')`, []any{"a", "b"}},
		{"trailing comma list", `['a',]`, []any{"a"}},
		{"single element tuple", `('a',)`, []any{"a"}},
		{"paren grouping is not tuple", `('a')`, "a"},
		{"empty list", `[]`, []any{}},
		{"nested", `{'python': ['ldap', 'requests']}`, map[string]any{"python": []any{"ldap", "requests"}}},
		{"dict trailing comma", `{'a': 1,}`, map[string]any{"a": int64(1)}},
		{"mixed dict", `{'name': 'X', 'ok': True, 'n': None}`, map[string]any{"name": "X", "ok": true, "n": nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse([]byte(tt.src))
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tt.src, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse(%q) = %#v, want %#v", tt.src, got, tt.want)
			}
		})
	}
}

func TestParse_SetAndPrefixedStrings(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want any
	}{
		{"set literal", `{'a', 'b'}`, []any{"a", "b"}},
		{"set trailing comma", `{'a',}`, []any{"a"}},
		{"empty dict", `{}`, map[string]any{}},
		{"raw string keeps backslash", `r'a\nb'`, `a\nb`},
		{"u prefix", `u'hi'`, "hi"},
		{"b prefix", `b'x'`, "x"},
		{"dict with set value", `{'web.assets': {'a/**/*'}}`, map[string]any{"web.assets": []any{"a/**/*"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse([]byte(tt.src))
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tt.src, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse(%q) = %#v, want %#v", tt.src, got, tt.want)
			}
		})
	}
}

func TestParse_CommentsAndCoding(t *testing.T) {
	src := "# -*- coding: utf-8 -*-\n# Part of Odoo.\n{'name': 'Base'}  # trailing comment\n"
	got, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	want := map[string]any{"name": "Base"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestParse_Errors(t *testing.T) {
	bad := []string{
		`get_version()`,
		`{'a': }`,
		`{`,
		`{'a': 1 'b': 2}`,
		`{name: 'x'}`,
		``,
		`{'a': 1} extra`,
	}
	for _, src := range bad {
		if _, err := Parse([]byte(src)); err == nil {
			t.Errorf("expected error for %q, got nil", src)
		}
	}
}
