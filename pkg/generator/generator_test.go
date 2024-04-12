package generator_test

import (
	"os"
	"testing"

	"github.com/zarldev/goenums/pkg/generator"
)

func TestParseAndGenerateSimple(t *testing.T) {
	t.Log("TestParseAndGenerate")
	filename := "testdata/validation/status.go"
	err := generator.ParseAndGenerate(filename, true)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
	// check if the generated file exists
	filename = "statuses_enums.go"
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
	filename := "testdata/solarsystem/planets.go"
	err := generator.ParseAndGenerate(filename, false)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
	// check if the generated file exists
	filename = "planets_enums.go"
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
	filename := "testdata/sale/discount.go"
	err := generator.ParseAndGenerate(filename, false)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
	// check if the generated file exists
	filename = "discounttypes_enums.go"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("Expected file to exist, got %v", err)
	}
	// cleanup
	err = os.Remove(filename)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}
