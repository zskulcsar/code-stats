package metrics

import (
	"fmt"
	"go/ast"
	"math"
	"strings"
)

// Halstead metrics, see https://en.wikipedia.org/wiki/Halstead_complexity_measures
type HalsteadMetric struct {
	fn1 float64 // the number of distinct operators
	fn2 float64 // the number of distinct operands
	fN1 float64 // the total number of operators
	fN2 float64 // the total number of operands

	operators map[string]int
	operands  map[string]int
}

func (hm *HalsteadMetric) Init() {
	// Less than ideal, somewhat bad object design here
	hm.operators = map[string]int{}
	hm.operands = map[string]int{}
}

func (hm *HalsteadMetric) Vocabulary() float64 {
	hm.calculate()
	return hm.fn1 + hm.fn2
}

func (hm *HalsteadMetric) Length() float64 {
	hm.calculate()
	return hm.fN1 + hm.fN2
}

func (hm *HalsteadMetric) EstimatedLength() float64 {
	hm.calculate()
	return hm.fn1*math.Log2(hm.fn1) + hm.fn2*math.Log2(hm.fn2)
}

func (hm *HalsteadMetric) Volume() float64 {
	hm.calculate()
	return hm.Length() * math.Log2(hm.Vocabulary())
}

func (hm *HalsteadMetric) Difficulty() float64 {
	hm.calculate()
	return hm.fn1 / 2 * hm.fN2 / hm.fn2
}

func (hm *HalsteadMetric) Effort() float64 {
	hm.calculate()
	return hm.Difficulty() * hm.Volume()
}

func (hm *HalsteadMetric) calculate() {
	hm.fn1 = float64(len(hm.operators))
	hm.fn2 = float64(len(hm.operands))

	for _, v := range hm.operators {
		hm.fN1 += float64(v)
	}
	for _, v := range hm.operands {
		hm.fN2 += float64(v)
	}
}

func (hm *HalsteadMetric) String() string {
	var sb = strings.Builder{}
	sb.WriteString("\"Halstead\",")
	sb.WriteString(fmt.Sprintf("%.2f", hm.Vocabulary()))
	sb.WriteString(",")
	sb.WriteString(fmt.Sprintf("%.2f", hm.Length()))
	sb.WriteString(",")
	sb.WriteString(fmt.Sprintf("%.2f", hm.EstimatedLength()))
	sb.WriteString(",")
	sb.WriteString(fmt.Sprintf("%.2f", hm.Volume()))
	sb.WriteString(",")
	sb.WriteString(fmt.Sprintf("%.2f", hm.Difficulty()))
	sb.WriteString(",")
	sb.WriteString(fmt.Sprintf("%.2f", hm.Effort()))
	sb.WriteString(
		fmt.Sprintf(" :: n1 = %.2f, n2 = %.2f, N1 = %.2f, N2 = %.2f",
			hm.fn1, hm.fn2, hm.fN1, hm.fN2))
	return sb.String()
}

func (hm *HalsteadMetric) Visit(node ast.Node) (w ast.Visitor) {
	if node == nil {
		return nil
	}
	hm.walkDecl(node)
	return hm
}

// Steal, see https://github.com/shoooooman/go-complexity-analysis/blob/master/complexity.go
func (hm *HalsteadMetric) walkDecl(n ast.Node) {
	switch n := n.(type) {
	case *ast.GenDecl:
		hm.appendValidSymb(n.Lparen.IsValid(), n.Rparen.IsValid(), "()")

		if n.Tok.IsOperator() {
			hm.operators[n.Tok.String()]++
		} else {
			hm.operands[n.Tok.String()]++
		}
		for _, s := range n.Specs {
			hm.walkSpec(s)
		}
	case *ast.FuncDecl:
		hm.operators["func"]++
		hm.operators[n.Name.Name]++
		if n.Recv == nil {
			hm.operators["()"]++
		} else {
			hm.operators["()"] += 2
		}
		hm.walkStmt(n.Body)
	}
}

func (hm *HalsteadMetric) walkStmt(n ast.Node) {
	switch n := n.(type) {
	case *ast.DeclStmt:
		hm.walkDecl(n.Decl)
	case *ast.ExprStmt:
		hm.walkExpr(n.X)
	case *ast.SendStmt:
		hm.walkExpr(n.Chan)
		if n.Arrow.IsValid() {
			hm.operators["<-"]++
		}
		hm.walkExpr(n.Value)
	case *ast.IncDecStmt:
		hm.walkExpr(n.X)
		if n.Tok.IsOperator() {
			hm.operators[n.Tok.String()]++
		}
	case *ast.AssignStmt:
		if n.Tok.IsOperator() {
			hm.operators[n.Tok.String()]++
		}
		for _, exp := range n.Lhs {
			hm.walkExpr(exp)
		}
		for _, exp := range n.Rhs {
			hm.walkExpr(exp)
		}
	case *ast.GoStmt:
		if n.Go.IsValid() {
			hm.operators["go"]++
		}
		hm.walkExpr(n.Call)
	case *ast.DeferStmt:
		if n.Defer.IsValid() {
			hm.operators["defer"]++
		}
		hm.walkExpr(n.Call)
	case *ast.ReturnStmt:
		if n.Return.IsValid() {
			hm.operators["return"]++
		}
		for _, e := range n.Results {
			hm.walkExpr(e)
		}
	case *ast.BranchStmt:
		if n.Tok.IsOperator() {
			hm.operators[n.Tok.String()]++
		} else {
			hm.operands[n.Tok.String()]++
		}
		if n.Label != nil {
			hm.walkExpr(n.Label)
		}
	case *ast.BlockStmt:
		hm.appendValidSymb(n.Lbrace.IsValid(), n.Rbrace.IsValid(), "{}")
		for _, s := range n.List {
			hm.walkStmt(s)
		}
	case *ast.IfStmt:
		if n.If.IsValid() {
			hm.operators["if"]++
		}
		if n.Init != nil {
			hm.walkStmt(n.Init)
		}
		hm.walkExpr(n.Cond)
		hm.walkStmt(n.Body)
		if n.Else != nil {
			hm.operators["else"]++
			hm.walkStmt(n.Else)
		}
	case *ast.SwitchStmt:
		if n.Switch.IsValid() {
			hm.operators["switch"]++
		}
		if n.Init != nil {
			hm.walkStmt(n.Init)
		}
		if n.Tag != nil {
			hm.walkExpr(n.Tag)
		}
		hm.walkStmt(n.Body)
	case *ast.SelectStmt:
		if n.Select.IsValid() {
			hm.operators["select"]++
		}
		hm.walkStmt(n.Body)
	case *ast.ForStmt:
		if n.For.IsValid() {
			hm.operators["for"]++
		}
		if n.Init != nil {
			hm.walkStmt(n.Init)
		}
		if n.Cond != nil {
			hm.walkExpr(n.Cond)
		}
		if n.Post != nil {
			hm.walkStmt(n.Post)
		}
		hm.walkStmt(n.Body)
	case *ast.RangeStmt:
		if n.For.IsValid() {
			hm.operators["for"]++
		}
		if n.Key != nil {
			hm.walkExpr(n.Key)
			if n.Tok.IsOperator() {
				hm.operators[n.Tok.String()]++
			} else {
				hm.operands[n.Tok.String()]++
			}
		}
		if n.Value != nil {
			hm.walkExpr(n.Value)
		}
		hm.operators["range"]++
		hm.walkExpr(n.X)
		hm.walkStmt(n.Body)
	case *ast.CaseClause:
		if n.List == nil {
			hm.operators["default"]++
		} else {
			for _, c := range n.List {
				hm.walkExpr(c)
			}
		}
		if n.Colon.IsValid() {
			hm.operators[":"]++
		}
		if n.Body != nil {
			for _, b := range n.Body {
				hm.walkStmt(b)
			}
		}
	}
}

func (hm *HalsteadMetric) walkSpec(spec ast.Spec) {
	switch spec := spec.(type) {
	case *ast.ValueSpec:
		for _, n := range spec.Names {
			hm.walkExpr(n)
			if spec.Type != nil {
				hm.walkExpr(spec.Type)
			}
			if spec.Values != nil {
				for _, v := range spec.Values {
					hm.walkExpr(v)
				}
			}
		}
	}
}

func (hm *HalsteadMetric) walkExpr(exp ast.Expr) {
	switch exp := exp.(type) {
	case *ast.ParenExpr:
		hm.appendValidSymb(exp.Lparen.IsValid(), exp.Rparen.IsValid(), "()")
		hm.walkExpr(exp.X)
	case *ast.SelectorExpr:
		hm.walkExpr(exp.X)
		hm.walkExpr(exp.Sel)
	case *ast.IndexExpr:
		hm.walkExpr(exp.X)
		hm.appendValidSymb(exp.Lbrack.IsValid(), exp.Rbrack.IsValid(), "{}")
		hm.walkExpr(exp.Index)
	case *ast.SliceExpr:
		hm.walkExpr(exp.X)
		hm.appendValidSymb(exp.Lbrack.IsValid(), exp.Rbrack.IsValid(), "[]")
		if exp.Low != nil {
			hm.walkExpr(exp.Low)
		}
		if exp.High != nil {
			hm.walkExpr(exp.High)
		}
		if exp.Max != nil {
			hm.walkExpr(exp.Max)
		}
	case *ast.TypeAssertExpr:
		hm.walkExpr(exp.X)
		hm.appendValidSymb(exp.Lparen.IsValid(), exp.Rparen.IsValid(), "()")
		if exp.Type != nil {
			hm.walkExpr(exp.Type)
		}
	case *ast.CallExpr:
		hm.walkExpr(exp.Fun)
		hm.appendValidSymb(exp.Lparen.IsValid(), exp.Rparen.IsValid(), "()")
		if exp.Ellipsis != 0 {
			hm.operators["..."]++
		}
		for _, a := range exp.Args {
			hm.walkExpr(a)
		}
	case *ast.StarExpr:
		if exp.Star.IsValid() {
			hm.operators["*"]++
		}
		hm.walkExpr(exp.X)
	case *ast.UnaryExpr:
		if exp.Op.IsOperator() {
			hm.operators[exp.Op.String()]++
		} else {
			hm.operands[exp.Op.String()]++
		}
		hm.walkExpr(exp.X)
	case *ast.BinaryExpr:
		hm.walkExpr(exp.X)
		hm.operators[exp.Op.String()]++
		hm.walkExpr(exp.Y)
	case *ast.KeyValueExpr:
		hm.walkExpr(exp.Key)
		if exp.Colon.IsValid() {
			hm.operators[":"]++
		}
		hm.walkExpr(exp.Value)
	case *ast.BasicLit:
		if exp.Kind.IsLiteral() {
			hm.operands[exp.Value]++
		} else {
			hm.operators[exp.Value]++
		}
	case *ast.FuncLit:
		hm.walkExpr(exp.Type)
		hm.walkStmt(exp.Body)
	case *ast.CompositeLit:
		hm.appendValidSymb(exp.Lbrace.IsValid(), exp.Rbrace.IsValid(), "{}")
		if exp.Type != nil {
			hm.walkExpr(exp.Type)
		}
		for _, e := range exp.Elts {
			hm.walkExpr(e)
		}
	case *ast.Ident:
		if exp.Obj == nil {
			hm.operators[exp.Name]++
		} else {
			hm.operands[exp.Name]++
		}
	case *ast.Ellipsis:
		if exp.Ellipsis.IsValid() {
			hm.operators["..."]++
		}
		if exp.Elt != nil {
			hm.walkExpr(exp.Elt)
		}
	case *ast.FuncType:
		if exp.Func.IsValid() {
			hm.operators["func"]++
		}
		hm.appendValidSymb(true, true, "()")
		if exp.Params.List != nil {
			for _, f := range exp.Params.List {
				hm.walkExpr(f.Type)
			}
		}
	case *ast.ChanType:
		if exp.Begin.IsValid() {
			hm.operators["chan"]++
		}
		if exp.Arrow.IsValid() {
			hm.operators["<-"]++
		}
		hm.walkExpr(exp.Value)
	}
}

func (hm *HalsteadMetric) appendValidSymb(lvalid bool, rvalid bool, symb string) {
	if lvalid && rvalid {
		hm.operators[symb]++
	}
}
