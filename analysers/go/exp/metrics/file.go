package metrics

import (
	"go/ast"
	"strconv"
	"strings"
)

// Simple file based metrics
type FileMetric struct {
	fileName      string
	fileABCMetric ABCMetric
	abcMetrics    []ABCMetric
	// Metrics
	nrOfImports              int
	nrOfFunctionDeclarations int
	nrOfLines                FileClocStat
	nrOfStructs              int
}

func NewFileMetric(fileName string) FileMetric {
	fm := FileMetric{}
	fm.abcMetrics = make([]ABCMetric, 0)
	fm.fileName = fileName
	fm.fileABCMetric = ABCMetric{signature: fileName}
	return fm
}

func (fm *FileMetric) FileName() string {
	return fm.fileName
}

func (fm *FileMetric) ABCMetrics() (fileABCMetrics []ABCMetric) {
	return fm.abcMetrics
}

func (fm *FileMetric) FileABCMetric() (fileABCMetric ABCMetric) {
	return fm.fileABCMetric
}

func (fm *FileMetric) GenerateMetrics(tree *ast.File) (err error) {
	// Basic code metrics: imports, functions, structures
	ast.Inspect(tree, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.ImportSpec:
			fm.nrOfImports++
		case *ast.FuncDecl:
			fm.nrOfFunctionDeclarations++
			fm.GenerateABCMetrics(n)
		case *ast.StructType:
			fm.nrOfStructs++
		}
		return true
	})
	// Calculate ABC metrics for the file
	fm.calcABCSum()
	fm.CodeSize()
	// Use 'cloc' to calculate the lines of code metrics
	fm.nrOfLines, err = fileCLOC(fm.fileName)
	if err != nil {
		return
	}
	return nil
}

func (fm *FileMetric) GenerateABCMetrics(node ast.Node) {
	var abcm = ABCMetric{}
	ast.Walk(&abcm, node)
	fm.abcMetrics = append(fm.abcMetrics, abcm)
}

func (fm *FileMetric) CodeSize() (codeSize int) {
	return fm.fileABCMetric.CodeSize()
}

func (fm *FileMetric) calcABCSum() (fileABCM ABCMetric) {
	for _, abcm := range fm.abcMetrics {
		fm.fileABCMetric.AssingmentAdd(abcm.Assingments())
		fm.fileABCMetric.BranchAdd(abcm.Branches())
		fm.fileABCMetric.ConditionAdd(abcm.Conditionals())
	}
	return fm.fileABCMetric
}

func (fm *FileMetric) String() string {
	var sb = strings.Builder{}
	sb.WriteString("File,\"")
	sb.WriteString(fm.fileName)
	sb.WriteString("\",")
	sb.WriteString(strconv.Itoa(fm.nrOfImports))
	sb.WriteString(",")
	sb.WriteString(strconv.Itoa(fm.nrOfFunctionDeclarations))
	sb.WriteString(",")
	sb.WriteString(strconv.Itoa(fm.nrOfLines.Go.Code))
	sb.WriteString(",")
	sb.WriteString(strconv.Itoa(fm.nrOfStructs))
	return sb.String()
}
