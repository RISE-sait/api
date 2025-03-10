package countries

import "github.com/biter777/countries"

func IsValidAlpha2Code(code string) bool {
	country := countries.ByName(code)
	return country != countries.Unknown
}
