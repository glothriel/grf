import json
import shlex
import subprocess
import sys
import threading
import time

import requests

setups = [
    ""
    # "--worker-class=gevent --worker-connections=5000 --workers 8",
    # "--worker-class=gevent --worker-connections=5000 --workers 32",
    # '--worker-class=gevent --worker-connections=5000 --workers 64',
    # '--worker-class=gevent --worker-connections=1000 --workers 8',
    # '--worker-class=gevent --worker-connections=1000 --workers 32',
    # '--worker-class=gevent --worker-connections=1000 --workers 64',
    # '--workers 8',
    # '--workers 32',
    # '--workers 64',
    # '--workers 8 --threads 2',
    # '--workers 8 --threads 4',
    # '--workers 8 --threads 8',
    # '--workers 16 --threads 2',
    # '--workers 16 --threads 4',
    # '--workers 32 --threads 2',
]

results = dict()


def r(process, **kwargs):
    print(f">>> {process}")
    return subprocess.run(shlex.split(process), **kwargs)


def docker(*command):
    cmd = ["docker"] + list(command)
    print(f">>> {cmd}")
    try:
        p = subprocess.run(cmd, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        return [ln for ln in p.stdout.decode().split("\n") if ln]
    except Exception:
        print(p.stdout.decode())
        print(p.stderr.decode())
        raise


def get_raw_stats(container_id):
    cols = [c.strip() for c in docker("stats", "--no-stream", container_id)[1].split(" ") if c]
    return {
        "CPUPerc": cols[2],
        "MemUsage": cols[3],
    }


def parse_cpuperc(s):
    return float(s.rstrip("%")) / 100


def parse_memusage_to_mb(s):
    s = s.split("/")[0].strip()
    if s.endswith("MiB"):
        return float(s.rstrip("MiB"))
    elif s.endswith("GiB"):
        return float(s.rstrip("GiB")) * 1024
    else:
        raise Exception(f"Unexpected memusage value: {s}")


def empty_measurements():
    return {
        "Amount of used CPU cores": [],
        "MB of used memory": [],
    }


def collect_metrics_in_the_background(container_id, measurements, stop):
    # Print docker stats for the container for as long as it is running
    while True:
        current_stats = get_raw_stats(container_id)

        measurements["Amount of used CPU cores"].append(parse_cpuperc(current_stats["CPUPerc"]))
        measurements["MB of used memory"].append(parse_memusage_to_mb(current_stats["MemUsage"]))
        if stop():
            return
        time.sleep(1)


def wait_for_port(port):
    while True:
        try:
            requests.get(f"http://localhost:{port}/asdfasdxczc1231412").status_code == 404
            break
        except Exception:
            pass
        time.sleep(0.1)


if __name__ == "__main__":
    assert len(sys.argv) == 2, "Usage: monitor.py <image_name>"

    image_name = sys.argv[1]

    r(f"docker build . -t {image_name}")

    for setup in setups:
        try:
            p = r(
                f"docker run -d -p 8080:8080 {image_name} --db=/tmp/db.sqlite3",
                stdout=subprocess.PIPE,
            )
            container_id = p.stdout.decode().strip()
            wait_for_port(8080)

            # Give it some time, gunicorn opens the ports before workers are booted
            time.sleep(5)

            stopped = False
            metrics = empty_measurements()
            metrics_thread = threading.Thread(
                target=collect_metrics_in_the_background,
                args=(container_id, metrics, lambda: stopped),
            )
            metrics_thread.start()

            r("k6 run tests/k6/test.js --vus 10 --duration 10s")

            stopped = True
            metrics_thread.join()

            with open("artifacts/benchmark.json") as f:
                report = json.loads(f.read())
                results[setup] = {
                    "reqps": report["metrics"]["http_reqs"]["values"]["rate"],
                    "latency": report["metrics"]["http_req_duration"]["values"],
                    "resources": metrics,
                }

        finally:
            try:
                the_container_id = container_id
            except NameError:
                the_container_id = p.stdout.decode().strip()
            r(f"docker rm -f {the_container_id}")

    print(json.dumps(results, indent=2))
