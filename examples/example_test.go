package examples_test

import (
	"encoding/json"
	"fmt"

	// These would be your generated enum packages
	"github.com/zarldev/goenums/examples/solarsystem"
	"github.com/zarldev/goenums/examples/validation"
)

// Example showing how to access enum values
func Example_basicUsage() {
	// Access enum constants safely
	myStatus := validation.Statuses.BOOKED

	// Convert to string
	fmt.Println("Status:", myStatus.String())

	// Parse from string
	parsed, err := validation.ParseStatus("SKIPPED")
	if err != nil {
		fmt.Println("Error:", err)
	}

	// Validate enum values
	if !parsed.IsValid() {
		fmt.Println("Invalid status")
	} else {
		fmt.Println("Valid status:", parsed)
	}

	// Output:
	// Status: BOOKED
	// Valid status: SKIPPED
}

// Example showing extended enum types with custom fields
func Example_extendedTypes() {
	earthWeight := 100.0
	mars := solarsystem.Planets.MARS

	// Access custom fields on the enum
	fmt.Printf("Weight on %s: %.2f kg (gravity: %.3f)\n",
		mars,
		earthWeight*mars.Gravity,
		mars.Gravity)

	// Output:
	// Weight on Mars: 37.70 kg (gravity: 0.377)
}

// Example showing iteration over enum values
func Example_iteration() {
	// Using slice-based access
	fmt.Println("All statuses:")
	for s := range validation.Statuses.All() {
		fmt.Printf("- %s\n", s)
	}

	// Output:
	// All statuses:
	// - UNKNOWN
	// - FAILED
	// - PASSED
	// - SKIPPED
	// - SCHEDULED
	// - RUNNING
	// - BOOKED
}

// Example showing JSON marshaling/unmarshaling
func Example_jsonHandling() {
	// Create a struct with enum fields
	type Task struct {
		ID     int               `json:"id"`
		Status validation.Status `json:"status"`
	}

	// Create a task with enum value
	task := Task{
		ID:     123,
		Status: validation.Statuses.RUNNING,
	}

	// Marshal to JSON
	data, err := json.Marshal(task)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("JSON:", string(data))

	// Unmarshal from JSON
	var newTask Task
	jsonData := `{"id":456,"status":"RUNNING"}`
	_ = json.Unmarshal([]byte(jsonData), &newTask)
	fmt.Printf("Unmarshaled task: ID=%d, Status=%s\n", newTask.ID, newTask.Status)

	// Output:
	// JSON: {"id":123,"status":"RUNNING"}
	// Unmarshaled task: ID=456, Status=RUNNING
}

// Example showing exhaustive handling
func Example_exhaustiveHandling() {
	// Keep track of which status was processed
	processed := make(map[string]bool)

	// Use the exhaustive function to ensure we process every status
	validation.ExhaustiveStatuses(func(s validation.Status) {
		// In a real application, you would handle each status differently
		processed[s.String()] = true
	})

	// Verify all statuses were processed
	allProcessed := true
	for s := range validation.Statuses.All() {
		if !processed[s.String()] {
			allProcessed = false
			fmt.Printf("Status %s was not processed\n", s)
		}
	}

	fmt.Println("All statuses processed:", allProcessed)

	// Output:
	// All statuses processed: true
}

// Example showing how to use enums with database operations (simulated)
func Example_databaseIntegration() {
	// In a real application, this would be database code
	// Here we simulate scanning from a database row

	// Simulate scanning a string value from DB
	var planetEnum solarsystem.Planet
	dbValue := "Mars" // Value from database

	// Scan the value (implements sql.Scanner interface)
	err := planetEnum.Scan(dbValue)
	if err != nil {
		fmt.Println("Error:", err)
	}

	fmt.Println("Planet from DB:", planetEnum)
	fmt.Printf("Has rings: %v\n", planetEnum.Rings)

	// Output:
	// Planet from DB: Mars
	// Has rings: false
}
