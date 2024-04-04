package milkyway

type planet int // Gravity[float64],RadiusKm[float64],MassKg[float64],OrbitKm[float64],OrbitDays[float64],SurfacePressureBars[float64],Moons[int],Rings[bool]

const (
	unknown planet = iota // invalid
	mercury               // Mercury 0.378,2439.7,3.3e23,57910000,88,0.0000000001,0,false
	venus                 // Venus 0.907,6051.8,4.87e24,108200000,225,92,0,false
	earth                 // Earth 1,6378.1,5.97e24,149600000,365,1,1,false
	mars                  // Mars 0.377,3389.5,6.42e23,227900000,687,0.01,2,false
	jupiter               // Jupiter 2.36,69911,1.90e27,778600000,4333,20,4,true
	saturn                // Saturn 0.916,58232,5.68e26,1433500000,10759,1,7,true
	uranus                // Uranus 0.889,25362,8.68e25,2872500000,30687,1.3,13,true
	neptune               // Neptune 1.12,24622,1.02e26,4495100000,60190,1.5,2,true
)
