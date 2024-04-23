package generator_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/zarldev/goenums/examples/sale"
	"github.com/zarldev/goenums/pkg/generator"
	"github.com/zarldev/goenums/pkg/generator/testdata/orders"
	"github.com/zarldev/goenums/pkg/generator/testdata/planets"
	planetsgravityonly "github.com/zarldev/goenums/pkg/generator/testdata/planets_gravity_only"
	plannetssimple "github.com/zarldev/goenums/pkg/generator/testdata/planets_simple"
	"github.com/zarldev/goenums/pkg/generator/testdata/validation"
)

var (
	testCases = []struct {
		name     string
		filename string
		failfast bool
		expected string
	}{
		{
			name:     "TestParseAndGenerate-Statuses-Strings",
			filename: "testdata/validation-strings/status.go",
			failfast: false,
			expected: "testdata/validation-strings/statuses_enums.go",
		},
		{
			name:     "TestParseAndGenerate-Statuses",
			filename: "testdata/validation/status.go",
			failfast: false,
			expected: "testdata/validation/statuses_enums.go",
		},

		{
			name:     "TestParseAndGenerate-Planets",
			filename: "testdata/planets/planets.go",
			failfast: false,
			expected: "testdata/planets/planets_enums.go",
		},
		{
			name:     "TestParseAndGenerate-PlanetsGravityOnly",
			filename: "testdata/planets_gravity_only/planets.go",
			failfast: false,
			expected: "testdata/planets_gravity_only/planets_enums.go",
		},
		{
			name:     "TestParseAndGenerate-PlanetsSimple",
			filename: "testdata/planets_simple/planets.go",
			failfast: false,
			expected: "testdata/planets_simple/planets_enums.go",
		},
		{
			name:     "TestParseAndGenerate-DiscountTypes",
			filename: "testdata/sale/discount.go",
			failfast: true,
			expected: "testdata/sale/discounttypes_enums.go",
		},
		{
			name:     "TestParseAndGenerate-Orders",
			filename: "testdata/orders/orders.go",
			failfast: false,
			expected: "testdata/orders/orders_enums.go",
		},
	}
)

func TestGenerator(t *testing.T) {
	// Setup
	// Clean up all previously generated files
	for _, tc := range testCases {
		err := os.Remove(tc.expected)
		if err != nil {
			t.Errorf("failed to cleanup generated files, got %v", err)
		}
	}
	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := generator.ParseAndGenerate(tc.filename, tc.failfast)
			if err != nil {
				t.Errorf("failed to generate enums for %s, got %v", tc.filename, err)
			}
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Check if the generated file exists
			_, err := os.Stat(tc.expected)
			if err != nil {
				t.Errorf("failed to find generated file %s, got %v", tc.expected, err)
			}
		})
	}
}

var (
	testCasesWithInvalid = []struct {
		name     string
		enums    []fmt.Stringer
		expected []string
	}{

		{
			name: "TestParseAndGenerate-Statuses",
			enums: []fmt.Stringer{
				validation.Statuses.FAILED,
				validation.Statuses.PASSED,
				validation.Statuses.SKIPPED,
				validation.Statuses.SCHEDULED,
				validation.Statuses.RUNNING,
				validation.Statuses.BOOKED,
			},
			expected: []string{
				"failed",
				"passed",
				"skipped",
				"scheduled",
				"running",
				"booked",
			},
		},
		{
			name: "TestParseAndGenerate-Planets",
			enums: []fmt.Stringer{
				planets.Planets.MERCURY,
				planets.Planets.VENUS,
				planets.Planets.EARTH,
				planets.Planets.MARS,
				planets.Planets.JUPITER,
				planets.Planets.SATURN,
				planets.Planets.URANUS,
				planets.Planets.NEPTUNE,
			},
			expected: []string{
				"Mercury",
				"Venus",
				"Earth",
				"Mars",
				"Jupiter",
				"Saturn",
				"Uranus",
				"Neptune",
			},
		},
		{
			name: "TestParseAndGenerate-Planets",
			enums: []fmt.Stringer{
				planets.Planets.MERCURY,
				planets.Planets.VENUS,
				planets.Planets.EARTH,
				planets.Planets.MARS,
				planets.Planets.JUPITER,
				planets.Planets.SATURN,
				planets.Planets.URANUS,
				planets.Planets.NEPTUNE,
			},
			expected: []string{
				"Mercury",
				"Venus",
				"Earth",
				"Mars",
				"Jupiter",
				"Saturn",
				"Uranus",
				"Neptune",
			},
		},
		{
			name: "TestParseAndGenerate-PlanetsGravityOnly",
			enums: []fmt.Stringer{
				planetsgravityonly.Planets.MERCURY,
				planetsgravityonly.Planets.VENUS,
				planetsgravityonly.Planets.EARTH,
				planetsgravityonly.Planets.MARS,
				planetsgravityonly.Planets.JUPITER,
				planetsgravityonly.Planets.SATURN,
				planetsgravityonly.Planets.URANUS,
				planetsgravityonly.Planets.NEPTUNE,
			},
			expected: []string{
				"mercury",
				"venus",
				"earth",
				"mars",
				"jupiter",
				"saturn",
				"uranus",
				"neptune",
			},
		},
		{
			name: "TestParseAndGenerate-PlanetsSimple",
			enums: []fmt.Stringer{
				plannetssimple.Planets.MERCURY,
				plannetssimple.Planets.VENUS,
				plannetssimple.Planets.EARTH,
				plannetssimple.Planets.MARS,
				plannetssimple.Planets.JUPITER,
				plannetssimple.Planets.SATURN,
				plannetssimple.Planets.URANUS,
				plannetssimple.Planets.NEPTUNE,
			},
			expected: []string{
				"Mercury",
				"Venus",
				"Earth",
				"Mars",
				"Jupiter",
				"Saturn",
				"Uranus",
				"Neptune",
			},
		},
		{
			name: "TestParseAndGenerate-DiscountTypes",
			enums: []fmt.Stringer{
				sale.DiscountTypes.SALE,
				sale.DiscountTypes.PERCENTAGE,
				sale.DiscountTypes.AMOUNT,
				sale.DiscountTypes.GIVEAWAY,
			},
			expected: []string{
				"sale",
				"percentage",
				"amount",
				"giveaway",
			},
		},
		{
			name: "TestParseAndGenerate-Orders",
			enums: []fmt.Stringer{
				orders.Orders.CREATED,
				orders.Orders.APPROVED,
				orders.Orders.PROCESSING,
				orders.Orders.READYTOSHIP,
				orders.Orders.SHIPPED,
				orders.Orders.DELIVERED,
				orders.Orders.CANCELLED,
			},
			expected: []string{
				"CREATED",
				"APPROVED",
				"PROCESSING",
				"READY_TO_SHIP",
				"SHIPPED",
				"DELIVERED",
				"CANCELLED",
			},
		},
	}
)

func TestGeneratedEnums(t *testing.T) {
	// Run test cases
	for _, tc := range testCasesWithInvalid {
		t.Run(tc.name, func(t *testing.T) {
			for i, v := range tc.enums {
				if v.String() != tc.expected[i] {
					fmt.Printf("expected: %s, got: %s\n", tc.expected[i], v.String())
					t.Errorf("failed to get the correct string representation for %s, got %v", tc.name, v.String())
				}
			}
		})
	}
}
