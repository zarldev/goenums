---
layout: default
title: Examples
---

Explore different usage patterns and features with practical examples:

- [Basic Examples]({{ '/examples/basic' | relative_url }}) → Core functionality examples
- [Advanced Examples]({{ '/examples/advanced' | relative_url }}) → More complex scenarios and features

## Featured Example: Solar System Planets

This example demonstrates an extended enum type with custom fields:

```go
package solarsystem

type planet int // Gravity float64,RadiusKm float64,MassKg float64,OrbitKm float64

//go:generate goenums planets.go
const (
    unknown planet = iota // invalid
    mercury               // Mercury 0.378,2439.7,3.3e23,57910000
    venus                 // Venus 0.907,6051.8,4.87e24,108200000
    earth                 // Earth 1,6378.1,5.97e24,149600000
    mars                  // Mars 0.377,3389.5,6.42e23,227900000
    jupiter               // Jupiter 2.36,69911,1.90e27,778600000
    saturn                // Saturn 0.916,58232,5.68e26,1433500000
    uranus                // Uranus 0.889,25362,8.68e25,2872500000
    neptune               // Neptune 1.12,24622,1.02e26,4495100000
)
```

After generation, we can use the extended enum type:

```go
earthWeight := 100.0
fmt.Printf("Weight on %s: %.2f kg\n", 
    solarsystem.Planets.MARS, 
    earthWeight * solarsystem.Planets.MARS.Gravity)

// Iterate over all planets
for p := range solarsystem.Planets.All() {
    fmt.Printf("Planet: %s\n", p)
}

// Exhaustive handling
solarsystem.ExhaustivePlanets(func(p solarsystem.Planet) {
    // Process each planet
    switch p {
    case solarsystem.Planets.NEPTUNE:
        // Handle neptune
    }
})
```
For more examples, see the [testdata](https://github.com/zarldev/goenums/tree/main/internal/testdata) directory.