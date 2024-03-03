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

type Field[Model any] struct {
	name               string
	representationFunc RepresentationFunc
	internalValueFunc  InternalValueFunc

	Readable bool
	Writable bool
}

func (s *Field[Model]) Name() string {
	return s.name
}

func (s *Field[Model]) ToRepresentation(intVal models.InternalValue, ctx *gin.Context) (any, error) {
	return s.representationFunc(intVal, s.name, ctx)
}

func (s *Field[Model]) ToInternalValue(reprModel map[string]any, ctx *gin.Context) (any, error) {
	return s.internalValueFunc(reprModel, s.name, ctx)
}

func (s *Field[Model]) ReadOnly() *Field[Model] {
	s.Readable = true
	s.Writable = false
	return s
}

func (s *Field[Model]) WriteOnly() *Field[Model] {
	s.Readable = false
	s.Writable = true
	return s
}

func (s *Field[Model]) ReadWrite() *Field[Model] {
	s.Readable = true
	s.Writable = true
	return s
}

func (s *Field[Model]) WithRepresentationFunc(f RepresentationFunc) *Field[Model] {
	s.representationFunc = f
	return s
}

func (s *Field[Model]) WithInternalValueFunc(f InternalValueFunc) *Field[Model] {
	s.internalValueFunc = f
	return s
}

func NewField[Model any](name string) *Field[Model] {
	return &Field[Model]{
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

func StaticValue[Model any](v any) func(oldField *Field[Model]) {
	return func(oldField *Field[Model]) {
		oldField.WithInternalValueFunc(
			func(m map[string]any, s string, ctx *gin.Context) (any, error) {
				return v, nil
			},
		)
	}
}
