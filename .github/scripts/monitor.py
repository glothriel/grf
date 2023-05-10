import json
import subprocess
import sys
import time


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


def get_container_id(image_name):
    for process in docker("ps")[1:]:
        cols = [c.strip() for c in process.split(" ") if c]
        if cols[1] == image_name:
            return cols[0]
    return None


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


def get_chart_html(stats):
    charts_html = ""
    for stat_name, values in stats.items():
        charts_html += f"""
        <div style="width: 800px; height: 400px; margin: 0 20px; position: relative;">
        <h4> {stat_name} </h4>
        <canvas id="{stat_name}"></canvas>
        <script>
        new Chart("{stat_name}", {{
        type: "line",
        "maintainAspectRatio": false,
        data: {{
            labels: {json.dumps([f'{i}s' for i in list(range(len(values)))])},
            datasets: [{{
            backgroundColor:"rgba(0,0,255,1.0)",
            borderColor: "rgba(0,0,255,0.1)",
            data: {json.dumps(values)},
            }}]
        }},
        }});
        </script>
        </div>
        <hr>
        """
    return f"""
    <html>
    <head>
    <script
        src="https://cdnjs.cloudflare.com/ajax/libs/Chart.js/2.9.4/Chart.js">
    </script>
    </head>
    <body>
    {charts_html}
    </body>
    </html>
    """


if __name__ == "__main__":
    assert len(sys.argv) == 3, "Usage: monitor.py <image_name> <output_file>"

    image_name = sys.argv[1]
    container_id = None
    while container_id is None:
        try:
            container_id = get_container_id(image_name)
        except Exception:
            pass

        time.sleep(1)
    # each value in the list contains measaurement for a second
    measurements = {
        "Amount of used CPU cores": [],
        "MB of used memory": [],
    }
    # Print docker stats for the container for as long as it is running
    while True:
        current_stats = get_raw_stats(container_id)

        measurements["Amount of used CPU cores"].append(parse_cpuperc(current_stats["CPUPerc"]))
        measurements["MB of used memory"].append(parse_memusage_to_mb(current_stats["MemUsage"]))
        file_name = sys.argv[2]
        with open(file_name, "w") as f:
            f.write(get_chart_html(measurements))
        time.sleep(1)
