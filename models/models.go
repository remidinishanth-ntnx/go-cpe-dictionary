package models

import (
	"github.com/jinzhu/gorm"
)

// CategorizedCpe :
// https://cpe.mitre.org/specification/CPE_2.3_for_ITSAC_Nov2011.pdf
type CategorizedCpe struct {
	gorm.Model `json:"-" xml:"-"`
	Title      string
	CpeURI     string
	CpeFS      string
}
