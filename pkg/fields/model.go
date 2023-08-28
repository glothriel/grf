package fields

type GRFRepresentable interface {
	ToRepresentation() (any, error)
}

type GRFParsable interface {
	FromRepresentation(any) error
}
