package testdata

import (
	"embed"
	_ "embed"
	"io"
	"io/fs"

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

var (
	defaultConfig        = config.Configuration{}
	failFastLegacyConfig = config.Configuration{Failfast: true, Legacy: true}

	InputOutputTestCases = []struct {
		Name string

		Config              config.Configuration
		Source              enum.Source
		ExpectedFiles       []string
		RepresentationCount int
		Err                 error
	}{
		{
			Name:                "status - default",
			Source:              source.FromFileSystem(FS, "status/status.go"),
			Config:              defaultConfig,
			ExpectedFiles:       []string{"status/statuses_enums.go"},
			RepresentationCount: 1,
		},
		{
			Name:                "status - failfast & legacy",
			Source:              source.FromFileSystem(FS, "status/status.go"),
			Config:              failFastLegacyConfig,
			ExpectedFiles:       []string{"status/statuses_enums.go"},
			RepresentationCount: 1,
		},
		{
			Name:                "status with alias - default",
			Source:              source.FromFileSystem(FS, "status_alias/status.go"),
			Config:              defaultConfig,
			ExpectedFiles:       []string{"status/statuses_enums.go"},
			RepresentationCount: 1,
		},
		{
			Name:                "status with alias - failfast & legacy",
			Source:              source.FromFileSystem(FS, "status_alias/status.go"),
			Config:              failFastLegacyConfig,
			ExpectedFiles:       []string{"status/statuses_enums.go"},
			RepresentationCount: 1,
		},
		{
			Name:                "planets with attributes - default",
			Source:              source.FromFileSystem(FS, "planets/planets.go"),
			Config:              defaultConfig,
			ExpectedFiles:       []string{"planets/planets_enums.go"},
			RepresentationCount: 1,
		},
		{
			Name:                "planets with attributes - failfast & legacy",
			Source:              source.FromFileSystem(FS, "planets/planets.go"),
			Config:              failFastLegacyConfig,
			ExpectedFiles:       []string{"planets/planets_enums.go"},
			RepresentationCount: 1,
		},
		{
			Name:                "planets gravity only - default",
			Source:              source.FromFileSystem(FS, "planets_gravity_only/planets.go"),
			Config:              defaultConfig,
			ExpectedFiles:       []string{"planets_gravity_only/planets_enums.go"},
			RepresentationCount: 1,
		},
		{
			Name:                "planets gravity only - failfast & legacy",
			Source:              source.FromFileSystem(FS, "planets_gravity_only/planets.go"),
			Config:              failFastLegacyConfig,
			ExpectedFiles:       []string{"planets_gravity_only/planets_enums.go"},
			RepresentationCount: 1,
		},
		{
			Name:                "planets simple - default",
			Source:              source.FromFileSystem(FS, "planets_simple/planets.go"),
			Config:              defaultConfig,
			ExpectedFiles:       []string{"planets_simple/planets_enums.go"},
			RepresentationCount: 1,
		},
		{
			Name:                "planets simple - failfast & legacy",
			Source:              source.FromFileSystem(FS, "planets_simple/planets.go"),
			Config:              failFastLegacyConfig,
			ExpectedFiles:       []string{"planets_simple/planets_enums.go"},
			RepresentationCount: 1,
		},
		{
			Name:                "sale - default",
			Config:              defaultConfig,
			Source:              source.FromFileSystem(FS, "sale/sale.go"),
			ExpectedFiles:       []string{"sale/sales_enums.go"},
			RepresentationCount: 1,
		},
		{
			Name:                "sale - failfast & legacy",
			Config:              failFastLegacyConfig,
			Source:              source.FromFileSystem(FS, "sale/sale.go"),
			ExpectedFiles:       []string{"sale/sales_enums.go"},
			RepresentationCount: 1,
		},
		{
			Name:                "tickets - default",
			Source:              source.FromFileSystem(FS, "tickets/tickets.go"),
			Config:              defaultConfig,
			ExpectedFiles:       []string{"tickets/tickets_enums.go"},
			RepresentationCount: 1,
		},
		{
			Name:                "tickets - failfast & legacy",
			Source:              source.FromFileSystem(FS, "tickets/tickets.go"),
			Config:              failFastLegacyConfig,
			ExpectedFiles:       []string{"tickets/tickets_enums.go"},
			RepresentationCount: 1,
		},
		{
			Name:                "orders - default",
			Source:              source.FromFileSystem(FS, "orders/orders.go"),
			Config:              defaultConfig,
			ExpectedFiles:       []string{"orders/orders_enums.go"},
			RepresentationCount: 1,
		},
		{
			Name:                "orders - failfast & legacy",
			Source:              source.FromFileSystem(FS, "orders/orders.go"),
			Config:              failFastLegacyConfig,
			ExpectedFiles:       []string{"orders/orders_enums.go"},
			RepresentationCount: 1,
		},
		{
			Name:                "multiple in one file - default",
			Source:              source.FromFileSystem(FS, "multiple/multiple.go"),
			Config:              defaultConfig,
			ExpectedFiles:       []string{"multiple/statuses_enums.go", "multiple/orders_enums.go"},
			RepresentationCount: 2,
		},
		{
			Name:                "multiple in one file - failfast & legacy",
			Source:              source.FromFileSystem(FS, "multiple/multiple.go"),
			Config:              failFastLegacyConfig,
			ExpectedFiles:       []string{"multiple/statuses_enums.go", "multiple/orders_enums.go"},
			RepresentationCount: 2,
		},
		{
			Name:   "non-existent file",
			Source: source.FromFileSystem(FS, "non/existent/file.go"),
			Config: defaultConfig,
			Err:    fs.ErrNotExist,
		},
	}
)
