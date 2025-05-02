package main

import (
	"fmt"

	sale "github.com/zarldev/goenums/internal/testdata/plural"
)

func main() {
	sale.ExhaustiveDiscountTypes(func(dt sale.DiscountType) {
		fmt.Printf("Name: %v\n", dt)
		fmt.Printf("Available: %v\n", dt.Available)
		fmt.Printf("Started: %v\n", dt.Started)
		fmt.Printf("Finished: %v\n", dt.Finished)
		fmt.Printf("Cancelled: %v\n", dt.Cancelled)
		fmt.Printf("Duration: %v\n", dt.Duration)
		fmt.Println("---")
	})
}
