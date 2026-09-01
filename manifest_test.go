package odoomanifest

import (
	"reflect"
	"testing"
)

func TestParse_FullManifest(t *testing.T) {
	src := []byte(`# -*- coding: utf-8 -*-
# Part of Odoo. See LICENSE file for full copyright and licensing details.
{
    'name': 'test-uninstall',
    'version': '0.1',
    'category': 'Hidden/Tests',
    'description': """A module to test the uninstall feature.""",
    'depends': ['base'],
    'data': ['ir.model.access.csv'],
    'installable': True,
    'author': 'Odoo S.A.',
    'license': 'LGPL-3',
}
`)
	m, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if m.Name != "test-uninstall" {
		t.Errorf("Name = %q", m.Name)
	}
	if m.Version != "0.1" {
		t.Errorf("Version = %q", m.Version)
	}
	if m.Category != "Hidden/Tests" {
		t.Errorf("Category = %q", m.Category)
	}
	if m.Description != "A module to test the uninstall feature." {
		t.Errorf("Description = %q", m.Description)
	}
	if !reflect.DeepEqual(m.Depends, []string{"base"}) {
		t.Errorf("Depends = %#v", m.Depends)
	}
	if !m.Installable {
		t.Errorf("Installable = false, want true")
	}
	if m.Author != "Odoo S.A." {
		t.Errorf("Author = %q", m.Author)
	}
	if m.License != "LGPL-3" {
		t.Errorf("License = %q", m.License)
	}
}

func TestParse_Defaults(t *testing.T) {
	m, err := Parse([]byte(`{'name': 'X', 'version': '19.0.1.0.0'}`))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if m.License != "LGPL-3" {
		t.Errorf("License default = %q, want LGPL-3", m.License)
	}
	if m.Category != "Uncategorized" {
		t.Errorf("Category default = %q, want Uncategorized", m.Category)
	}
	if !m.Installable {
		t.Errorf("Installable default = false, want true")
	}
	if m.Application {
		t.Errorf("Application default = true, want false")
	}
	if m.AutoInstall {
		t.Errorf("AutoInstall default = true, want false")
	}
	if len(m.Depends) != 0 {
		t.Errorf("Depends default = %#v, want empty", m.Depends)
	}
}

func TestParse_ExternalDependencies(t *testing.T) {
	m, err := Parse([]byte(`{'name': 'X', 'external_dependencies': {'python': ['ldap'], 'bin': ['wkhtmltopdf']}}`))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	want := map[string][]string{"python": {"ldap"}, "bin": {"wkhtmltopdf"}}
	if !reflect.DeepEqual(m.ExternalDependencies, want) {
		t.Errorf("ExternalDependencies = %#v, want %#v", m.ExternalDependencies, want)
	}
}

func TestParse_AutoInstall(t *testing.T) {
	list, err := Parse([]byte(`{'name': 'X', 'auto_install': ['sale', 'stock']}`))
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !list.AutoInstall {
		t.Errorf("auto_install list should mean AutoInstall=true")
	}
	b, err := Parse([]byte(`{'name': 'X', 'auto_install': True}`))
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !b.AutoInstall {
		t.Errorf("auto_install True should mean AutoInstall=true")
	}
}

func TestParse_Summary(t *testing.T) {
	m, err := Parse([]byte(`{'name': 'X', 'summary': 'short desc', 'website': 'https://example.com'}`))
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if m.Summary != "short desc" {
		t.Errorf("Summary = %q", m.Summary)
	}
	if m.Website != "https://example.com" {
		t.Errorf("Website = %q", m.Website)
	}
}

func TestParse_NotADict(t *testing.T) {
	if _, err := Parse([]byte(`['not', 'a', 'dict']`)); err == nil {
		t.Errorf("expected error for non-dict manifest")
	}
}
