package metrics

import (
	"go/ast"
	"go/token"
	"math"
	"strconv"
	"strings"
)

// Assingments
//   - Occurrence of an assignment operator: '=', '*=', '/=', '%=', '+=', '<<=', '>>=', '&=', '^='.
//   - Occurrence of an shor variable decl operator: ':=' (technically this is an assingment)
//   - Occurrence of an increment or a decrement operator: '++', '--'.
//
// Branches
//   - Occurrence of a function call.
//   - Occurrence of a GOTO
//   - Occurrence of a 'go'
//
// # Conditionals
//   - Occurrence of a conditional operator: '<', '>', '<=', '>=', '==', '!='.
//   - Occurrence of the following keywords: 'else', 'case'.
type ABCMetric struct {
	signature string
	// Metrics
	assingments  int
	branches     int
	conditionals int
	// Code size
	codeSize int
}

func (abcm *ABCMetric) Visit(node ast.Node) (w ast.Visitor) {
	switch t := node.(type) {
	// Set the signature
	case *ast.FuncDecl:
		var sb = strings.Builder{}
		sb.WriteString(t.Name.String())
		sb.WriteString("(")

		params := t.Type.Params.List
		for i, v := range params {
			sb.WriteString(v.Names[0].Name)
			if i < len(params)-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteString(")")
		abcm.signature = sb.String()
	// Assingments
	case *ast.AssignStmt:
		abcm.Assingment()
	case *ast.IncDecStmt:
		abcm.Assingment()
	// Branches
	case *ast.CallExpr, *ast.GoStmt:
		abcm.Branch()
	case *ast.BranchStmt:
		if t.Tok == token.GOTO {
			abcm.Branch()
		}
	// Conditionals
	case *ast.CaseClause:
		abcm.Conditional()
	case *ast.IfStmt:
		abcm.Conditional()
		if t.Else != nil {
			abcm.Conditional()
		}
	case *ast.BinaryExpr:
		// This is to check if there is a || or && in an if statement, 'cause ast
		// places BinaryExpr on each side of the BinaryExpr reflected by IfStmt.Cond.
		// Makes sense, but boy the below looks ugly.
		switch t.X.(type) {
		case *ast.BinaryExpr:
			switch t.Y.(type) {
			case *ast.BinaryExpr:
				abcm.Conditional()
			}
		}
	}
	return abcm
}

func (abcm *ABCMetric) Signature() string {
	return abcm.signature
}

func (abcm *ABCMetric) Assingments() int {
	return abcm.assingments
}

func (abcm *ABCMetric) Assingment() {
	abcm.assingments++
}

func (abcm *ABCMetric) AssingmentAdd(a int) {
	abcm.assingments += a
}

func (abcm *ABCMetric) Branches() int {
	return abcm.branches
}

func (abcm *ABCMetric) Branch() {
	abcm.branches++
}

func (abcm *ABCMetric) BranchAdd(b int) {
	abcm.branches += b
}

func (abcm *ABCMetric) Conditional() {
	abcm.conditionals++
}

func (abcm *ABCMetric) ConditionAdd(c int) {
	abcm.conditionals += c
}

func (abcm *ABCMetric) Conditionals() int {
	return abcm.conditionals
}

func (abcm *ABCMetric) CodeSize() int {
	if abcm.codeSize == 0 {
		abcm.codeSize = int(
			math.Sqrt(
				math.Pow(float64(abcm.assingments), 2) +
					math.Pow(float64(abcm.branches), 2) +
					math.Pow(float64(abcm.conditionals), 2)))
	}
	return abcm.codeSize
}

func (abcm *ABCMetric) String() string {
	var sb = strings.Builder{}
	sb.WriteString("ABC,\"")
	sb.WriteString(abcm.signature)
	sb.WriteString("\",")
	sb.WriteString(strconv.Itoa(abcm.CodeSize()))
	sb.WriteString(",")
	sb.WriteString(strconv.Itoa(abcm.assingments))
	sb.WriteString(",")
	sb.WriteString(strconv.Itoa(abcm.branches))
	sb.WriteString(",")
	sb.WriteString(strconv.Itoa(abcm.conditionals))
	return sb.String()
}
