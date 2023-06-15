package fields

import (
	"database/sql"
	"database/sql/driver"
)

type GRFRepresentable interface {
	ToRepresentation() (any, error)
}

type GRFParsable interface {
	FromRepresentation(any) error
}

type ModelField[Model any] interface {
	GRFRepresentable
	GRFParsable
	sql.Scanner
	driver.Valuer
}
