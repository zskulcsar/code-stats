import ast
from collections import Counter
from logging import getLogger

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
        self.nroffunctiondeclarations = int
        self.nrOflines = None  #                FileClocStat
        self.nrOfclasses = 0  #              int
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

    def visit_Import(self, node: ast.Import):
        self.count_imports(node)

    def visit_ImportFrom(self, node: ast.ImportFrom):
        self.count_imports(node)

    def count_imports(self, imp: ast.Import | ast.ImportFrom):
        for alias in imp.names:
            log.debug(f"{type(imp)} :: {alias.name}")
            self.imports[alias.name] += 1

    def generic_visit(self, node: ast.AST) -> any:
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

    def print_tabs(self):
        self.tabs += 1

    def __str__(self) -> str:
        return f"""
				File,
				{self.filename},{self.nrofimports},
				{self.nroffunctiondeclarations},
				{self.nrOflines},{self.nrOfclasses}
				"""

    # def visit_Import(self, node: ast.Import) -> ast.Any:
    # 	print(f"--- visit_Import {node}")
    # 	return super().visit_Import(node)

    # def visit_FunctionDef(self, node: ast.FunctionDef) -> ast.Any:
    # 	print(f"--- visit_FunctionDef {node}")
    # 	return super().visit_FunctionDef(node)

    # def visit_AsyncFor(self, node: ast.AsyncFor) -> ast.Any:
    # 	print(f"--- visit_FunctionDef {node}")
    # 	return super().visit_AsyncFor(node)

    # def visit_ClassDef(self, node: ast.ClassDef) -> ast.Any:
    # 	return super().visit_ClassDef(node)
