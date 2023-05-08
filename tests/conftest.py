import os
import socket
import subprocess
import tempfile
import time
from contextlib import closing, contextmanager

import pytest
import requests
from tests.utils import run_process


class BinaryBuilder:
    executables = {}

    @classmethod
    def build(cls, example_project_name):
        if example_project_name not in cls.executables:
            cls.executables[example_project_name] = cls._build(example_project_name)
        return cls.executables[example_project_name]

    @classmethod
    def _build(cls, example_project_name):
        tmp = tempfile.NamedTemporaryFile(prefix="grf", delete=False)
        run_process(
            [
                "go",
                "build",
                "-o",
                tmp.name,
                os.path.join("pkg", "examples", example_project_name, "main.go"),
            ]
        )
        return tmp.name

    @classmethod
    def clean(cls):
        for executable in cls.executables.values():
            os.remove(executable)
        cls.executables = {}


@pytest.fixture
def free_port():
    with closing(socket.socket(socket.AF_INET, socket.SOCK_STREAM)) as s:
        s.bind(("", 0))
        s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        return s.getsockname()[1]


class Server:
    def __init__(self, process, url):
        self.process = process
        self.url = url


class ServerFactory:
    def __init__(self, tmpdir, port, binary_builder):
        self.tmpdir = tmpdir
        self.port = str(port)
        self.binary_builder = binary_builder

    @contextmanager
    def create(self, project_name):
        prc = subprocess.Popen(
            [
                self.binary_builder.build(project_name),
                "--port",
                self.port,
                "--db",
                os.path.join(self.tmpdir, "database.db"),
            ],
            env={},
        )
        try:
            base_url = f"http://localhost:{self.port}"
            waited_iterations = 0
            while waited_iterations < 30:
                try:
                    requests.get(f"{base_url}", timeout=3)
                except requests.exceptions.ConnectionError:
                    time.sleep(0.1)
                    waited_iterations += 1
                else:
                    break
            else:
                raise RuntimeError("The server never got up")
            yield Server(prc, base_url)
        finally:
            prc.terminate()


@pytest.fixture(scope="session")
def binary_builder():
    try:
        yield BinaryBuilder
    finally:
        BinaryBuilder.clean()


@pytest.fixture
def server_factory(tmpdir, free_port, binary_builder):
    yield ServerFactory(tmpdir, free_port, binary_builder)
