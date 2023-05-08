import logging
import os
import re
from subprocess import PIPE, run

logger = logging.getLogger(__name__)


def run_process(process, **kwargs):
    def prepare_env():
        stripped_env = kwargs.pop("env", os.environ)
        extra_env = kwargs.pop("extra_env", {})
        # concatenate stripped env with extra ones
        return {**stripped_env, **extra_env}

    logger.info(" ".join(process))
    check_returncode = kwargs.pop("check_returncode", True)
    rt = run(
        process,
        stdout=PIPE,
        stderr=PIPE,
        env=prepare_env(),
        **kwargs,
    )
    try:
        if check_returncode:
            rt.check_returncode()
    finally:
        logger.info(rt.stdout.decode())
        logger.info(f"Return code: {rt.returncode}")
        stderr = rt.stderr.decode()
        if stderr:
            logger.error(stderr)
    return rt


class AnyUUID(str):
    REGEX = r"^[0-9a-f]{8}-([0-9a-f]{4}-){3}[0-9a-f]{12}$"

    def __eq__(self, __value: object) -> bool:
        if not re.match(self.REGEX, __value, re.I):
            return False
        return True
