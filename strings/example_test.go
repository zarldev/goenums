package strings_test

import (
	"fmt"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/strings"
)

// Example demonstrates basic string handling utilities for enum generation.
func Example_basic() {
	// Convert to plural forms
	fmt.Println("Basic pluralization:")
	fmt.Println("status ->", strings.Plural("status"))
	fmt.Println("dog ->", strings.Plural("dog"))
	fmt.Println("box ->", strings.Plural("box"))
	fmt.Println("city ->", strings.Plural("city"))

	// Output:
	// Basic pluralization:
	// status -> statuses
	// dog -> dogs
	// box -> boxes
	// city -> cities
}

// Example_irregularPlurals shows how the package handles irregular plurals.
func Example_irregularPlurals() {
	fmt.Println("Irregular pluralization:")
	fmt.Println("child ->", strings.Plural("child"))
	fmt.Println("person ->", strings.Plural("person"))
	fmt.Println("tooth ->", strings.Plural("tooth"))
	fmt.Println("mouse ->", strings.Plural("mouse"))
	fmt.Println("index ->", strings.Plural("index"))

	// Output:
	// Irregular pluralization:
	// child -> children
	// person -> people
	// tooth -> teeth
	// mouse -> mice
	// index -> indices
}

// Example_compoundPlurals demonstrates pluralization of compound words.
func Example_compoundPlurals() {
	fmt.Println("Compound word pluralization:")
	fmt.Println("dog_house ->", strings.Plural("dog_house"))
	fmt.Println("status_code ->", strings.Plural("status_code"))
	fmt.Println("mouse_trap ->", strings.Plural("mouse_trap"))

	// Output:
	// Compound word pluralization:
	// dog_house -> dog_houses
	// status_code -> status_codes
	// mouse_trap -> mouse_traps
}

// Example_camelCase demonstrates converting strings to camel case.
func Example_camelCase() {
	fmt.Println("Camel case conversion:")
	fmt.Println("hello_world ->", strings.CamelCase("hello_world"))
	fmt.Println("dog_house ->", strings.CamelCase("dog_house"))
	fmt.Println("DOG_HOUSE ->", strings.CamelCase("DOG_HOUSE"))
	fmt.Println("status ->", strings.CamelCase("status"))
	fmt.Println("HTTP_response ->", strings.CamelCase("HTTP_response"))

	// Output:
	// Camel case conversion:
	// hello_world -> HelloWorld
	// dog_house -> DogHouse
	// DOG_HOUSE -> DogHouse
	// status -> Status
	// HTTP_response -> HttpResponse
}

// Example_pluralAndCamel demonstrates getting both plural and camel case forms.
func Example_pluralAndCamel() {
	fmt.Println("Plural and camel case:")

	plural, camelPlural := strings.PluralAndCamelPlural("status")
	fmt.Println("status ->", plural, ",", camelPlural)

	plural, camelPlural = strings.PluralAndCamelPlural("user_profile")
	fmt.Println("user_profile ->", plural, ",", camelPlural)

	plural, camelPlural = strings.PluralAndCamelPlural("HTTP_CODE")
	fmt.Println("HTTP_CODE ->", plural, ",", camelPlural)

	// Output:
	// Plural and camel case:
	// status -> statuses , Statuses
	// user_profile -> user_profiles , UserProfiles
	// HTTP_CODE -> HTTP_CODES , HttpCodes
}

// Example_enumBuilder demonstrates using the EnumBuilder for efficient string building.
func Example_enumBuilder() {
	// Create some sample enum representation
	representation := enum.Representation{
		Enums: []enum.Enum{
			{
				Info: enum.Info{
					Upper: "RED",
					Alias: "Red",
				},
			},
			{
				Info: enum.Info{
					Upper: "GREEN",
					Alias: "Green",
				},
			},
			{
				Info: enum.Info{
					Upper: "BLUE",
					Alias: "Blue",
				},
			},
		},
	}

	// Create a new EnumBuilder with preallocated buffer
	builder := strings.NewEnumBuilder(representation)

	// Build a string representation
	builder.WriteString("// Color enum constants\n")
	builder.WriteString("const (\n")

	for _, e := range representation.Enums {
		builder.WriteString("\t")
		builder.WriteString(e.Info.Upper)
		builder.WriteString(" = ")
		builder.WriteString("\"")
		builder.WriteString(e.Info.Alias)
		builder.WriteString("\"\n")
	}

	builder.WriteString(")")

	fmt.Println(builder.String())

	// Output:
	// // Color enum constants
	// const (
	// 	RED = "Red"
	// 	GREEN = "Green"
	// 	BLUE = "Blue"
	// )
}

// Example_utilityWrappers shows how the package wraps standard library functions.
func Example_utilityWrappers() {
	// Using wrapped string utility functions
	text := "  hello_world  "

	fmt.Println("Original:", fmt.Sprintf("[%s]", text))
	fmt.Println("TrimSpace:", fmt.Sprintf("[%s]", strings.TrimSpace(text)))
	// Use string literal to capture exact spaces
	fmt.Printf("ToUpper: [%s]\n", strings.ToUpper(text))
	fmt.Println("ToLower:", strings.ToLower("HELLO"))
	fmt.Println("Contains 'world':", strings.Contains(text, "world"))
	fmt.Println("Split by '_':", strings.Split(strings.TrimSpace(text), "_"))
	fmt.Println("Join with '-':", strings.Join([]string{"a", "b", "c"}, "-"))

	// Output:
	// Original: [  hello_world  ]
	// TrimSpace: [hello_world]
	// ToUpper: [  HELLO_WORLD  ]
	// ToLower: hello
	// Contains 'world': true
	// Split by '_': [hello world]
	// Join with '-': a-b-c
}
