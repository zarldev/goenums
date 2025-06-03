package testdata

import (
	"embed"
	"time"

	"io"
	"io/fs"
	"testing"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/file"
	"github.com/zarldev/goenums/generator/config"
	"github.com/zarldev/goenums/generator/gofile"
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

	Config             config.Configuration
	Source             enum.Source
	p                  enum.Parser
	ExpectedFiles      []string
	GenerationRequests []enum.GenerationRequest
	Err                error
	Validate           func(t *testing.T, fs file.ReadStatFS)
}

func (i *InputOutputTest) SetParser(p enum.Parser) {
	i.p = p
}

func (i *InputOutputTest) Parser() enum.Parser {
	if i.p != nil {
		return i.p
	}
	return gofile.NewParser(
		gofile.WithParserConfiguration(i.Config),
		gofile.WithSource(i.Source))
}

var (
	statusRepresentation = enum.GenerationRequest{
		Package: "validation",
		Imports: []string{},
		EnumIota: enum.EnumIota{
			Type:       "status",
			Comment:    "",
			Fields:     []enum.Field{},
			Opener:     " ",
			Closer:     " ",
			StartIndex: 0,
			Enums: []enum.Enum{
				{
					Name:    "FAILED",
					Index:   0,
					Fields:  []enum.Field{},
					Aliases: []string{},
					Valid:   true,
				},
				{
					Name:    "PASSED",
					Index:   1,
					Fields:  []enum.Field{},
					Aliases: []string{},
					Valid:   true,
				},
				{
					Name:    "SKIPPED",
					Index:   2,
					Fields:  []enum.Field{},
					Aliases: []string{},
					Valid:   true,
				},
				{
					Name:    "SCHEDULED",
					Index:   3,
					Fields:  []enum.Field{},
					Aliases: []string{},
					Valid:   true,
				},
				{
					Name:    "RUNNING",
					Index:   4,
					Fields:  []enum.Field{},
					Aliases: []string{},
					Valid:   true,
				},
				{
					Name:    "BOOKED",
					Index:   5,
					Fields:  []enum.Field{},
					Aliases: []string{},
					Valid:   true,
				},
				{
					Name:    "CANCELLED",
					Index:   6,
					Fields:  []enum.Field{},
					Aliases: []string{},
					Valid:   true,
				},
				{
					Name:    "FAILED",
					Index:   7,
					Fields:  []enum.Field{},
					Aliases: []string{},
					Valid:   true,
				},
				{
					Name:    "COMPLETED",
					Index:   8,
					Fields:  []enum.Field{},
					Aliases: []string{},
					Valid:   true,
				},
			},
		},
		Version:        "test",
		SourceFilename: "status.go",
		OutputFilename: "",
		Configuration: config.Configuration{
			Failfast:    false,
			Legacy:      false,
			Insensitive: false,
			Handlers: config.Handlers{
				JSON:   true,
				Text:   true,
				YAML:   true,
				SQL:    true,
				Binary: true,
			},
		},
	}
	planetsRepresentation = enum.GenerationRequest{
		Package: "solarsystem",
		Imports: []string{},
		EnumIota: enum.EnumIota{
			Type:       "planet",
			Comment:    "",
			Fields:     []enum.Field{},
			Opener:     " ",
			Closer:     " ",
			StartIndex: 0,
			Enums: []enum.Enum{
				{
					Name:    "MERCURY",
					Index:   0,
					Fields:  []enum.Field{},
					Aliases: []string{},
					Valid:   true,
				},
				{
					Name:    "VENUS",
					Index:   1,
					Fields:  []enum.Field{},
					Aliases: []string{},
					Valid:   true,
				},
				{
					Name:    "EARTH",
					Index:   2,
					Fields:  []enum.Field{},
					Aliases: []string{},
					Valid:   true,
				},
				{
					Name:    "MARS",
					Index:   3,
					Fields:  []enum.Field{},
					Aliases: []string{},
					Valid:   true,
				},
				{
					Name:    "JUPITER",
					Index:   4,
					Fields:  []enum.Field{},
					Aliases: []string{},
					Valid:   true,
				},
				{
					Name:    "SATURN",
					Index:   5,
					Fields:  []enum.Field{},
					Aliases: []string{},
					Valid:   true,
				},
				{
					Name:    "URANUS",
					Index:   6,
					Fields:  []enum.Field{},
					Aliases: []string{},
					Valid:   true,
				},
				{
					Name:    "NEPTUNE",
					Index:   7,
					Fields:  []enum.Field{},
					Aliases: []string{},
					Valid:   true,
				},
			},
		},
		Version:        "test",
		SourceFilename: "planets.go",
		OutputFilename: "",
		Configuration: config.Configuration{
			Failfast:    false,
			Legacy:      false,
			Insensitive: false,
			Handlers: config.Handlers{
				JSON:   true,
				Text:   true,
				YAML:   true,
				SQL:    true,
				Binary: true,
			},
		},
	}
	saleRepresentation = enum.GenerationRequest{
		Package:        "sale",
		Version:        "test",
		SourceFilename: "sale.go",
		OutputFilename: "",
		Configuration: config.Configuration{
			Failfast:    false,
			Legacy:      false,
			Insensitive: false,
			Handlers: config.Handlers{
				JSON:   true,
				Text:   true,
				YAML:   true,
				SQL:    true,
				Binary: true,
			},
		},
		Imports: []string{},
		EnumIota: enum.EnumIota{
			Type:    "sale",
			Comment: "",
			Fields: []enum.Field{
				{Name: "Duration", Value: time.Duration(0)},
				{Name: "Amount", Value: 0},
			},
			Opener:     "",
			Closer:     "",
			StartIndex: 0,
			Enums: []enum.Enum{
				{
					Name:  "SALE",
					Index: 0,
					Fields: []enum.Field{
						{Name: "Duration", Value: "48h"},
						{Name: "Amount", Value: "25"},
					},
					Aliases: []string{},
					Valid:   true,
				},
				{
					Name:  "DISCOUNT",
					Index: 1,
					Fields: []enum.Field{
						{Name: "Duration", Value: "72h"},
						{Name: "Amount", Value: "50"},
					},
					Aliases: []string{},
					Valid:   true,
				},
				{
					Name:  "REFUND",
					Index: 2,
					Fields: []enum.Field{
						{Name: "Duration", Value: "0s"},
						{Name: "Amount", Value: "100"},
					},
					Aliases: []string{},
					Valid:   true,
				},
				{
					Name:  "GIVEAWAY",
					Index: 3,
					Fields: []enum.Field{
						{Name: "Duration", Value: "0s"},
						{Name: "Amount", Value: "100"},
					},
					Aliases: []string{},
					Valid:   true,
				},
				{
					Name:  "CANCELLED",
					Index: 4,
					Fields: []enum.Field{
						{Name: "Duration", Value: "0s"},
						{Name: "Amount", Value: "0"},
					},
					Aliases: []string{},
					Valid:   true,
				},
			},
		},
	}
	ticketsRepresentation = enum.GenerationRequest{
		Package: "tickets",
		Imports: []string{},
		EnumIota: enum.EnumIota{
			Type:    "ticket",
			Comment: "// comment string, validstate bool",
			Fields: []enum.Field{
				{Name: "comment", Value: ""},
				{Name: "validstate", Value: false},
			},
			Opener:     " ",
			Closer:     " ",
			StartIndex: 0,
			Enums: []enum.Enum{
				{
					Name:  "unknown",
					Index: 0,
					Fields: []enum.Field{
						{Name: "comment", Value: "Ticket not found"},
						{Name: "validstate", Value: false},
					},
					Aliases: []string{"Not Found", "Missing"},
					Valid:   false,
				},
				{
					Name:  "created",
					Index: 1,
					Fields: []enum.Field{
						{Name: "comment", Value: "Ticket created successfully"},
						{Name: "validstate", Value: true},
					},
					Aliases: []string{"Created Successfully", "Created"},
					Valid:   true,
				},
				{
					Name:  "pending",
					Index: 2,
					Fields: []enum.Field{
						{Name: "comment", Value: "Ticket is being processed"},
						{Name: "validstate", Value: true},
					},
					Aliases: []string{"In Progress", "Pending"},
					Valid:   true,
				},
				{
					Name:  "approval_pending",
					Index: 3,
					Fields: []enum.Field{
						{Name: "comment", Value: "Ticket is pending approval"},
						{Name: "validstate", Value: true},
					},
					Aliases: []string{"Pending Approval", "Approval Pending"},
					Valid:   true,
				},
				{
					Name:  "approval_accepted",
					Index: 4,
					Fields: []enum.Field{
						{Name: "comment", Value: "Ticket has been fully approved"},
						{Name: "validstate", Value: true},
					},
					Aliases: []string{"Fully Approved", "Approval Accepted"},
					Valid:   true,
				},
				{
					Name:  "rejected",
					Index: 5,
					Fields: []enum.Field{
						{Name: "comment", Value: "Ticket has been rejected"},
						{Name: "validstate", Value: false},
					},
					Aliases: []string{"Has Been Rejected", "Rejected"},
					Valid:   false,
				},
				{
					Name:  "completed",
					Index: 6,
					Fields: []enum.Field{
						{Name: "comment", Value: "Ticket has been completed"},
						{Name: "validstate", Value: false},
					},
					Aliases: []string{"Successfully Completed", "Completed"},
					Valid:   false,
				},
			},
		},
		Version:        "test",
		SourceFilename: "tickets.go",
		OutputFilename: "",
		Configuration: config.Configuration{
			Failfast:    false,
			Legacy:      false,
			Insensitive: false,
			Handlers: config.Handlers{
				JSON:   true,
				Text:   true,
				YAML:   true,
				SQL:    true,
				Binary: true,
			},
		},
	}

	ordersRepresentation = enum.GenerationRequest{
		Package: "orders",
		EnumIota: enum.EnumIota{
			Type:       "order",
			Comment:    "",
			Fields:     []enum.Field{},
			Opener:     " ",
			Closer:     " ",
			StartIndex: 0,
			Enums: []enum.Enum{
				{
					Name:  "created",
					Index: 0,
					Fields: []enum.Field{
						{Name: "Duration", Value: "24h"},
					},
					Aliases: []string{"CREATED"},
					Valid:   true,
				},
				{
					Name:  "approved",
					Index: 1,
					Fields: []enum.Field{
						{Name: "Duration", Value: "48h"},
					},
					Aliases: []string{"APPROVED"},
					Valid:   true,
				},
				{
					Name:  "processing",
					Index: 2,
					Fields: []enum.Field{
						{Name: "Duration", Value: "72h"},
					},
					Aliases: []string{"PROCESSING"},
					Valid:   true,
				},
				{
					Name:  "readyToShip",
					Index: 3,
					Fields: []enum.Field{
						{Name: "Duration", Value: "96h"},
					},
					Aliases: []string{"READY_TO_SHIP"},
					Valid:   true,
				},
				{
					Name:    "shipped",
					Index:   4,
					Fields:  []enum.Field{},
					Aliases: []string{"SHIPPED"},
					Valid:   true,
				},
				{
					Name:    "delivered",
					Index:   5,
					Fields:  []enum.Field{},
					Aliases: []string{"DELIVERED"},
					Valid:   true,
				},
				{
					Name:    "cancelled",
					Index:   6,
					Fields:  []enum.Field{},
					Aliases: []string{"CANCELLED"},
					Valid:   true,
				},
			},
		},
		Version:        "test",
		SourceFilename: "orders.go",
		OutputFilename: "",
		Configuration: config.Configuration{
			Failfast:    false,
			Legacy:      false,
			Insensitive: false,
			Handlers: config.Handlers{
				JSON:   true,
				Text:   true,
				YAML:   true,
				SQL:    true,
				Binary: true,
			},
		},
		Imports: []string{},
	}
	// package skipvalues

	// type version int

	// //go:generate goenums skipvalues.go
	// const (
	// 	V1 version = iota + 1
	// 	_
	// 	V3
	// 	V4
	// 	_
	// 	_
	// 	V7
	// )

	skipValuesRepresentation = enum.GenerationRequest{
		Package: "skipvalues",
		EnumIota: enum.EnumIota{
			Type:       "version",
			Comment:    "",
			Fields:     []enum.Field{},
			Opener:     " ",
			Closer:     " ",
			StartIndex: 1,
			Enums: []enum.Enum{
				{
					Name:    "V1",
					Index:   1,
					Fields:  []enum.Field{},
					Aliases: []string{},
					Valid:   true,
				},
				{
					Name:    "V3",
					Index:   3,
					Fields:  []enum.Field{},
					Aliases: []string{},
					Valid:   true,
				},
				{
					Name:    "V4",
					Index:   4,
					Fields:  []enum.Field{},
					Aliases: []string{},
					Valid:   true,
				},
				{
					Name:    "V7",
					Index:   7,
					Fields:  []enum.Field{},
					Aliases: []string{},
					Valid:   true,
				},
			},
		},
		Version:        "test",
		SourceFilename: "skipvalues.go",
		OutputFilename: "",
		Configuration: config.Configuration{
			Failfast:    false,
			Legacy:      false,
			Insensitive: false,
			Handlers: config.Handlers{
				JSON:   true,
				Text:   true,
				YAML:   true,
				SQL:    true,
				Binary: true,
			},
		},
		Imports: []string{},
	}
	//package discount
	// type discountType int // Available bool, Started bool, Finished bool, Cancelled bool, Duration time.Duration

	// const (
	// 	sale       discountType = iota + 1 // false,true,true,false,172h
	// 	percentage                         // false,false,false,false,24h
	// 	amount                             // false,false,false,false,48h
	// 	giveaway                           // true,true,false,false,72h
	// )

	discountRepresentation = enum.GenerationRequest{
		Package: "discount",
		EnumIota: enum.EnumIota{
			Type:    "discountType",
			Comment: "// Available bool, Started bool, Finished bool, Cancelled bool, Duration time.Duration",
			Fields: []enum.Field{
				{Name: "Available", Value: false},
				{Name: "Started", Value: false},
				{Name: "Finished", Value: false},
				{Name: "Cancelled", Value: false},
				{Name: "Duration", Value: time.Duration(0)},
			},
			Enums: []enum.Enum{
				{
					Name:    "sale",
					Index:   1,
					Fields:  []enum.Field{},
					Aliases: []string{},
					Valid:   true,
				},
				{
					Name:    "percentage",
					Index:   2,
					Fields:  []enum.Field{},
					Aliases: []string{},
					Valid:   true,
				},
				{
					Name:    "amount",
					Index:   3,
					Fields:  []enum.Field{},
					Aliases: []string{},
					Valid:   true,
				},
				{
					Name:    "giveaway",
					Index:   4,
					Fields:  []enum.Field{},
					Aliases: []string{},
					Valid:   true,
				},
			},
		},
		Version:        "test",
		SourceFilename: "discount.go",
		OutputFilename: "",
		Configuration: config.Configuration{
			Failfast:    false,
			Legacy:      false,
			Insensitive: false,
			Handlers: config.Handlers{
				JSON:   true,
				Text:   true,
				YAML:   true,
				SQL:    true,
				Binary: true,
			},
		},
		Imports: []string{},
	}
)
var (
	DefaultConfig        = config.Configuration{}
	FailFastLegacyConfig = config.Configuration{Failfast: true, Legacy: true}

	InputOutputTestCases = []InputOutputTest{
		{
			Name:               "enum with invalid entry - status - default",
			Source:             source.FromFileSystem(FS, "invalid/status.go"),
			Config:             DefaultConfig,
			ExpectedFiles:      []string{"invalid/statuses_enums.go"},
			GenerationRequests: []enum.GenerationRequest{statusRepresentation},
		},
		{
			Name:               "enum with invalid entry - status - failfast & legacy",
			Source:             source.FromFileSystem(FS, "invalid/status.go"),
			Config:             FailFastLegacyConfig,
			ExpectedFiles:      []string{"invalid/statuses_enums.go"},
			GenerationRequests: []enum.GenerationRequest{statusRepresentation},
		},
		{
			Name:               "enum with invalid entry with alias - status - default",
			Source:             source.FromFileSystem(FS, "invalid_alias/status.go"),
			Config:             DefaultConfig,
			ExpectedFiles:      []string{"invalid_alias/statuses_enums.go"},
			GenerationRequests: []enum.GenerationRequest{statusRepresentation},
		},
		{
			Name:               "enum with invalid entry with alias - status - failfast & legacy",
			Source:             source.FromFileSystem(FS, "invalid_alias/status.go"),
			Config:             FailFastLegacyConfig,
			ExpectedFiles:      []string{"invalid_alias/statuses_enums.go"},
			GenerationRequests: []enum.GenerationRequest{statusRepresentation},
		},
		{
			Name:               "enum with attributes - planets - default",
			Source:             source.FromFileSystem(FS, "attributes/planets.go"),
			Config:             DefaultConfig,
			ExpectedFiles:      []string{"attributes/planets_enums.go"},
			GenerationRequests: []enum.GenerationRequest{planetsRepresentation},
		},
		{
			Name:               "enum with attributes - planets - failfast & legacy",
			Source:             source.FromFileSystem(FS, "attributes/planets.go"),
			Config:             FailFastLegacyConfig,
			ExpectedFiles:      []string{"attributes/planets_enums.go"},
			GenerationRequests: []enum.GenerationRequest{planetsRepresentation},
		},
		{
			Name:               "enum with values - planets gravity only value - default",
			Source:             source.FromFileSystem(FS, "values/planets.go"),
			Config:             DefaultConfig,
			ExpectedFiles:      []string{"values/planets_enums.go"},
			GenerationRequests: []enum.GenerationRequest{planetsRepresentation},
		},
		{
			Name:               "enum with values only - sale discount types - default",
			Source:             source.FromFileSystem(FS, "values_only/discount.go"),
			Config:             DefaultConfig,
			ExpectedFiles:      []string{"values_only/discounttypes_enums.go"},
			GenerationRequests: []enum.GenerationRequest{discountRepresentation},
		},
		{
			Name:               "enum with values - planets gravity only value - failfast & legacy",
			Source:             source.FromFileSystem(FS, "values/planets.go"),
			Config:             FailFastLegacyConfig,
			ExpectedFiles:      []string{"values/planets_enums.go"},
			GenerationRequests: []enum.GenerationRequest{planetsRepresentation},
		},
		{
			Name:               "enum with time - sales - default",
			Config:             DefaultConfig,
			Source:             source.FromFileSystem(FS, "time/sale.go"),
			ExpectedFiles:      []string{"time/sales_enums.go"},
			GenerationRequests: []enum.GenerationRequest{saleRepresentation},
		},
		{
			Name:               "enum with time - sales - failfast & legacy",
			Config:             FailFastLegacyConfig,
			Source:             source.FromFileSystem(FS, "time/sale.go"),
			ExpectedFiles:      []string{"time/sales_enums.go"},
			GenerationRequests: []enum.GenerationRequest{saleRepresentation},
		},
		{
			Name:               "enum with quoted strings - tickets - default",
			Source:             source.FromFileSystem(FS, "quotes/tickets.go"),
			Config:             DefaultConfig,
			ExpectedFiles:      []string{"quotes/tickets_enums.go"},
			GenerationRequests: []enum.GenerationRequest{ticketsRepresentation},
		},
		{
			Name:               "enum with plural type - discount - failfast & legacy",
			Config:             FailFastLegacyConfig,
			Source:             source.FromFileSystem(FS, "time/sale.go"),
			ExpectedFiles:      []string{"time/sales_enums.go"},
			GenerationRequests: []enum.GenerationRequest{saleRepresentation},
		},
		{
			Name:               "enum with plural type - discount - default",
			Source:             source.FromFileSystem(FS, "plural/discount.go"),
			Config:             DefaultConfig,
			ExpectedFiles:      []string{"plural/discounttypes_enums.go"},
			GenerationRequests: []enum.GenerationRequest{saleRepresentation},
		},
		{
			Name:               "enum with quoted strings - tickets - failfast & legacy",
			Source:             source.FromFileSystem(FS, "quotes/tickets.go"),
			Config:             FailFastLegacyConfig,
			ExpectedFiles:      []string{"quotes/tickets_enums.go"},
			GenerationRequests: []enum.GenerationRequest{ticketsRepresentation},
		},
		{
			Name:               "enum with aliases - orders - default",
			Source:             source.FromFileSystem(FS, "aliases/orders.go"),
			Config:             DefaultConfig,
			ExpectedFiles:      []string{"aliases/orders_enums.go"},
			GenerationRequests: []enum.GenerationRequest{ordersRepresentation},
		},
		{
			Name:               "enum with aliases - orders - failfast & legacy",
			Source:             source.FromFileSystem(FS, "aliases/orders.go"),
			Config:             FailFastLegacyConfig,
			ExpectedFiles:      []string{"aliases/orders_enums.go"},
			GenerationRequests: []enum.GenerationRequest{ordersRepresentation},
		},
		{
			Name:               "multiple enums in one file - default",
			Source:             source.FromFileSystem(FS, "multiple/multiple.go"),
			Config:             DefaultConfig,
			ExpectedFiles:      []string{"multiple/statuses_enums.go", "multiple/orders_enums.go"},
			GenerationRequests: []enum.GenerationRequest{statusRepresentation, ordersRepresentation},
		},
		{
			Name:               "multiple enums in one file - failfast & legacy",
			Source:             source.FromFileSystem(FS, "multiple/multiple.go"),
			Config:             FailFastLegacyConfig,
			ExpectedFiles:      []string{"multiple/statuses_enums.go", "multiple/orders_enums.go"},
			GenerationRequests: []enum.GenerationRequest{statusRepresentation, ordersRepresentation},
		},
		{
			Name:               "enum with skip values - default",
			Source:             source.FromFileSystem(FS, "skipvalues/skipvalues.go"),
			Config:             DefaultConfig,
			ExpectedFiles:      []string{"skipvalues/versions_enums.go"},
			GenerationRequests: []enum.GenerationRequest{skipValuesRepresentation},
		},
		{
			Name:               "enum with skip values - failfast & legacy",
			Source:             source.FromFileSystem(FS, "skipvalues/skipvalues.go"),
			Config:             FailFastLegacyConfig,
			ExpectedFiles:      []string{"skipvalues/versions_enums.go"},
			GenerationRequests: []enum.GenerationRequest{skipValuesRepresentation},
		},
		{
			Name:   "non-existent file",
			Source: source.FromFileSystem(FS, "non/existent/file.go"),
			Config: DefaultConfig,
			Err:    fs.ErrNotExist,
		},
	}
)
