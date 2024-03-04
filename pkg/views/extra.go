package views

type ExtraAction[Model any] struct {
	Method       string
	RelativePath string
	Handler      ViewSetHandlerFunc[Model]
}

func NewExtraAction[Model any](method string, relativePath string, handler ViewSetHandlerFunc[Model]) *ExtraAction[Model] {
	return &ExtraAction[Model]{
		Method:       method,
		RelativePath: relativePath,
		Handler:      handler,
	}
}
