# -*- coding: utf-8 -*-
import logging
import subprocess
from typing import List, Union

from .config import config

logger = logging.getLogger(__name__)


class ShellProcess:
    """Wrapper for shell command."""

    def __init__(
        self,
        command: str = "/bin/bash",
        strip_newlines: bool = False,
        return_err_output: bool = True,
    ):
        """Initialize with stripping newlines."""
        self.strip_newlines = strip_newlines
        self.return_err_output = return_err_output
        self.command = command

    def run(self, args: Union[str, List[str]], input=None) -> str:
        """Run the command."""
        if isinstance(args, str):
            args = [args]
        commands = ";".join(args)
        if not commands.startswith(self.command):
            commands = f"{self.command} {commands}"

        return self.exec(commands, input=input)

    def exec(self, commands: Union[str, List[str]], input=None) -> str:
        """Run commands and return final output."""
        if isinstance(commands, str):
            commands = [commands]
        commands = ";".join(commands)
        logger.debug(f"Executing command: {commands}")
        try:
            output = subprocess.run(
                commands,
                shell=True,
                check=True,
                input=input,
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
                timeout=config.timeout,
            ).stdout.decode()
        except subprocess.TimeoutExpired:
            return f"Command execution timed out after {config.timeout} seconds"
        except subprocess.CalledProcessError as error:
            if self.return_err_output:
                return error.stdout.decode()
            return str(error)
        if self.strip_newlines:
            output = output.strip()
        return output
