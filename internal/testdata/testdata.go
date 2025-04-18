package testdata

import (
	"embed"

	"io"
	"io/fs"
	"testing"
	"time"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/file"
	"github.com/zarldev/goenums/generator/config"
	"github.com/zarldev/goenums/source"
)

//go:embed *
var testDataFS embed.FS

var (
	_ file.ReadStatFS        = TestDataFS{}
	_ file.CreateWriteFileFS = TestDataFS{}
)

type TestDataFS struct {
	read  embed.FS
	write *file.MemFS
}

// Create implements file.CreateWriteFileFS.
func (fs TestDataFS) Create(name string) (io.WriteCloser, error) {
	return fs.write.Create(name)
}

// WriteFile implements file.CreateWriteFileFS.
func (fs TestDataFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return fs.write.WriteFile(name, data, perm)
}

// Open implements file.ReadStatFS.
func (fs TestDataFS) Open(name string) (fs.File, error) {
	f, err := fs.read.Open(name)
	if err != nil {
		f, err = fs.write.Open(name)
		if err != nil {
			return nil, err
		}
	}
	return f, nil
}

// Stat implements file.ReadStatFS.
func (fs TestDataFS) Stat(name string) (fs.FileInfo, error) {
	f, err := fs.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f.Stat()
}

var FS = TestDataFS{
	read:  testDataFS,
	write: file.NewMemFS(),
}

func (fs TestDataFS) ReadFile(name string) ([]byte, error) {
	b, err := fs.read.ReadFile(name)
	if err != nil {
		b, err = fs.write.ReadFile(name)
		if err != nil {
			return nil, err
		}
	}
	return b, nil
}

type InputOutputTest struct {
	Name string

	Config          config.Configuration
	Source          enum.Source
	ExpectedFiles   []string
	Representations []enum.Representation
	Err             error
	Validate        func(t *testing.T, fs file.ReadStatFS)
}

var (
	statusRepresentation = enum.Representation{
		Version:        "v1.0.0",
		GenerationTime: time.Now(),
		PackageName:    "statuses",
		SourceFilename: "status/status.go",
		TypeInfo: enum.TypeInfo{
			Name:        "status",
			Camel:       "Status",
			Lower:       "status",
			Upper:       "STATUS",
			Plural:      "statuses",
			PluralCamel: "Statuses",
			NameTypePair: []enum.NameTypePair{
				{
					Name:  "unknown",
					Type:  "status",
					Value: "0",
				},
				{
					Name:  "failed",
					Type:  "status",
					Value: "1",
				},
				{
					Name:  "passed",
					Type:  "status",
					Value: "2",
				},
				{
					Name:  "skipped",
					Type:  "status",
					Value: "3",
				},
				{
					Name:  "scheduled",
					Type:  "status",
					Value: "4",
				},
				{
					Name:  "running",
					Type:  "status",
					Value: "5",
				},
				{
					Name:  "booked",
					Type:  "status",
					Value: "6",
				},
			},
		},
		Enums: []enum.Enum{
			{
				Info: enum.Info{
					Name:  "unknown",
					Upper: "UNKNOWN",
					Alias: "unknown",
					Value: 0,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "status",
					Camel: "Status",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "unknown",
							Type:  "status",
							Value: "0",
						},
					},
				},
			},
			{
				Info: enum.Info{
					Name:  "failed",
					Upper: "FAILED",
					Alias: "failed",
					Value: 1,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "status",
					Camel: "Status",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "failed",
							Type:  "status",
							Value: "1",
						},
					},
				},
				Raw: enum.Raw{
					Comment: "FAILED",
				},
			},
			{
				Info: enum.Info{
					Name:  "passed",
					Upper: "PASSED",
					Alias: "passed",
					Value: 2,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "status",
					Camel: "Status",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "passed",
							Type:  "status",
							Value: "2",
						},
					},
				},
			},
			{
				Info: enum.Info{
					Name:  "skipped",
					Upper: "SKIPPED",
					Alias: "skipped",
					Value: 3,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "status",
					Camel: "Status",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "skipped",
							Type:  "status",
							Value: "3",
						},
					},
				},
			},
			{
				Info: enum.Info{
					Name:  "scheduled",
					Upper: "SCHEDULED",
					Alias: "scheduled",
					Value: 4,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "status",
					Camel: "Status",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "scheduled",
							Type:  "status",
							Value: "4",
						},
					},
				},
			},
			{
				Info: enum.Info{
					Name:  "running",
					Upper: "RUNNING",
					Alias: "running",
					Value: 5,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "status",
					Camel: "Status",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "running",
							Type:  "status",
							Value: "5",
						},
					},
				},
			},
			{
				Info: enum.Info{
					Name:  "booked",
					Upper: "BOOKED",
					Alias: "booked",
					Value: 6,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "status",
					Camel: "Status",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "booked",
							Type:  "status",
							Value: "6",
						},
					},
				},
				Raw: enum.Raw{
					Comment: "BOOKED",
				},
			},
		},
	}
	planetsRepresentation = enum.Representation{
		Version:        "v1.0.0",
		GenerationTime: time.Now(),
		PackageName:    "planets",
		SourceFilename: "planets/planets.go",
		TypeInfo: enum.TypeInfo{
			Name:        "planet",
			Camel:       "Planet",
			Lower:       "planet",
			Upper:       "PLANET",
			Plural:      "planets",
			PluralCamel: "Planets",
			NameTypePair: []enum.NameTypePair{
				{
					Name:  "unknown",
					Type:  "planet",
					Value: "0",
				},
				{
					Name:  "mercury",
					Type:  "planet",
					Value: "1",
				},
				{
					Name:  "venus",
					Type:  "planet",
					Value: "2",
				},
				{
					Name:  "earth",
					Type:  "planet",
					Value: "3",
				},
				{
					Name:  "mars",
					Type:  "planet",
					Value: "4",
				},
				{
					Name:  "jupiter",
					Type:  "planet",
					Value: "5",
				},
				{
					Name:  "saturn",
					Type:  "planet",
					Value: "6",
				},
				{
					Name:  "uranus",
					Type:  "planet",
					Value: "7",
				},
				{
					Name:  "neptune",
					Type:  "planet",
					Value: "8",
				},
			},
		},
		Enums: []enum.Enum{
			{
				Info: enum.Info{
					Name:  "unknown",
					Upper: "UNKNOWN",
					Alias: "unknown",
					Value: 0,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "planet",
					Camel: "Planet",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "unknown",
							Type:  "planet",
							Value: "0",
						},
					},
				},
			},
			{
				Info: enum.Info{
					Name:  "mercury",
					Upper: "MERCURY",
					Alias: "mercury",
					Value: 1,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "planet",
					Camel: "Planet",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "mercury",
							Type:  "planet",
							Value: "1",
						},
					},
				},
			},
			{
				Info: enum.Info{
					Name:  "venus",
					Upper: "VENUS",
					Alias: "venus",
					Value: 2,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "planet",
					Camel: "Planet",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "venus",
							Type:  "planet",
							Value: "2",
						},
					},
				},
			},
			{
				Info: enum.Info{
					Name:  "earth",
					Upper: "EARTH",
					Alias: "earth",
					Value: 3,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "planet",
					Camel: "Planet",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "earth",
							Type:  "planet",
							Value: "3",
						},
					},
				},
			},
			{
				Info: enum.Info{
					Name:  "mars",
					Upper: "MARS",
					Alias: "mars",
					Value: 4,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "planet",
					Camel: "Planet",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "mars",
							Type:  "planet",
							Value: "4",
						},
					},
				},
			},
			{
				Info: enum.Info{
					Name:  "jupiter",
					Upper: "JUPITER",
					Alias: "jupiter",
					Value: 5,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "planet",
					Camel: "Planet",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "jupiter",
							Type:  "planet",
							Value: "5",
						},
					},
				},
			},
			{
				Info: enum.Info{
					Name:  "saturn",
					Upper: "SATURN",
					Alias: "saturn",
					Value: 6,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "planet",
					Camel: "Planet",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "saturn",
							Type:  "planet",
							Value: "6",
						},
					},
				},
			},
			{
				Info: enum.Info{
					Name:  "uranus",
					Upper: "URANUS",
					Alias: "uranus",
					Value: 7,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "planet",
					Camel: "Planet",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "uranus",
							Type:  "planet",
							Value: "7",
						},
					},
				},
			},
			{
				Info: enum.Info{
					Name:  "neptune",
					Upper: "NEPTUNE",
					Alias: "neptune",
					Value: 8,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "planet",
					Camel: "Planet",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "neptune",
							Type:  "planet",
							Value: "8",
						},
					},
				},
				Raw: enum.Raw{
					Comment: "NEPTUNE",
				},
			},
		},
	}
	saleRepresentation = enum.Representation{
		Version:        "v1.0.0",
		GenerationTime: time.Now(),
		PackageName:    "sale",
		SourceFilename: "sale/sale.go",
		TypeInfo: enum.TypeInfo{
			Name:        "sale",
			Camel:       "Sale",
			Lower:       "sale",
			Upper:       "SALE",
			Plural:      "sales",
			PluralCamel: "Sales",
			NameTypePair: []enum.NameTypePair{
				{
					Name:  "sales",
					Type:  "sale",
					Value: "0",
				},
				{
					Name:  "percentage",
					Type:  "sale",
					Value: "1",
				},
				{
					Name:  "amount",
					Type:  "sale",
					Value: "2",
				},
				{
					Name:  "giveaway",
					Type:  "sale",
					Value: "3",
				},
			},
		},
		Enums: []enum.Enum{
			{
				Info: enum.Info{
					Name:  "sales",
					Upper: "SALES",
					Alias: "sales",
					Value: 0,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "sale",
					Camel: "Sale",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "sales",
							Type:  "sale",
							Value: "0",
						},
					},
				},
				Raw: enum.Raw{
					Comment: "SALES",
				},
			},
			{
				Info: enum.Info{
					Name:  "percentage",
					Upper: "PERCENTAGE",
					Alias: "percentage",
					Value: 1,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "sale",
					Camel: "Sale",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "percentage",
							Type:  "sale",
							Value: "1",
						},
					},
					Index: 1,
				},
			},
			{
				Info: enum.Info{
					Name:  "amount",
					Upper: "AMOUNT",
					Alias: "amount",
					Value: 2,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "sale",
					Camel: "Sale",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "amount",
							Type:  "sale",
							Value: "2",
						},
					},
					Index: 2,
				},
			},
			{
				Info: enum.Info{
					Name:  "giveaway",
					Upper: "GIVEAWAY",
					Alias: "giveaway",
					Value: 3,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "sale",
					Camel: "Sale",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "giveaway",
							Type:  "sale",
							Value: "3",
						},
					},
					Index: 3,
				},
				Raw: enum.Raw{
					Comment: "GIVEAWAY",
				},
			},
		},
	}
	ticketsRepresentation = enum.Representation{
		Version:        "v1.0.0",
		GenerationTime: time.Now(),
		PackageName:    "tickets",
		SourceFilename: "tickets/tickets.go",
		TypeInfo: enum.TypeInfo{
			Name:        "ticket",
			Camel:       "Ticket",
			Lower:       "ticket",
			Upper:       "TICKET",
			Plural:      "tickets",
			PluralCamel: "Tickets",
			NameTypePair: []enum.NameTypePair{
				{
					Name:  "unknown",
					Type:  "ticket",
					Value: "0",
				},
				{
					Name:  "open",
					Type:  "ticket",
					Value: "1",
				},
				{
					Name:  "closed",
					Type:  "ticket",
					Value: "2",
				},
				{
					Name:  "cancelled",
					Type:  "ticket",
					Value: "3",
				},
				{
					Name:  "expired",
					Type:  "ticket",
					Value: "4",
				},
			},
		},
		Enums: []enum.Enum{
			{
				Info: enum.Info{
					Name:  "unknown",
					Upper: "UNKNOWN",
					Alias: "unknown",
					Value: 0,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "ticket",
					Camel: "Ticket",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "unknown",
							Type:  "ticket",
							Value: "0",
						},
					},
					Index: 0,
				},
			},
			{
				Info: enum.Info{
					Name:  "open",
					Upper: "OPEN",
					Alias: "open",
					Value: 1,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "ticket",
					Camel: "Ticket",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "open",
							Type:  "ticket",
							Value: "1",
						},
					},
					Index: 1,
				},
			},
			{
				Info: enum.Info{
					Name:  "closed",
					Upper: "CLOSED",
					Alias: "closed",
					Value: 2,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "ticket",
					Camel: "Ticket",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "closed",
							Type:  "ticket",
							Value: "2",
						},
					},
					Index: 2,
				},
				Raw: enum.Raw{
					Comment: "CLOSED",
				},
			},
		},
	}
	ordersRepresentation = enum.Representation{
		Version:        "v1.0.0",
		GenerationTime: time.Now(),
		PackageName:    "orders",
		SourceFilename: "orders/orders.go",
		TypeInfo: enum.TypeInfo{
			Name:        "order",
			Camel:       "Order",
			Lower:       "order",
			Upper:       "ORDER",
			Plural:      "orders",
			PluralCamel: "Orders",
			NameTypePair: []enum.NameTypePair{
				{
					Name:  "created",
					Type:  "order",
					Value: "0",
				},
				{
					Name:  "approved",
					Type:  "order",
					Value: "1",
				},
				{
					Name:  "processing",
					Type:  "order",
					Value: "2",
				},
				{
					Name:  "readyToShip",
					Type:  "order",
					Value: "3",
				},
				{
					Name:  "shipped",
					Type:  "order",
					Value: "4",
				},
				{
					Name:  "delivered",
					Type:  "order",
					Value: "5",
				},
				{
					Name:  "cancelled",
					Type:  "order",
					Value: "6",
				},
			},
		},
		Enums: []enum.Enum{
			{
				Info: enum.Info{
					Name:  "created",
					Upper: "CREATED",
					Alias: "created",
					Value: 0,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "order",
					Camel: "Order",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "created",
							Type:  "order",
							Value: "0",
						},
					}, Index: 0,
				},
			},
			{
				Info: enum.Info{
					Name:  "approved",
					Upper: "APPROVED",
					Alias: "approved",
					Value: 1,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "order",
					Camel: "Order",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "approved",
							Type:  "order",
							Value: "1",
						},
					},
					Index: 1,
				},
			},
			{
				Info: enum.Info{
					Name:  "processing",
					Upper: "PROCESSING",
					Alias: "processing",
					Value: 2,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "order",
					Camel: "Order",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "processing",
							Type:  "order",
							Value: "2",
						},
					},
				},
			},
			{
				Info: enum.Info{
					Name:  "readyToShip",
					Upper: "READY_TO_SHIP",
					Alias: "readyToShip",
					Value: 3,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "order",
					Camel: "Order",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "readyToShip",
							Type:  "order",
							Value: "3",
						},
					},
					Index: 3,
				},
			},
			{
				Info: enum.Info{
					Name:  "shipped",
					Upper: "SHIPPED",
					Alias: "shipped",
					Value: 4,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "order",
					Camel: "Order",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "shipped",
							Type:  "order",
							Value: "4",
						},
					},
				},
			},
			{
				Info: enum.Info{
					Name:  "delivered",
					Upper: "DELIVERED",
					Alias: "delivered",
					Value: 5,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "order",
					Camel: "Order",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "delivered",
							Type:  "order",
							Value: "5",
						},
					},
				},
			},
			{
				Info: enum.Info{
					Name:  "cancelled",
					Upper: "CANCELLED",
					Alias: "cancelled",
					Value: 6,
					Valid: true,
				},
				TypeInfo: enum.TypeInfo{
					Name:  "order",
					Camel: "Order",
					NameTypePair: []enum.NameTypePair{
						{
							Name:  "cancelled",
							Type:  "order",
							Value: "6",
						},
					},
					Index: 6,
				},
			},
		},
	}
)
var (
	DefaultConfig        = config.Configuration{}
	FailFastLegacyConfig = config.Configuration{Failfast: true, Legacy: true}

	InputOutputTestCases = []InputOutputTest{
		{
			Name:            "status - default",
			Source:          source.FromFileSystem(FS, "status/status.go"),
			Config:          DefaultConfig,
			ExpectedFiles:   []string{"status/statuses_enums.go"},
			Representations: []enum.Representation{statusRepresentation},
		},
		{
			Name:            "status - failfast & legacy",
			Source:          source.FromFileSystem(FS, "status/status.go"),
			Config:          FailFastLegacyConfig,
			ExpectedFiles:   []string{"status/statuses_enums.go"},
			Representations: []enum.Representation{statusRepresentation},
		},
		{
			Name:            "status with alias - default",
			Source:          source.FromFileSystem(FS, "status_alias/status.go"),
			Config:          DefaultConfig,
			ExpectedFiles:   []string{"status/statuses_enums.go"},
			Representations: []enum.Representation{statusRepresentation},
		},
		{
			Name:            "status with alias - failfast & legacy",
			Source:          source.FromFileSystem(FS, "status_alias/status.go"),
			Config:          FailFastLegacyConfig,
			ExpectedFiles:   []string{"status/statuses_enums.go"},
			Representations: []enum.Representation{statusRepresentation},
		},
		{
			Name:            "planets with attributes - default",
			Source:          source.FromFileSystem(FS, "planets/planets.go"),
			Config:          DefaultConfig,
			ExpectedFiles:   []string{"planets/planets_enums.go"},
			Representations: []enum.Representation{planetsRepresentation},
		},
		{
			Name:            "planets with attributes - failfast & legacy",
			Source:          source.FromFileSystem(FS, "planets/planets.go"),
			Config:          FailFastLegacyConfig,
			ExpectedFiles:   []string{"planets/planets_enums.go"},
			Representations: []enum.Representation{planetsRepresentation},
		},
		{
			Name:            "planets gravity only - default",
			Source:          source.FromFileSystem(FS, "planets_gravity_only/planets.go"),
			Config:          DefaultConfig,
			ExpectedFiles:   []string{"planets_gravity_only/planets_enums.go"},
			Representations: []enum.Representation{planetsRepresentation},
		},
		{
			Name:            "planets gravity only - failfast & legacy",
			Source:          source.FromFileSystem(FS, "planets_gravity_only/planets.go"),
			Config:          FailFastLegacyConfig,
			ExpectedFiles:   []string{"planets_gravity_only/planets_enums.go"},
			Representations: []enum.Representation{planetsRepresentation},
		},
		{
			Name:            "planets simple - default",
			Source:          source.FromFileSystem(FS, "planets_simple/planets.go"),
			Config:          DefaultConfig,
			ExpectedFiles:   []string{"planets_simple/planets_enums.go"},
			Representations: []enum.Representation{planetsRepresentation},
		},
		{
			Name:            "planets simple - failfast & legacy",
			Source:          source.FromFileSystem(FS, "planets_simple/planets.go"),
			Config:          FailFastLegacyConfig,
			ExpectedFiles:   []string{"planets_simple/planets_enums.go"},
			Representations: []enum.Representation{planetsRepresentation},
		},
		{
			Name:            "sale - default",
			Config:          DefaultConfig,
			Source:          source.FromFileSystem(FS, "sale/sale.go"),
			ExpectedFiles:   []string{"sale/sales_enums.go"},
			Representations: []enum.Representation{saleRepresentation},
		},
		{
			Name:            "sale - failfast & legacy",
			Config:          FailFastLegacyConfig,
			Source:          source.FromFileSystem(FS, "sale/sale.go"),
			ExpectedFiles:   []string{"sale/sales_enums.go"},
			Representations: []enum.Representation{saleRepresentation},
		},
		{
			Name:            "tickets - default",
			Source:          source.FromFileSystem(FS, "tickets/tickets.go"),
			Config:          DefaultConfig,
			ExpectedFiles:   []string{"tickets/tickets_enums.go"},
			Representations: []enum.Representation{ticketsRepresentation},
		},
		{
			Name:            "tickets - failfast & legacy",
			Source:          source.FromFileSystem(FS, "tickets/tickets.go"),
			Config:          FailFastLegacyConfig,
			ExpectedFiles:   []string{"tickets/tickets_enums.go"},
			Representations: []enum.Representation{ticketsRepresentation},
		},
		{
			Name:            "orders - default",
			Source:          source.FromFileSystem(FS, "orders/orders.go"),
			Config:          DefaultConfig,
			ExpectedFiles:   []string{"orders/orders_enums.go"},
			Representations: []enum.Representation{ordersRepresentation},
		},
		{
			Name:            "orders - failfast & legacy",
			Source:          source.FromFileSystem(FS, "orders/orders.go"),
			Config:          FailFastLegacyConfig,
			ExpectedFiles:   []string{"orders/orders_enums.go"},
			Representations: []enum.Representation{ordersRepresentation},
		},
		{
			Name:            "multiple in one file - default",
			Source:          source.FromFileSystem(FS, "multiple/multiple.go"),
			Config:          DefaultConfig,
			ExpectedFiles:   []string{"multiple/statuses_enums.go", "multiple/orders_enums.go"},
			Representations: []enum.Representation{statusRepresentation, ordersRepresentation},
		},
		{
			Name:            "multiple in one file - failfast & legacy",
			Source:          source.FromFileSystem(FS, "multiple/multiple.go"),
			Config:          FailFastLegacyConfig,
			ExpectedFiles:   []string{"multiple/statuses_enums.go", "multiple/orders_enums.go"},
			Representations: []enum.Representation{statusRepresentation, ordersRepresentation},
		},
		{
			Name:   "non-existent file",
			Source: source.FromFileSystem(FS, "non/existent/file.go"),
			Config: DefaultConfig,
			Err:    fs.ErrNotExist,
		},
	}
)
