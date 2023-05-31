
```go
package foo

func main(){

	app := grf.NewApp().ModelView(
		"/foo",
		NewFooModelView[Foo]().WithSerializer(
			NewModelSerializer[Foo]().WithField(
				"foo_field",
				function(f *fields.F){
					f.toRepresentationFunc = func(fv interface{}, grfCtx *grfctx.Context){

					}
				}
			)
		).WithValidator()
	).ModelView(
		"/bar",
		NewBarModelView[Foo]().View(func(v views.V) views.V {
			return v.WithAuthentication(
				jwt.JWTAuthentication()
			).WithThrottling(
				sdg
			)
		})
	).View(
		"/login",
		LoginView(),
	)
}
```

ModelView:
* WithSerializer
* WithPagination
* WithFilter

App:
* WithValidator -> Serializer???? japrldl
* WithPagination -> ModelView :ok:

Może zamiast tego.... Copy?
