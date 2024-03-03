import datetime
from itertools import chain

import pytest
import requests
from tests.utils import AnyUUID


class RequestTestCase:
    def __init__(self, request, expected_status=None, expected_response=None, skip=False):
        self.request = request
        self.expected_status = expected_status
        self.expected_response = expected_response
        self.skip = skip


class FieldEndpoint:
    def __init__(self, path: str, create_cases=None) -> None:
        self.path = path
        self.create_cases = create_cases or []


NOW = datetime.datetime.now().isoformat()
ENDPOINTS = (
    FieldEndpoint(
        path="/bool_field",
        create_cases=(
            RequestTestCase(
                request={"value": True},
                expected_status=201,
                expected_response={"value": True},
            ),
            RequestTestCase(
                request={"value": False},
                expected_status=201,
                expected_response={"value": False},
            ),
            RequestTestCase(
                request={"value": 1},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": 1.337},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": {"foo": "bar"}},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": [1, 2, 3]},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": [True, False]},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": "hueble"},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": "true"},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": "True"},
                expected_status=400,
            ),
        ),
    ),
    FieldEndpoint(
        path="/string_field",
        create_cases=(
            RequestTestCase(
                request={"value": "hello world"},
                expected_status=201,
                expected_response={"value": "hello world"},
            ),
            RequestTestCase(
                request={"value": 1},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": 1.337},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": {"foo": "bar"}},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": [1, 2, 3]},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": [True, False]},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": True},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": False},
                expected_status=400,
            ),
        ),
    ),
    FieldEndpoint(
        path="/null_string_field",
        create_cases=(
            RequestTestCase(
                skip="Nullable fields are not supported yet.",
                request={"value": "hello world"},
                expected_status=201,
                expected_response={"value": "hello world"},
            ),
            RequestTestCase(
                skip="Nullable fields are not supported yet.",
                request={"value": 1},
                expected_status=400,
            ),
            RequestTestCase(
                skip="Nullable fields are not supported yet.",
                request={"value": 1.337},
                expected_status=400,
            ),
            RequestTestCase(
                skip="Nullable fields are not supported yet.",
                request={"value": {"foo": "bar"}},
                expected_status=400,
            ),
            RequestTestCase(
                skip="Nullable fields are not supported yet.",
                request={"value": [1, 2, 3]},
                expected_status=400,
            ),
            RequestTestCase(
                skip="Nullable fields are not supported yet.",
                request={"value": [True, False]},
                expected_status=400,
            ),
            RequestTestCase(
                skip="Nullable fields are not supported yet.",
                request={"value": True},
                expected_status=400,
            ),
            RequestTestCase(
                skip="Nullable fields are not supported yet.",
                request={"value": False},
                expected_status=400,
            ),
        ),
    ),
    FieldEndpoint(
        path="/int_field",
        create_cases=(
            RequestTestCase(
                request={"value": 1337},
                expected_status=201,
                expected_response={"value": 1337},
            ),
            RequestTestCase(
                request={"value": -1337},
                expected_status=201,
                expected_response={"value": -1337},
            ),
            RequestTestCase(
                request={"value": 13.00},
                expected_status=201,
                expected_response={"value": 13},
            ),
            RequestTestCase(
                request={"value": 1.337},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": {"foo": "bar"}},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": [1, 2, 3]},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": [True, False]},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": True},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": False},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": "hello world"},
                expected_status=400,
            ),
        ),
    ),
    FieldEndpoint(
        path="/uint_field",
        create_cases=(
            RequestTestCase(
                request={"value": 1337},
                expected_status=201,
                expected_response={"value": 1337},
            ),
            RequestTestCase(
                request={"value": 13.00},
                expected_status=201,
                expected_response={"value": 13},
            ),
            RequestTestCase(
                request={"value": -1337},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": 1.337},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": {"foo": "bar"}},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": [1, 2, 3]},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": [True, False]},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": True},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": False},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": "hello world"},
                expected_status=400,
            ),
        ),
    ),
    FieldEndpoint(
        path="/float_field",
        create_cases=(
            RequestTestCase(
                request={"value": 1.337},
                expected_status=201,
                expected_response={"value": 1.337},
            ),
            RequestTestCase(
                request={"value": 1337},
                expected_status=201,
                expected_response={"value": 1337.0},
            ),
            RequestTestCase(
                request={"value": {"foo": "bar"}},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": [1, 2, 3]},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": [True, False]},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": True},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": False},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": "hello world"},
                expected_status=400,
            ),
        ),
    ),
    FieldEndpoint(
        path="/datetime_field",
        create_cases=(
            RequestTestCase(
                skip="Datetime is not yet supported",
                request={"value": NOW},
                expected_status=201,
                expected_response={"value": NOW},
            ),
            RequestTestCase(
                request={"value": 1.337},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": 1.337},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": 1337},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": {"foo": "bar"}},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": [1, 2, 3]},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": [True, False]},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": True},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": False},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": "hello world"},
                expected_status=400,
            ),
        ),
    ),
    FieldEndpoint(
        path="/string_slice_field",
        create_cases=(
            RequestTestCase(
                request={"value": ["hello world"]},
                expected_status=201,
                expected_response={"value": ["hello world"]},
            ),
            RequestTestCase(
                request={"value": 1},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": 1.337},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": {"foo": "bar"}},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": [1, 2, 3]},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": ["1", 2]},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": [True, False]},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": True},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": False},
                expected_status=400,
            ),
        ),
    ),
    FieldEndpoint(
        path="/float_slice_field",
        create_cases=(
            RequestTestCase(
                request={"value": [1.337]},
                expected_status=201,
                expected_response={"value": [1.337]},
            ),
            RequestTestCase(
                request={"value": [1, 2, 3]},
                expected_status=201,
                expected_response={"value": [1, 2, 3]},
            ),
            RequestTestCase(
                request={"value": 1},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": 1.337},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": {"foo": "bar"}},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": ["1", 2]},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": [True, False]},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": True},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": False},
                expected_status=400,
            ),
        ),
    ),
    FieldEndpoint(
        path="/int_slice_field",
        create_cases=(
            RequestTestCase(
                skip="Int slice is not yet supported",
                request={"value": [1]},
                expected_status=201,
                expected_response={"value": [1]},
            ),
            RequestTestCase(
                request={"value": 1},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": 1.337},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": {"foo": "bar"}},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": ["1", 2]},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": [True, False]},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": True},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": False},
                expected_status=400,
            ),
        ),
    ),
    FieldEndpoint(
        path="/map_slice_field",
        create_cases=(
            RequestTestCase(
                request={"value": [{"key": "hello", "value": "world"}]},
                expected_status=201,
                expected_response={"value": [{"key": "hello", "value": "world"}]},
            ),
            RequestTestCase(
                request={"value": {"foo": "bar"}},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": [1, 2, 3]},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": ["1", 2]},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": [True, False]},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": True},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": False},
                expected_status=400,
            ),
        ),
    ),
    FieldEndpoint(
        path="/bool_slice_field",
        create_cases=(
            RequestTestCase(
                request={"value": [True]},
                expected_status=201,
                expected_response={"value": [True]},
            ),
            RequestTestCase(
                request={"value": [False]},
                expected_status=201,
                expected_response={"value": [False]},
            ),
            RequestTestCase(
                request={"value": [1, 2, 3]},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": ["1", 2]},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": ["true", "false"]},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": {"foo": "bar"}},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": True},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": False},
                expected_status=400,
            ),
        ),
    ),
    FieldEndpoint(
        path="/any_slice_field",
        create_cases=(
            RequestTestCase(
                request={"value": [True, "1", 2, 3.45]},
                expected_status=201,
                expected_response={"value": [True, "1", 2, 3.45]},
            ),
            RequestTestCase(
                request={"value": [1, 2, 3]},
                expected_status=201,
            ),
            RequestTestCase(
                request={"value": [{"foo": "bar"}, {"bar": "baz"}, 1, True]},
                expected_status=201,
                expected_response={"value": [{"foo": "bar"}, {"bar": "baz"}, 1, True]},
            ),
            RequestTestCase(
                request={"value": ["true", "false"]},
                expected_status=201,
            ),
            RequestTestCase(
                request={"value": {"foo": "bar"}},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": True},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": False},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": "ads"},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": 1},
                expected_status=400,
            ),
            RequestTestCase(
                request={"value": 1.23},
                expected_status=400,
            ),
        ),
    ),
    FieldEndpoint(
        path="/two_d_string_slice_field",
        create_cases=(
            RequestTestCase(
                skip="2D slices are not yet supported",
                request={"value": [["hello", "world"], ["foo", "bar"]]},
                expected_status=201,
                expected_response={"value": [["hello", "world"], ["foo", "bar"]]},
            ),
        ),
    ),
    FieldEndpoint(
        path="/null_bool_field",
        create_cases=(
            RequestTestCase(
                request={"value": True},
                expected_status=201,
                expected_response={"value": True},
            ),
            RequestTestCase(
                request={"value": None},
                expected_status=201,
                expected_response={"value": None},
            ),
        ),
    ),
    FieldEndpoint(
        path="/null_int16_field",
        create_cases=(
            RequestTestCase(
                request={"value": 1337},
                expected_status=201,
                expected_response={"value": 1337},
            ),
            RequestTestCase(
                request={"value": None},
                expected_status=201,
                expected_response={"value": None},
            ),
        ),
    ),
    FieldEndpoint(
        path="/null_int32_field",
        create_cases=(
            RequestTestCase(
                request={"value": 1337},
                expected_status=201,
                expected_response={"value": 1337},
            ),
            RequestTestCase(
                request={"value": None},
                expected_status=201,
                expected_response={"value": None},
            ),
        ),
    ),
    FieldEndpoint(
        path="/null_int64_field",
        create_cases=(
            RequestTestCase(
                request={"value": 1337},
                expected_status=201,
                expected_response={"value": 1337},
            ),
            RequestTestCase(
                request={"value": None},
                expected_status=201,
                expected_response={"value": None},
            ),
        ),
    ),
    FieldEndpoint(
        path="/null_string_field",
        create_cases=(
            RequestTestCase(
                request={"value": "hello world"},
                expected_status=201,
                expected_response={"value": "hello world"},
            ),
            RequestTestCase(
                request={"value": None},
                expected_status=201,
                expected_response={"value": None},
            ),
        ),
    ),
    FieldEndpoint(
        path="/null_float64_field",
        create_cases=(
            RequestTestCase(
                request={"value": 1.337},
                expected_status=201,
                expected_response={"value": 1.337},
            ),
            RequestTestCase(
                request={"value": None},
                expected_status=201,
                expected_response={"value": None},
            ),
        ),
    ),
    FieldEndpoint(
        path="/null_byte_field",
        create_cases=(
            RequestTestCase(
                request={"value": "a"},
                expected_status=201,
                expected_response={"value": "a"},
            ),
            RequestTestCase(
                request={"value": None},
                expected_status=201,
                expected_response={"value": None},
            ),
        ),
    ),
)


@pytest.mark.parametrize(
    "case",
    list(
        chain.from_iterable(
            [
                [
                    (
                        e.path,
                        case,
                    )
                    for case in e.create_cases
                ]
                for e in ENDPOINTS
            ]
        )
    ),
    ids=[
        f"{case[0]}={case[1].request['value']}({case[1].request['value'].__class__.__name__}): {case[1].expected_status}"  # noqa
        for case in list(
            chain.from_iterable(
                [
                    [
                        (
                            e.path,
                            case,
                        )
                        for case in e.create_cases
                    ]
                    for e in ENDPOINTS
                ]
            )
        )
    ],
)
def test_create_success(server_factory, case):
    with server_factory.create("alltypes") as server:
        field = case[1]
        if field.skip:
            pytest.skip(field.skip)
        response = requests.post(
            f"{server.url}{case[0]}",
            json=field.request,
        )
        try:
            if field.expected_status is not None:
                assert response.status_code == field.expected_status
            if field.expected_response is not None:
                assert strip_created_updated_at(response.json()) == dict(
                    **{
                        "id": AnyUUID(),
                    },
                    **field.expected_response,
                )
        except AssertionError:
            print(response.json())
            raise


def strip_created_updated_at(product_json):
    return dict({k: v for k, v in product_json.items() if k not in ("created_at", "updated_at")})
