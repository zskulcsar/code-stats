package metrics

import (
	"go/ast"
	"strings"
)

// Gets the signature of a function as string
func GetFuncSignature(f *ast.FuncDecl) string {
	var sb = strings.Builder{}
	sb.WriteString(f.Name.String())
	sb.WriteString("(")

	params := f.Type.Params.List
	for i, v := range params {
		if v.Names != nil {
			sb.WriteString(v.Names[0].Name)
			if i < len(params)-1 {
				sb.WriteString(", ")
			}
		}
	}
	sb.WriteString(")")
	return sb.String()
}
