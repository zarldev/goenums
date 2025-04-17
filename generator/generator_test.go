package generator_test

import (
	"os"
	"testing"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/generator"
	"github.com/zarldev/goenums/generator/config"
	"github.com/zarldev/goenums/generator/gofile"
	"github.com/zarldev/goenums/source"
)

var (
	testCases = []struct {
		name     string
		Source   enum.Source
		Config   config.Configuration
		expected []string
	}{
		{
			name:     "TestParseAndGenerate-Statuses-Strings-failfast-legacy",
			Source:   source.FromFile("../internal/testdata/validation-strings/status.go"),
			Config:   config.Configuration{Failfast: true, Legacy: true},
			expected: []string{"../internal/testdata/validation-strings/statuses_enums.go"},
		},
		{
			name:     "TestParseAndGenerate-Statuses-failfast-legacy",
			Source:   source.FromFile("../internal/testdata/validation/status.go"),
			Config:   config.Configuration{Failfast: true, Legacy: true},
			expected: []string{"../internal/testdata/validation/statuses_enums.go"},
		},

		{
			name:     "TestParseAndGenerate-Planets-failfast-legacy",
			Source:   source.FromFile("../internal/testdata/planets/planets.go"),
			Config:   config.Configuration{Failfast: true, Legacy: true},
			expected: []string{"../internal/testdata/planets/planets_enums.go"},
		},
		{
			name:     "TestParseAndGenerate-PlanetsGravityOnly-failfast-legacy",
			Source:   source.FromFile("../internal/testdata/planets_gravity_only/planets.go"),
			Config:   config.Configuration{Failfast: true, Legacy: true},
			expected: []string{"../internal/testdata/planets_gravity_only/planets_enums.go"},
		},
		{
			name:     "TestParseAndGenerate-PlanetsSimple-failfast-legacy",
			Source:   source.FromFile("../internal/testdata/planets_simple/planets.go"),
			Config:   config.Configuration{Failfast: true, Legacy: true},
			expected: []string{"../internal/testdata/planets_simple/planets_enums.go"},
		},
		{
			name:     "TestParseAndGenerate-DiscountTypes-failfast-legacy",
			Source:   source.FromFile("../internal/testdata/sale/discount.go"),
			Config:   config.Configuration{Failfast: true, Legacy: true},
			expected: []string{"../internal/testdata/sale/discounttypes_enums.go"},
		},
		{
			name:     "TestParseAndGenerate-Orders-failfast-legacy",
			Source:   source.FromFile("../internal/testdata/orders/orders.go"),
			Config:   config.Configuration{Failfast: true, Legacy: true},
			expected: []string{"../internal/testdata/orders/orders_enums.go"},
		},
		{
			name:   "TestParseAndGenerate-Multiple-OrdersSales-failfast-legacy",
			Source: source.FromFile("../internal/testdata/multiple/multiple.go"),
			Config: config.Configuration{Failfast: true, Legacy: true},
			expected: []string{"../internal/testdata/multiple/orders_enums.go",
				"../internal/testdata/multiple/statuses_enums.go"},
		},
		{
			name:     "TestParseAndGenerate-TicketStatuses-Spaces-failfast-legacy",
			Source:   source.FromFile("../internal/testdata/spaces/tickets.go"),
			Config:   config.Configuration{Failfast: true, Legacy: true},
			expected: []string{"../internal/testdata/spaces/ticketstatuses_enums.go"},
		},
		{
			name:     "TestParseAndGenerate-TicketStatuses-Spaces",
			Source:   source.FromFile("../internal/testdata/spaces/tickets.go"),
			Config:   config.Configuration{Failfast: false, Legacy: false},
			expected: []string{"../internal/testdata/spaces/ticketstatuses_enums.go"},
		},
		{
			name:     "TestParseAndGenerate-Statuses-Strings",
			Source:   source.FromFile("../internal/testdata/validation-strings/status.go"),
			Config:   config.Configuration{Failfast: false, Legacy: false},
			expected: []string{"../internal/testdata/validation-strings/statuses_enums.go"},
		},
		{
			name:     "TestParseAndGenerate-Statuses",
			Source:   source.FromFile("../internal/testdata/validation/status.go"),
			Config:   config.Configuration{Failfast: false, Legacy: false},
			expected: []string{"../internal/testdata/validation/statuses_enums.go"},
		},
		{
			name:     "TestParseAndGenerate-Planets",
			Source:   source.FromFile("../internal/testdata/planets/planets.go"),
			Config:   config.Configuration{Failfast: false, Legacy: false},
			expected: []string{"../internal/testdata/planets/planets_enums.go"},
		},
		{
			name:     "TestParseAndGenerate-PlanetsGravityOnly",
			Source:   source.FromFile("../internal/testdata/planets_gravity_only/planets.go"),
			Config:   config.Configuration{Failfast: false, Legacy: false},
			expected: []string{"../internal/testdata/planets_gravity_only/planets_enums.go"},
		},
		{
			name:     "TestParseAndGenerate-PlanetsSimple",
			Source:   source.FromFile("../internal/testdata/planets_simple/planets.go"),
			Config:   config.Configuration{Failfast: false, Legacy: false},
			expected: []string{"../internal/testdata/planets_simple/planets_enums.go"},
		},
		{
			name:     "TestParseAndGenerate-DiscountTypes",
			Source:   source.FromFile("../internal/testdata/sale/discount.go"),
			Config:   config.Configuration{Failfast: false, Legacy: false},
			expected: []string{"../internal/testdata/sale/discounttypes_enums.go"},
		},
		{
			name:     "TestParseAndGenerate-Orders",
			Source:   source.FromFile("../internal/testdata/orders/orders.go"),
			Config:   config.Configuration{Failfast: false, Legacy: false},
			expected: []string{"../internal/testdata/orders/orders_enums.go"},
		},
		{
			name:   "TestParseAndGenerate-Multiple-OrdersSales",
			Source: source.FromFile("../internal/testdata/multiple/multiple.go"),
			Config: config.Configuration{Failfast: false, Legacy: false},
			expected: []string{"../internal/testdata/multiple/orders_enums.go",
				"../internal/testdata/multiple/statuses_enums.go"},
		},
	}
)

func TestGenerator(t *testing.T) {
	// Run test cases
	for _, tc := range testCases {
		// Setup
		// Clean up all previously generated files
		for _, filename := range tc.expected {
			err := os.Remove(filename)
			if err != nil {
				t.Errorf("failed to cleanup generated files, got %v", err)
			}
		}
		t.Run(tc.name, func(t *testing.T) {
			parser := gofile.NewParser(tc.Config, tc.Source)
			wri := gofile.NewWriter(tc.Config)
			p := generator.New(tc.Config, parser, wri)
			// Run
			err := p.ParseAndWrite(t.Context())
			if err != nil {
				t.Errorf("failed to generate enums for %s, got %v", tc.Source.Filename(), err)
			}
		})
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Check if the generated file exists
			for _, filename := range tc.expected {
				_, err := os.Stat(filename)
				if err != nil {
					t.Errorf("failed to find generated file %s, got %v", tc.expected, err)
				}
			}
		})
	}
}

// var (
// 	testCasesWithInvalid = []struct {
// 		name     string
// 		enums    []fmt.Stringer
// 		expected []string
// 	}{

// 		{
// 			name: "TestParseAndGenerate-Statuses",
// 			enums: []fmt.Stringer{
// 				validation.Statuses.FAILED,
// 				validation.Statuses.PASSED,
// 				validation.Statuses.SKIPPED,
// 				validation.Statuses.SCHEDULED,
// 				validation.Statuses.RUNNING,
// 				validation.Statuses.BOOKED,
// 			},
// 			expected: []string{
// 				"failed",
// 				"passed",
// 				"skipped",
// 				"scheduled",
// 				"running",
// 				"booked",
// 			},
// 		},
// 		{
// 			name: "TestParseAndGenerate-Planets",
// 			enums: []fmt.Stringer{
// 				planets.Planets.MERCURY,
// 				planets.Planets.VENUS,
// 				planets.Planets.EARTH,
// 				planets.Planets.MARS,
// 				planets.Planets.JUPITER,
// 				planets.Planets.SATURN,
// 				planets.Planets.URANUS,
// 				planets.Planets.NEPTUNE,
// 			},
// 			expected: []string{
// 				"Mercury",
// 				"Venus",
// 				"Earth",
// 				"Mars",
// 				"Jupiter",
// 				"Saturn",
// 				"Uranus",
// 				"Neptune",
// 			},
// 		},
// 		{
// 			name: "TestParseAndGenerate-PlanetsGravityOnly",
// 			enums: []fmt.Stringer{
// 				planets_gravity_only.Planets.MERCURY,
// 				planets_gravity_only.Planets.VENUS,
// 				planets_gravity_only.Planets.EARTH,
// 				planets_gravity_only.Planets.MARS,
// 				planets_gravity_only.Planets.JUPITER,
// 				planets_gravity_only.Planets.SATURN,
// 				planets_gravity_only.Planets.URANUS,
// 				planets_gravity_only.Planets.NEPTUNE,
// 			},
// 			expected: []string{
// 				"mercury",
// 				"venus",
// 				"earth",
// 				"mars",
// 				"jupiter",
// 				"saturn",
// 				"uranus",
// 				"neptune",
// 			},
// 		},
// 		{
// 			name: "TestParseAndGenerate-PlanetsSimple",
// 			enums: []fmt.Stringer{
// 				planets_simple.Planets.MERCURY,
// 				planets_simple.Planets.VENUS,
// 				planets_simple.Planets.EARTH,
// 				planets_simple.Planets.MARS,
// 				planets_simple.Planets.JUPITER,
// 				planets_simple.Planets.SATURN,
// 				planets_simple.Planets.URANUS,
// 				planets_simple.Planets.NEPTUNE,
// 			},
// 			expected: []string{
// 				"Mercury",
// 				"Venus",
// 				"Earth",
// 				"Mars",
// 				"Jupiter",
// 				"Saturn",
// 				"Uranus",
// 				"Neptune",
// 			},
// 		},
// 		{
// 			name: "TestParseAndGenerate-DiscountTypes",
// 			enums: []fmt.Stringer{
// 				sale.DiscountTypes.SALE,
// 				sale.DiscountTypes.PERCENTAGE,
// 				sale.DiscountTypes.AMOUNT,
// 				sale.DiscountTypes.GIVEAWAY,
// 			},
// 			expected: []string{
// 				"sale",
// 				"percentage",
// 				"amount",
// 				"giveaway",
// 			},
// 		},
// 		{
// 			name: "TestParseAndGenerate-Orders",
// 			enums: []fmt.Stringer{
// 				orders.Orders.CREATED,
// 				orders.Orders.APPROVED,
// 				orders.Orders.PROCESSING,
// 				orders.Orders.READYTOSHIP,
// 				orders.Orders.SHIPPED,
// 				orders.Orders.DELIVERED,
// 				orders.Orders.CANCELLED,
// 			},
// 			expected: []string{
// 				"CREATED",
// 				"APPROVED",
// 				"PROCESSING",
// 				"READY_TO_SHIP",
// 				"SHIPPED",
// 				"DELIVERED",
// 				"CANCELLED",
// 			},
// 		},
// 		{
// 			name: "TestParseAndGenerate-TicketStatuses-Spaces",
// 			enums: []fmt.Stringer{
// 				spaces.TicketStatuses.PENDING,
// 				spaces.TicketStatuses.APPROVED,
// 				spaces.TicketStatuses.REJECTED,
// 				spaces.TicketStatuses.COMPLETED,
// 			},
// 			expected: []string{
// 				"In Progress",
// 				"Fully Approved",
// 				"Has Been Rejected",
// 				"Successfully Completed",
// 			},
// 		},
// 	}
// )

// func TestGeneratedEnums(t *testing.T) {
// 	// Run test cases
// 	for _, tc := range testCasesWithInvalid {
// 		t.Run(tc.name, func(t *testing.T) {
// 			for i, v := range tc.enums {
// 				if v.String() != tc.expected[i] {
// 					fmt.Printf("expected: %s, got: %s\n", tc.expected[i], v.String())
// 					t.Errorf("failed to get the correct string representation for %s, got %v", tc.name, v.String())
// 				}
// 			}
// 		})
// 	}
// }

// func TestEnumParsing(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		input    string
// 		enum     interface{}
// 		wantErr  bool
// 		expected string
// 	}{
// 		{
// 			name:     "Planet Exact Match",
// 			input:    "Mercury",
// 			enum:     planets.ParsePlanet,
// 			wantErr:  false,
// 			expected: "Mercury",
// 		},
// 		// Case insensitive matches
// 		{
// 			name:     "Planet Lowercase",
// 			input:    "mercury",
// 			enum:     planets.ParsePlanet,
// 			wantErr:  false,
// 			expected: "Mercury",
// 		},
// 		{
// 			name:     "Planet Uppercase",
// 			input:    "MERCURY",
// 			enum:     planets.ParsePlanet,
// 			wantErr:  false,
// 			expected: "Mercury",
// 		},
// 		{
// 			name:     "Ticket Status With Spaces",
// 			input:    "In Progress",
// 			enum:     spaces.ParseTicketStatus,
// 			wantErr:  false,
// 			expected: "In Progress",
// 		},
// 		{
// 			name:     "Ticket Status Case Insensitive",
// 			input:    "in progress",
// 			enum:     spaces.ParseTicketStatus,
// 			wantErr:  false,
// 			expected: "In Progress",
// 		},
// 		// Invalid values
// 		// {
// 		// 	name:     "Invalid Planet",
// 		// 	input:    "Pluto",
// 		// 	enum:     planets.ParsePlanet,
// 		// 	wantErr:  true,
// 		// 	expected: "",
// 		// },
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			switch parser := tt.enum.(type) {
// 			case func(string) (planets.Planet, error):
// 				res, err := parser(tt.input)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("Parse error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if !tt.wantErr && res.String() != tt.expected {
// 					t.Errorf("got %v, want %v", res, tt.expected)
// 				}
// 			case func(string) (spaces.TicketStatus, error):
// 				res, err := parser(tt.input)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("Parse error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if !tt.wantErr && res.String() != tt.expected {
// 					t.Errorf("got %v, want %v", res, tt.expected)
// 				}
// 			}
// 		})
// 	}
// }
