from dataclasses import dataclass
from dataclasses_json import dataclass_json
from subprocess import run

from logging import getLogger

log = getLogger(__name__)


@dataclass_json
@dataclass(frozen=True)
class Header:
    cloc_url: str
    cloc_version: str
    elapsed_seconds: float
    n_files: int
    n_lines: int
    files_per_second: float
    lines_per_second: float


@dataclass_json
@dataclass(frozen=True)
class Python:
    nFiles: int
    blank: int
    comment: int
    code: int


@dataclass_json
@dataclass(frozen=True)
class Sum:
    blank: int
    comment: int
    code: int
    nFiles: int


@dataclass_json
@dataclass(frozen=True)
class FileClocStat:
    header: Header
    Python: Python
    SUM: Sum


def _file_cloc(file_name: str) -> FileClocStat:
    cp = run(["cloc", file_name, "--json"], capture_output=True)
    # TODO: cloc is weird: 'Unable to read:  /home/zsoltk/git/linux-rag-t2/backend/src/adapters/ollama/client.pyk' but returns with '0'
    cp.check_returncode()
    if cp.returncode != 0:
        raise ChildProcessError(cp.stderr)
    return FileClocStat.from_json(cp.stdout)  # type: ignore
