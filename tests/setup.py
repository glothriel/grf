#!/usr/bin/env python
from setuptools import find_packages, setup

setup(
    name="tests",
    version="0.0.0",
    description="Integration tests for gin-rest-framework",
    author="Konstanty Karagiorgis @glothriel",
    author_email="...",
    packages=find_packages(),
    python_requires=">=3.8.0",
    install_requires=[
        "pytest==7.3.1",
        "retry==0.9.2",
        "requests==2.30.0",
    ],
)
