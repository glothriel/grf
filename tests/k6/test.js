import http from 'k6/http';
import { check } from 'k6';


function cleanAllProducts(){
  var items = JSON.parse(http.get('http://localhost:8080/products').body);
  items.forEach(item => {
    var res = http.del(`http://localhost:8080/products/${item.id}`);
    check(res, {
      'product is deleted': (r) => r.status === 204,
    })
  });
}

export function setup() {
  cleanAllProducts();
  var items = JSON.parse(http.get('http://localhost:8080/products').body);
  if (items.length < 3) {
    var cat = http.post("http://localhost:8080/categories", JSON.stringify(
      {
        "name": "Category",
        "description": "This is wonderful category."
      }, { headers: { 'Content-Type': 'application/json' }})
    )
    for(var i = items.length; i < 3; i++) {
      var res = http.post('http://localhost:8080/products', JSON.stringify({
        name: `Product ${i}`,
        description: `Description ${i}`,
        price: "10.00",
        category_id: cat.json().id
      }), { headers: { 'Content-Type': 'application/json' }});
      check(res, {
        'product is created': (r) => r.status === 201,
      })
    }
  }
}

export default function () {
  http.get('http://localhost:8080/products');
}

export function teardown() {
  cleanAllProducts();
}

// This will export to HTML as filename "result.html" AND also stdout using the text summary
import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";
import { textSummary } from "https://jslib.k6.io/k6-summary/0.0.1/index.js";

export function handleSummary(data) {
  return {
    "artifacts/benchmark.json": JSON.stringify(data),
    stdout: textSummary(data, { indent: " ", enableColors: true }),
  };
}