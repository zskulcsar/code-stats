import ast
from typing import Any
from collections import Counter
from logging import getLogger
from .cloc import _file_cloc, FileClocStat
from .function import FunctionMetrics

log = getLogger(__name__)


class FileMetrics(ast.NodeVisitor):
    """Simple file based metrics"""

    def __init__(self, filename: str) -> None:
        self.filename: str = filename
        self.__fileabcmetric = None
        # To be fair: we can just iterate over the file and how Python is
        # that is likely better anyways
        self.fun_metrics: list[FunctionMetrics] = []
        self.filehalstead = None
        self.cyclocmetric = None
        # Basic file metrics
        self.nrofimports: int = 0
        self.imports: Counter[str] = Counter()
        self.nroffunctiondeclarations = 0
        self.nrOflines: FileClocStat
        self.classses: Counter[str] = Counter()
        # Formatting
        self.tabs: int = -1

    def generate_metrics(self):
        file = None
        try:
            file = open(self.filename, "r")
            root = ast.parse(
                file.read(), self.filename
            )  # , mode='exec', type_comments=False, feature_version=None
            self.visit(root)
        finally:
            if file is not None:
                file.close()

        # set the metrics
        self.nrofimports = len(self.imports)
        self.nrOflines = _file_cloc(self.filename)
        self.__calc_file_abc()

    # Imports
    def visit_Import(self, node: ast.Import) -> Any:
        self.__count_imports(node)

    def visit_ImportFrom(self, node: ast.ImportFrom) -> Any:
        self.__count_imports(node)

    # ABC Metric per function
    def visit_FunctionDef(self, node: ast.FunctionDef) -> Any:
        self.__abc_metric(node)
        return super().generic_visit(node)

    def visit_AsyncFunctionDef(self, node: ast.AsyncFunctionDef) -> Any:
        self.__abc_metric(node)
        return super().generic_visit(node)

    def visit_Lambda(self, node: ast.Lambda) -> Any:
        self.__abc_metric(node)
        return super().generic_visit(node)

    # AST
    def generic_visit(self, node: ast.AST) -> Any:
        """Inherited from [ast.NodeVisitor]

        :param self: Description
        :param node: Description
        :type node: ast.AST
        :return: Description
        :rtype: Any
        """
        self.tabs += 1
        if isinstance(node, ast.Module):
            log.debug(f"{''.join(['\t'] * self.tabs)}--- generic_visit {node}")
        else:
            log.debug(
                f"{''.join(['\t'] * self.tabs)}--- generic_visit {node} {ast.dump(node)}"
            )
        ret = super().generic_visit(node)
        self.tabs -= 1
        return ret

    # Helpers
    def __calc_file_abc(self):
        a = 0
        b = 0
        c = 0
        for abcm in self.fun_metrics:
            a += abcm.assingments
            b += abcm.branches
            c += abcm.conditionals
        self.__fileabcmetric = FunctionMetrics(self.filename, a, b, c)

    def __count_imports(self, imp: ast.Import | ast.ImportFrom):
        for alias in imp.names:
            log.debug(f"{type(imp)} :: {alias.name}")
            self.imports[alias.name] += 1

    def __nr_of_classes(self) -> int:
        return len(self.classses)

    def __nr_of_functions(self) -> int:
        """The sum of standalone functions + lambda functions + class methods"""
        nr_of_fun = 0
        for v in self.classses.values():
            nr_of_fun += v
        nr_of_fun += self.nroffunctiondeclarations
        return nr_of_fun

    def __abc_metric(self, node: ast.FunctionDef | ast.AsyncFunctionDef | ast.Lambda):
        self.nroffunctiondeclarations += 1
        if isinstance(node, ast.Lambda):
            abc = FunctionMetrics("lambda")
        else:
            abc = FunctionMetrics(node.name)
        self.fun_metrics.append(abc)
        abc.generic_visit(node)

    def __str__(self) -> str:
        return (
            f"File,"
            f"{self.filename},"
            f"{self.nrofimports},"
            f"{self.__nr_of_functions()},"
            f"{self.nrOflines.Python.code},"
            f"{self.__nr_of_classes()},\n\t"
            f"{self.__fileabcmetric},\n\t"
            f"{'\t'.join(str(fm) + '\n' for fm in self.fun_metrics)}"
        )
