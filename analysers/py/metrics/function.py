import ast
from logging import getLogger
from typing import Any
from math import sqrt, pow

log = getLogger(__name__)


# Assingments
#   - Occurrence of an assignment
#
# Branches
#   - Occurrence of a function call.
#   - Occurrence of an Async function call
#
# # Conditionals
#   - Occurrence of a conditional operator
#   - Occurrence of the following keywords: 'if', 'else', 'elif', 'except', 'case' as in ast.match_case.
class FunctionMetrics(ast.NodeVisitor):
    def __init__(self, signature: str | None, a=0, b=0, c=0) -> None:
        self.signature = signature
        # ABC Metrics
        self.assingments = a
        self.branches = b
        self.conditionals = c
        # ABC Code Size
        self.__code_size: float = 0
        self.code_size()

    # Assingments
    def visit_Assign(self, node: ast.Assign) -> Any:
        self.assingments += 1
        return super().generic_visit(node)

    # Branches
    def visit_Await(self, node: ast.Await) -> Any:
        self.branches += 1
        return super().generic_visit(node)

    def visit_Call(self, node: ast.Call) -> Any:
        self.branches += 1
        return super().generic_visit(node)

    # Conditionals
    def visit_If(self, node: ast.If) -> Any:
        if node.orelse is not None:
            self.conditionals += 1
        return super().generic_visit(node)

    def visit_Compare(self, node: ast.Compare) -> Any:
        self.conditionals += 1
        return super().generic_visit(node)

    def visit_ExceptHandler(self, node: ast.ExceptHandler) -> Any:
        self.conditionals += 1
        return super().generic_visit(node)

    def visit_For(self, node: ast.For) -> Any:
        self.conditionals += 1
        return super().generic_visit(node)

    def visit_match_case(self, node: ast.match_case) -> Any:
        self.conditionals += 1
        return super().generic_visit(node)

    # Other
    def code_size(self) -> float:
        if self.__code_size == 0:
            self.__code_size = sqrt(
                pow(self.assingments, 2)
                + pow(self.branches, 2)
                + pow(self.conditionals, 2)
            )
        return self.__code_size

    def cyclomatic_complexity(self) -> int:
        # Things is ... conditionals and Cyclomatic Complexity are the same
        return self.conditionals

    def __str__(self) -> str:
        return (
            f"ABC,"
            f"{self.signature},"
            f"{self.code_size()},"
            f"{self.assingments},"
            f"{self.branches},"
            f"{self.conditionals}\n\t"
            f"CYC,{self.signature},{self.conditionals}"
        )
