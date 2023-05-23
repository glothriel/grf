import pytest
import requests
from tests.utils import AnyUUID


@pytest.fixture
def some_products(server_factory):
    with server_factory.create("products") as server:
        for product in (
            ("Butter", "Tasty and organic made from milk from happy cows", "13.37"),
            ("Potatoes", "Freshly dug up from the ground", "0.99"),
            ("Bread", "Freshly baked and crispy", "1.99"),
            ("Apples", "Freshly picked from the tree", "21"),
        ):
            assert (
                requests.post(
                    f"{server.url}/products",
                    json={
                        "name": product[0],
                        "description": product[1],
                        "price": product[2],
                    },
                ).status_code
                == 201
            )
        yield server


def test_create_bad_missing_fields(server_factory):
    with server_factory.create("products") as server:
        response = requests.post(f"{server.url}/products", json={"name": "foo"})
        assert response.status_code == 400
        assert list(sorted(response.json()["errors"].keys())) == ["description"]


def test_create_superfluous_fields(server_factory):
    with server_factory.create("products") as server:
        response = requests.post(
            f"{server.url}/products",
            json={
                "name": "foo",
                "description": "bar",
                "superfluous": "baz",
            },
        )
        assert response.status_code == 400
        assert list(sorted(response.json()["errors"].keys())) == ["superfluous"]


def test_create_unexpected_field_type_unparsable_decimal(server_factory):
    with server_factory.create("products") as server:
        response = requests.post(
            f"{server.url}/products",
            json={
                "name": "foo",
                "description": "bar",
                "price": "huehuehue",
            },
        )
        assert response.status_code == 400
        assert list(sorted(response.json()["errors"].keys())) == ["price"]


def test_create_unexpected_field_type_name_as_float(server_factory):
    with server_factory.create("products") as server:
        response = requests.post(
            f"{server.url}/products",
            json={
                "name": 1.0,
                "description": "bar",
                "price": "13.37",
            },
        )
        assert response.status_code == 400
        assert list(sorted(response.json()["errors"].keys())) == ["name"]


def test_create_success(server_factory):
    with server_factory.create("products") as server:
        response = requests.post(
            f"{server.url}/products",
            json={
                "name": "foo",
                "description": "bar",
            },
        )
        assert strip_created_updated_at(response.json()) == {
            "id": AnyUUID(),
            "name": "foo",
            "description": "bar",
            "price": "0",
        }
        assert response.status_code == 201


def test_list(some_products):
    response = requests.get(f"{some_products.url}/products")
    assert response.status_code == 200
    assert list(sorted(response.json(), key=lambda x: x["name"])) == [
        {
            "id": AnyUUID(),
            "name": "Apples",
        },
        {
            "id": AnyUUID(),
            "name": "Bread",
        },
        {
            "id": AnyUUID(),
            "name": "Butter",
        },
        {
            "id": AnyUUID(),
            "name": "Potatoes",
        },
    ]


def test_retrieve(some_products):
    product = requests.get(f"{some_products.url}/products").json()[0]
    response = requests.get(f"{some_products.url}/products/{product['id']}")
    assert response.status_code == 200
    assert strip_created_updated_at(response.json()) == {
        "id": product["id"],
        "name": "Apples",
        "description": "Freshly picked from the tree",
        "price": "21",
    }


def test_update(some_products):
    product = requests.get(f"{some_products.url}/products").json()[0]

    response = requests.put(
        f"{some_products.url}/products/{product['id']}",
        json={
            "name": "updatedfoo",
            "description": "updatedbar",
        },
    )
    assert strip_created_updated_at(response.json()) == {
        "id": product["id"],
        "name": "updatedfoo",
        "description": "updatedbar",
        "price": "21",
    }
    assert response.status_code == 200


def test_delete(some_products):
    product = requests.get(f"{some_products.url}/products").json()[0]
    response = requests.delete(f"{some_products.url}/products/{product['id']}")
    assert response.status_code == 204
    assert response.content == b""
    assert requests.get(f"{some_products.url}/products/{product['id']}").status_code == 404


def strip_created_updated_at(product_json):
    return dict({k: v for k, v in product_json.items() if k not in ("created_at", "updated_at")})
