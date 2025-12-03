package metrics

import (
	"fmt"
	"go/ast"
	"go/token"
)

// Cyclomatic Complaxity, see https://en.wikipedia.org/wiki/Cyclomatic_complexity
type CyclomaticComplexityMetric struct {
	signature string // Function / method signature
	ccm       int    // The Cylomatic complexity
}

func (ccm *CyclomaticComplexityMetric) Visit(node ast.Node) (w ast.Visitor) {
	switch t := node.(type) {
	case *ast.FuncDecl:
		ccm.signature = GetFuncSignature(t)
	case *ast.IfStmt, *ast.ForStmt, *ast.CaseClause, *ast.CommClause:
		ccm.ccm++
	case *ast.BinaryExpr:
		if t.Op == token.LAND || t.Op == token.LOR {
			ccm.ccm++
		}
	}
	return ccm
}

func (ccm *CyclomaticComplexityMetric) String() string {
	return fmt.Sprintf("CYC,\"%s\",%d", ccm.signature, ccm.ccm)
}
