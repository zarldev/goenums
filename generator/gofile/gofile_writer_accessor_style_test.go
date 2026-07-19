package gofile_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/file"
	"github.com/zarldev/goenums/generator/config"
	"github.com/zarldev/goenums/generator/gofile"
)

func TestWriter_AccessorStyleGo(t *testing.T) {
	t.Parallel()
	fs := file.NewMemFS()
	request := benchmarkRequest(1)
	request.Configuration.AccessorStyle = config.AccessorStyleGo
	request.EnumIota.Enums = []enum.Enum{
		{Name: "pendingApproval", Value: 0, Valid: true, Aliases: []string{"Pending approval"}},
		{Name: "apiError", Value: 1, Valid: true, Aliases: []string{"API error"}},
		{Name: "http2Enabled", Value: 2, Valid: true, Aliases: []string{"HTTP2 enabled"}},
		{Name: "userID", Value: 3, Valid: true, Aliases: []string{"User ID"}},
	}

	writer := gofile.NewWriter(gofile.WithFileSystem(fs))
	if err := writer.Write(context.Background(), []enum.GenerationRequest{request}); err != nil {
		t.Fatal(err)
	}
	content, err := fs.ReadFile("statuses_enums.go")
	if err != nil {
		t.Fatal(err)
	}
	generated := string(content)
	for _, want := range []string{
		"PendingApproval",
		"APIError",
		"HTTP2Enabled",
		"UserID",
		"Statuses.PendingApproval",
		"Statuses.APIError",
		"Statuses.HTTP2Enabled",
		"Statuses.UserID",
		`"Pending approval"`,
		`"API error"`} {
		if !strings.Contains(generated, want) {
			t.Errorf("generated output missing %q", want)
		}
	}
	for _, unwanted := range []string{
		"PENDINGAPPROVAL",
		"APIERROR",
		"HTTP2ENABLED",
		"USERID",
	} {
		if strings.Contains(generated, unwanted) {
			t.Errorf("generated output unexpectedly contains %q", unwanted)
		}
	}
}

func TestWriter_AccessorStyleExplicitUpperMatchesDefault(t *testing.T) {
	t.Parallel()
	defaultFS := file.NewMemFS()
	upperFS := file.NewMemFS()
	defaultReq := benchmarkRequest(1)
	upperReq := benchmarkRequest(1)
	upperReq.Configuration.AccessorStyle = config.AccessorStyleUpper

	if err := gofile.NewWriter(gofile.WithFileSystem(defaultFS)).Write(context.Background(), []enum.GenerationRequest{defaultReq}); err != nil {
		t.Fatal(err)
	}
	if err := gofile.NewWriter(gofile.WithFileSystem(upperFS)).Write(context.Background(), []enum.GenerationRequest{upperReq}); err != nil {
		t.Fatal(err)
	}
	defaultContent, err := defaultFS.ReadFile("statuses_enums.go")
	if err != nil {
		t.Fatal(err)
	}
	upperContent, err := upperFS.ReadFile("statuses_enums.go")
	if err != nil {
		t.Fatal(err)
	}
	if string(defaultContent) != string(upperContent) {
		t.Fatal("explicit upper accessor style differed from default output")
	}
}

func TestWriter_AccessorStyleCollisions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		enums []enum.Enum
		want  error
	}{
		{
			name: "go style duplicate",
			enums: []enum.Enum{
				{Name: "pendingApproval", Value: 0, Valid: true},
				{Name: "pending_approval", Value: 1, Valid: true},
			},
			want: gofile.ErrAccessorCollision,
		},
		{
			name: "upper style duplicate",
			enums: []enum.Enum{
				{Name: "pending", Value: 0, Valid: true},
				{Name: "PENDING", Value: 1, Valid: true},
			},
			want: gofile.ErrAccessorCollision,
		},
		{
			name: "reserved All",
			enums: []enum.Enum{
				{Name: "all", Value: 0, Valid: true},
			},
			want: gofile.ErrAccessorCollision,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fs := file.NewMemFS()
			request := benchmarkRequest(1)
			request.Configuration.AccessorStyle = config.AccessorStyleGo
			request.EnumIota.Enums = tt.enums
			if tt.name == "upper style duplicate" {
				request.Configuration.AccessorStyle = config.AccessorStyleUpper
			}
			err := gofile.NewWriter(gofile.WithFileSystem(fs)).Write(context.Background(), []enum.GenerationRequest{request})
			if !errors.Is(err, tt.want) {
				t.Fatalf("Write() error = %v, want %v", err, tt.want)
			}
			if _, readErr := fs.ReadFile("statuses_enums.go"); readErr == nil {
				t.Fatal("Write() created output despite accessor error")
			}
		})
	}
}
