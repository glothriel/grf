# Gin rest framework

[![codecov](https://codecov.io/gh/glothriel/grf/branch/master/graph/badge.svg?token=RG7Q17TT73)](https://codecov.io/gh/glothriel/grf)

A REST framework for golang, highly inspired by Django REST framework. Main goal of this project is to achieve ease of implementing REST services similar to DRF with similar APIs but with better performance.

## DRF comparison

### Feature - wise

GRF is way simpler than DRF, mostly because it misses the "D" part - it doesn't have django, so all the goodies coming with Django won't be available. It also is quite fresh project and it doesn't support a lot of stuff DRF has, mainly:

* authentication and authorization (you can write some ie. JWT validating middleware, but no built in support)
* non-JSON responses
* HTML views
* Django admin - like UI

### Performance


TLDR for the results: GRF is able to withstand 4x more reqps than the best DRF scenario, while using 2x less CPU, 10x less Memory, 4x smaller docker image, and having roughly the same latency.

The benchmark was performed using the following scenario:

* Single model and ModelViewset in DRF / ModelView in GRF (pkg/examples/products.go)
* Local sqlite database saved in /tmp
* Perform load test by spamming list endpoint with 3 products (k6/test.js)
* Docker image based on alpine (GRF, glibc is required for sqlite) and distroless (DRF) started with `docker run`
* A single Standard_D4_v3 (4CPUs 16GB mem) Azure VM used
* k6 was launched on the same machine with `--vus 10 --duration 60s` for each scenario

| scenario | avg reqps | latency p(90) | latency p(95) | avg number of cores used | used memory | docker image size |
|----------|:-------------:|------:|------:|------:|------:|------:|
| GRF ModelListCreateView | 4067.983936 | 4.639916 | 6.5315644 | 1.686365217 | 18.3 | 30MB |
| DRF ModelViewset + gunicorn --worker-class=gevent --workers 4 --threads 4 | 1091.596524 | 4.4614407 | 6.09462 | 3.44176087 | 176 | 120MB |
| DRF ModelViewset + gunicorn --worker-class=gevent --workers 4 | 1065.373294 | 4.415479 | 4.777143 | 3.436626087 | 176.7 | 120MB |
| DRF ModelViewset + gunicorn --workers 4 --threads 2 | 1025.58744 | 16.321055 | 18.60752425 | 3.348573913 | 169.2 | 120MB |
| DRF ModelViewset + gunicorn --workers 4 --threads 4 | 983.6514014 | 16.257404 | 19.10055325 | 3.371121739 | 170.9 | 120MB |
| DRF ModelViewset + gunicorn --workers 8 --threads 2 | 969.2663741 | 19.25858 | 23.3209336 | 3.489 | 287.7 | 120MB |
| DRF ModelViewset + gunicorn --worker-class=gevent --workers 8 | 956.3142102 | 16.9649319 | 20.019959 | 3.497873913 | 321.5 | 120MB |
| DRF ModelViewset + gunicorn --worker-class=gevent --workers 8 --threads 2 | 939.6444841 | 17.1753752 | 20.0940822 | 3.502413043 | 321.8 | 120MB |
| DRF ModelViewset + gunicorn --worker-class=gevent --workers 2 --threads 8 | 874.5320259 | 27.934995 | 40.766281 | 2.027017391 | 102.6 | 120MB |
| DRF ModelViewset + gunicorn --worker-class=gevent --workers 2 | 859.6726658 | 28.636555 | 41.54380255 | 2.020104348 | 102.6 | 120MB |

Other gunicorn flag variants had even lower performance.


# Roadmap

* ~~Add rich context to serializer and fields (ability to extract gin request data directly from Field)~~
    * ~~Unlocks creating fields based on non-body data, gin context, etc - like `CurrentlyLoggedInUser` from Authorization header or so~~
* ~~Add support for setting up middleware (fore/after request is executed), including using Gin middlewares directly~~
* ~~Turn off passthrough if the serializer field type cannot be deduced on startup~~
* Improve unit test coverage
* Documentation portal
* Add support for complex to implement types:
    * time.Time
    * time.Duration
    * `list<int>`
    * pointer fields, for example `*string`
    * sql.Null.* fields, for example `sql.NullString`
    * JSON fields, both as text (sqlite) and as dedicated columns (eg. Postgres)
* Add proper support for custom validators
* Add support for model relations
* Add support for viewsets
* Add support for authentication
* Add support for complex pagination implementations
* Add support for permissions (authorization)
* Add support for automatic OpenAPI spec generation
* Add support for caches
* Add support for throttling
* Add support for output formatters
* Add support for non-JSON types
* Add support for translations (error messages)
* Add testing utils
