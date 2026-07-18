package testdata

import (
	"encoding"
	"encoding/json"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	aliases "github.com/zarldev/goenums/internal/testdata/aliases"
	attributes "github.com/zarldev/goenums/internal/testdata/attributes"
	insensitive "github.com/zarldev/goenums/internal/testdata/insensitive"
	invalid "github.com/zarldev/goenums/internal/testdata/invalid"
	invalidalias "github.com/zarldev/goenums/internal/testdata/invalid_alias"
	multiple "github.com/zarldev/goenums/internal/testdata/multiple"
	crypto "github.com/zarldev/goenums/internal/testdata/negative"
	plural "github.com/zarldev/goenums/internal/testdata/plural"
	tickets "github.com/zarldev/goenums/internal/testdata/quotes"
	skipvalues "github.com/zarldev/goenums/internal/testdata/skipvalues"
	timesale "github.com/zarldev/goenums/internal/testdata/time"
	typecomment "github.com/zarldev/goenums/internal/testdata/typecomment"
	values "github.com/zarldev/goenums/internal/testdata/values"
	valuesonly "github.com/zarldev/goenums/internal/testdata/values_only"
)

func TestGeneratedEnumFilesParse(t *testing.T) {
	t.Parallel()
	for _, filePath := range expectedGeneratedFiles() {
		filePath := filePath
		t.Run(filePath, func(t *testing.T) {
			t.Parallel()
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("read generated file: %v", err)
			}
			fset := token.NewFileSet()
			if _, err := parser.ParseFile(fset, filePath, content, parser.ParseComments); err != nil {
				t.Fatalf("parse generated file: %v", err)
			}
		})
	}
}

func TestGeneratedEnumInterfaces(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value any
	}{
		{"aliases order", aliases.Orders.CREATED},
		{"attributes planet", attributes.Planets.EARTH},
		{"insensitive operation", insensitive.OperationTypes.UPDATE},
		{"invalid status", invalid.Statuses.PASSED},
		{"invalid alias status", invalidalias.Statuses.PASSED},
		{"multiple order", multiple.Orders.CREATED},
		{"multiple status", multiple.Statuses.PASSED},
		{"negative algorithm", crypto.Algorithms.AES256},
		{"plural discount", plural.DiscountTypes.SALE},
		{"quotes ticket", tickets.Tickets.APPROVAL_PENDING},
		{"skipvalues version", skipvalues.Versions.V3},
		{"time sale", timesale.Sales.SALES},
		{"typecomment operation", typecomment.OperationTypes.UPDATE},
		{"values planet", values.Planets.EARTH},
		{"values only discount", valuesonly.DiscountTypes.SALE},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertGeneratedInterfaces(t, tt.value)
		})
	}
}

func TestGeneratedEnumRuntimeBehavior(t *testing.T) {
	t.Parallel()

	if invalid.Statuses.FAILED.IsValid() {
		t.Fatal("invalid status should not be valid")
	}
	if !invalid.Statuses.PASSED.IsValid() {
		t.Fatal("valid status should be valid")
	}

	if crypto.Algorithms.NONE.IsValid() {
		t.Fatal("negative fixture NONE should be invalid")
	}
	if got, err := crypto.ParseAlgorithm(0); err != nil || got != crypto.Algorithms.AES256 {
		t.Fatalf("ParseAlgorithm(0) = %v, %v; want AES256, nil", got, err)
	}
	if got, err := crypto.ParseAlgorithm(1); err != nil || got != crypto.Algorithms.CHACHA20 {
		t.Fatalf("ParseAlgorithm(1) = %v, %v; want CHACHA20, nil", got, err)
	}
	if _, err := crypto.ParseAlgorithm(-1); err == nil {
		t.Fatal("ParseAlgorithm(-1) accepted invalid NONE value")
	}

	if got, err := tickets.ParseTicket("Pending Approval"); err != nil || got != tickets.Tickets.APPROVAL_PENDING {
		t.Fatalf("ParseTicket(Pending Approval) = %v, %v", got, err)
	}
	text, err := tickets.Tickets.APPROVAL_PENDING.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(text), "Pending Approval"; got != want {
		t.Fatalf("MarshalText() = %q, want %q", got, want)
	}
	data, err := json.Marshal(tickets.Tickets.APPROVAL_PENDING)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), `"Pending Approval"`; got != want {
		t.Fatalf("json.Marshal() = %s, want %s", got, want)
	}

	if multiple.Orders.CREATED.String() == multiple.Statuses.FAILED.String() {
		t.Fatal("multiple enum package mixed order and status names")
	}
}

func assertGeneratedInterfaces(t *testing.T, value any) {
	t.Helper()
	if _, ok := value.(json.Marshaler); !ok {
		t.Fatalf("%T does not implement json.Marshaler", value)
	}
	if _, ok := value.(encoding.TextMarshaler); !ok {
		t.Fatalf("%T does not implement encoding.TextMarshaler", value)
	}
	if _, ok := value.(encoding.BinaryMarshaler); !ok {
		t.Fatalf("%T does not implement encoding.BinaryMarshaler", value)
	}
	if _, ok := value.(interface{ IsValid() bool }); !ok {
		t.Fatalf("%T does not expose IsValid", value)
	}
	if _, ok := value.(interface{ String() string }); !ok {
		t.Fatalf("%T does not expose String", value)
	}
}

func expectedGeneratedFiles() []string {
	return []string{
		filepath.Join("aliases", "orders_enums.go"),
		filepath.Join("attributes", "planets_enums.go"),
		filepath.Join("insensitive", "operationtypes_enums.go"),
		filepath.Join("invalid", "statuses_enums.go"),
		filepath.Join("invalid_alias", "statuses_enums.go"),
		filepath.Join("multiple", "orders_enums.go"),
		filepath.Join("multiple", "statuses_enums.go"),
		filepath.Join("negative", "algorithms_enums.go"),
		filepath.Join("plural", "discounttypes_enums.go"),
		filepath.Join("quotes", "tickets_enums.go"),
		filepath.Join("skipvalues", "versions_enums.go"),
		filepath.Join("time", "sales_enums.go"),
		filepath.Join("typecomment", "operationtypes_enums.go"),
		filepath.Join("values", "planets_enums.go"),
		filepath.Join("values_only", "discounttypes_enums.go"),
	}
}
