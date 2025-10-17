package vulpo

// JulianToYMD converts Julian day number to Year, Month, Day using the proper algorithm
// This matches the algorithm from astronomical sources and the mkfdbf C library
func JulianToYMD(jd int) (year, month, day int) {
	// Algorithm from "Numerical Recipes in C" and astronomical sources
	// This is the standard Julian to Gregorian calendar conversion

	a := jd + 32044
	b := (4*a + 3) / 146097
	c := a - (146097*b)/4
	d := (4*c + 3) / 1461
	e := c - (1461*d)/4
	m := (5*e + 2) / 153

	day = e - (153*m+2)/5 + 1
	month = m + 3 - 12*(m/10)
	year = 100*b + d - 4800 + m/10

	return year, month, day
}
