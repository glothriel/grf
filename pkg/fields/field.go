package fields

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/models"
)

type RepresentationFunc func(models.InternalValue, string, *gin.Context) (any, error)
type InternalValueFunc func(map[string]any, string, *gin.Context) (any, error)

type ErrorFieldIsNotPresentInPayload struct {
	name string
}

func (e ErrorFieldIsNotPresentInPayload) Error() string {
	return "Field `" + e.name + "` is not present in the payload"
}

func NewErrorFieldIsNotPresentInPayload(name string) ErrorFieldIsNotPresentInPayload {
	return ErrorFieldIsNotPresentInPayload{name: name}
}

type Field interface {
	Name() string
	ToRepresentation(models.InternalValue, *gin.Context) (any, error)
	ToInternalValue(map[string]any, *gin.Context) (any, error)

	IsReadable() bool
	IsWritable() bool

	WithReadOnly() Field
	WithWriteOnly() Field
	WithReadWrite() Field

	WithRepresentationFunc(RepresentationFunc) Field
	WithInternalValueFunc(InternalValueFunc) Field
}

type ConcreteField[Model any] struct {
	name               string
	representationFunc RepresentationFunc
	internalValueFunc  InternalValueFunc

	Readable bool
	Writable bool
}

func (s *ConcreteField[Model]) Name() string {
	return s.name
}

func (s *ConcreteField[Model]) ToRepresentation(intVal models.InternalValue, ctx *gin.Context) (any, error) {
	return s.representationFunc(intVal, s.name, ctx)
}

func (s *ConcreteField[Model]) ToInternalValue(reprModel map[string]any, ctx *gin.Context) (any, error) {
	return s.internalValueFunc(reprModel, s.name, ctx)
}

func (s *ConcreteField[Model]) WithReadOnly() Field {
	s.Readable = true
	s.Writable = false
	return s
}

func (s *ConcreteField[Model]) WithWriteOnly() Field {
	s.Readable = false
	s.Writable = true
	return s
}

func (s *ConcreteField[Model]) WithReadWrite() Field {
	s.Readable = true
	s.Writable = true
	return s
}

func (s *ConcreteField[Model]) IsReadable() bool {
	return s.Readable
}

func (s *ConcreteField[Model]) IsWritable() bool {
	return s.Writable
}

func (s *ConcreteField[Model]) WithRepresentationFunc(f RepresentationFunc) Field {
	s.representationFunc = f
	return s
}

func (s *ConcreteField[Model]) WithInternalValueFunc(f InternalValueFunc) Field {
	s.internalValueFunc = f
	return s
}

func NewField[Model any](name string) Field {
	return &ConcreteField[Model]{
		name: name,
		representationFunc: func(intVal models.InternalValue, name string, ctx *gin.Context) (any, error) {
			return intVal[name], nil
		},
		internalValueFunc: func(reprModel map[string]any, name string, ctx *gin.Context) (any, error) {
			return reprModel[name], nil
		},
		Readable: true,
		Writable: true,
	}
}

func StaticValue[Model any](v any) func(oldField Field) {
	return func(oldField Field) {
		oldField.WithInternalValueFunc(
			func(m map[string]any, s string, ctx *gin.Context) (any, error) {
				return v, nil
			},
		)
	}
}
