package enum_test

import (
	"testing"

	"github.com/zarldev/goenums/enum"
)

func BenchmarkParseValue_Int(b *testing.B) {
	b.ResetTimer()
	for range b.N {
		_, _ = enum.ParseValue("42", 0)
	}
}

func BenchmarkParseValue_String(b *testing.B) {
	b.ResetTimer()
	for range b.N {
		_, _ = enum.ParseValue("hello world", "")
	}
}

func BenchmarkParseValue_Bool(b *testing.B) {
	b.ResetTimer()
	for range b.N {
		_, _ = enum.ParseValue("true", false)
	}
}

func BenchmarkParseEnumAliases(b *testing.B) {
	input := "alias1,alias2,alias3,alias4,alias5"
	b.ResetTimer()
	for range b.N {
		_ = enum.ParseEnumAliases(input)
	}
}

func BenchmarkParseEnumFields(b *testing.B) {
	input := "42,true,hello,3.14"
	enumIota := enum.EnumIota{
		Fields: []enum.Field{
			{Name: "Number", Value: 0},
			{Name: "Flag", Value: false},
			{Name: "Text", Value: ""},
			{Name: "Float", Value: 0.0},
		},
	}
	for b.Loop() {
		_, _ = enum.ParseEnumFields(input, enumIota)
	}
}
