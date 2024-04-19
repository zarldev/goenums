package generator_test

import (
	"os"
	"testing"

	"github.com/zarldev/goenums/pkg/generator"
)

func TestParseAndGenerateSimple(t *testing.T) {
	t.Log("TestParseAndGenerate")
	path := "testdata/validation/"
	filename := path + "status.go"
	err := generator.ParseAndGenerate(filename, true)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
	// check if the generated file exists
	filename = path + "statuses_enums.go"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("Expected file to exist, got %v", err)
	}
	// cleanup
	err = os.Remove(filename)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

func TestParseAndGenerateComplex(t *testing.T) {
	t.Log("TestParseAndGenerate")
	path := "testdata/solarsystem/"
	filename := path + "planets.go"
	err := generator.ParseAndGenerate(filename, false)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
	// check if the generated file exists
	filename = path + "planets_enums.go"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("Expected file to exist, got %v", err)
	}
	// cleanup
	err = os.Remove(filename)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

func TestParseAndGenerateCamelCase(t *testing.T) {
	t.Log("TestParseAndGenerate")
	path := "testdata/sale/"
	filename := path + "discount.go"
	err := generator.ParseAndGenerate(filename, false)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
	// check if the generated file exists
	filename = path + "discounttypes_enums.go"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("Expected file to exist, got %v", err)
	}
	// cleanup
	err = os.Remove(filename)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

func TestParseAndGenerateOnlyStrings(t *testing.T) {
	t.Log("TestParseAndGenerate")
	path := "testdata/planets/"
	filename := path + "planets.go"
	err := generator.ParseAndGenerate(filename, false)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
	// check if the generated file exists
	filename = path + "planets_enums.go"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("Expected file to exist, got %v", err)
	}
	// cleanup
	err = os.Remove(filename)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

func TestParseAndGenerateOnlyExtensions(t *testing.T) {
	t.Log("TestParseAndGenerate")
	path := "testdata/planets-extended/"
	filename := path + "planets.go"
	err := generator.ParseAndGenerate(filename, false)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
	// check if the generated file exists
	filename = path + "planets_enums.go"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("Expected file to exist, got %v", err)
	}
	// cleanup
	err = os.Remove(filename)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}
