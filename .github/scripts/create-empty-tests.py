"""
This script creates empty test files for all packages in the repository,
if they don't have one already. Golang ignores from coverage reports packages
that don't have any tests :(

Reference: https://github.com/golang/go/issues/24570
"""
import os

for folder, subs, files in os.walk("./pkg"):
    if "/examples/" not in folder and (not any(file.endswith("_test.go") for file in files)):
        test_file = os.path.join(folder, "empty_test.go")
        with open(test_file, "w") as f:
            f.write("package " + os.path.basename(folder) + "\n")
