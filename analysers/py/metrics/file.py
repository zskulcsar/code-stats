import ast
from collections import Counter
from logging import getLogger
from .cloc import _file_cloc

log = getLogger(__name__)


class FileMetrics(ast.NodeVisitor):
    """Simple file based metrics"""

    def __init__(self, filename: str) -> None:
        self.filename = filename
        self.fileabcmetric = None
        self.abcmetrics = None  #   []ABCMetric
        self.filehalstead = None  # HalsteadMetric
        self.cyclocmetric = None  # []CyclomaticComplexityMetric
        # Basic file metrics
        self.nrofimports = 0
        self.imports = Counter()
        self.nroffunctiondeclarations = 0
        self.nrOflines = None  #                FileClocStat
        self.classses = Counter()  #              int
        # Formatting
        self.tabs = -1

    def generate_metrics(self):
        file = None
        try:
            # TODO: open file
            file = open(self.filename, "r")
            root = ast.parse(
                file.read(), filename=self.filename
            )  # , mode='exec', type_comments=False, feature_version=None
            self.visit(root)
        finally:
            # TODO: close the file
            if file != None:
                file.close()

        # set the metrics
        self.nrofimports = len(self.imports)
        self.nrOflines = _file_cloc(self.filename)

    def visit_Import(self, node: ast.Import):
        self._count_imports(node)

    def visit_ImportFrom(self, node: ast.ImportFrom):
        self._count_imports(node)

    # TODO: might be interesting to have class based metrics, like how many functions per class
    def visit_ClassDef(self, node: ast.ClassDef):
        self.classses[node.name] = 0
        for ch in ast.iter_child_nodes(node):
            match type(ch):
                case ast.FunctionDef:
                    self.classses[node.name] += 1
                case ast.Lambda:
                    self.classses[node.name] += 1

    def visit_FunctionDef(self, node: ast.FunctionDef):
        self.nroffunctiondeclarations += 1

    def visit_AsyncFunctionDef(self, node: ast.AsyncFunctionDef):
        self.nroffunctiondeclarations += 1

    def visit_Lambda(self, node: ast.Lambda):
        self.nroffunctiondeclarations

    def _count_imports(self, imp: ast.Import | ast.ImportFrom):
        for alias in imp.names:
            log.debug(f"{type(imp)} :: {alias.name}")
            self.imports[alias.name] += 1

    def _nr_of_classes(self) -> int:
        return len(self.classses)

    def _nr_of_functions(self) -> int:
        """The sum of standalone functions + lambda functions + class methods"""
        nr_of_fun = 0
        for v in self.classses.values():
            nr_of_fun += v
        nr_of_fun += self.nroffunctiondeclarations
        return nr_of_fun

    def generic_visit(self, node: ast.AST):
        """Inherited from [ast.NodeVisitor]

        :param self: Description
        :param node: Description
        :type node: ast.AST
        :return: Description
        :rtype: Any
        """
        self.tabs += 1
        log.debug(f"{''.join(['\t'] * self.tabs)}--- generic_visit {node}")
        ret = super().generic_visit(node)
        self.tabs -= 1
        return ret

    def __str__(self) -> str:
        return (
            f"File,"
            f"{self.filename},"
            f"{self.nrofimports},"
            f"{self._nr_of_functions()},"
            f"{self.nrOflines.Python.code},"
            f"{self._nr_of_classes()}"
        )
