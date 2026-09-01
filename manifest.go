package odoomanifest

import (
	"fmt"

	"github.com/wimwenigerkind/odoo-manifest/internal/pylit"
)

type Manifest struct {
	Name                 string
	Version              string
	Summary              string
	Description          string
	Author               string
	Website              string
	License              string
	Category             string
	Depends              []string
	ExternalDependencies map[string][]string
	Application          bool
	Installable          bool
	AutoInstall          bool
}

func Parse(src []byte) (Manifest, error) {
	v, err := pylit.Parse(src)
	if err != nil {
		return Manifest{}, err
	}
	dict, ok := v.(map[string]any)
	if !ok {
		return Manifest{}, fmt.Errorf("manifest is not a dict")
	}

	m := Manifest{
		License:     "LGPL-3",
		Category:    "Uncategorized",
		Installable: true,
	}
	m.Name = asString(dict["name"])
	m.Version = asString(dict["version"])
	m.Summary = asString(dict["summary"])
	m.Description = asString(dict["description"])
	m.Author = asString(dict["author"])
	m.Website = asString(dict["website"])
	if raw, ok := dict["license"]; ok {
		m.License = asString(raw)
	}
	if raw, ok := dict["category"]; ok {
		m.Category = asString(raw)
	}
	m.Depends = asStringSlice(dict["depends"])
	m.ExternalDependencies = asStringSliceMap(dict["external_dependencies"])
	m.Application = asBool(dict["application"], false)
	if raw, ok := dict["installable"]; ok {
		m.Installable = asBool(raw, true)
	}
	m.AutoInstall = asAutoInstall(dict["auto_install"])
	return m, nil
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

func asBool(v any, def bool) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return def
}

func asStringSlice(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func asStringSliceMap(v any) map[string][]string {
	dict, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	out := make(map[string][]string, len(dict))
	for k, val := range dict {
		out[k] = asStringSlice(val)
	}
	return out
}

func asAutoInstall(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case []any:
		return true
	default:
		return false
	}
}
