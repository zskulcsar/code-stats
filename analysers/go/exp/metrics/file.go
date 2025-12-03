package metrics

import (
	"fmt"
	"go/ast"
)

// Simple file based metrics
type FileMetric struct {
	fileName      string
	fileABCMetric ABCMetric
	abcMetrics    []ABCMetric
	fileHalstead  HalsteadMetric
	cycloCMetric  []CyclomaticComplexityMetric
	// Basic file metrics
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
	fm.fileHalstead.Init()
	fm.cycloCMetric = make([]CyclomaticComplexityMetric, 0)
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

func (fm *FileMetric) FileHalstead() (fileHalstead HalsteadMetric) {
	return fm.fileHalstead
}

func (fm *FileMetric) FileCyclomaticComplexity() (cycloComplexity []CyclomaticComplexityMetric) {
	return fm.cycloCMetric
}

func (fm *FileMetric) GenerateMetrics(tree *ast.File) (err error) {
	// Basic code metrics: imports, functions, structures
	fmt.Printf("--- %s\n", fm.fileName)
	ast.Inspect(tree, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.ImportSpec:
			fm.nrOfImports++
		case *ast.FuncDecl:
			fm.nrOfFunctionDeclarations++
			// We generate the ABC metric on the function level
			fm.GenerateABCMetrics(n)
			// Calculate the Cyclomatic Complexity on the function level
			fm.GenerateCyclomaticComplexity(n)
		case *ast.StructType:
			fm.nrOfStructs++
		}
		return true
	})
	// Calculate the Halstead metric on the file
	ast.Inspect(tree, func(n ast.Node) bool {
		fm.GenerateHalsteadMetrics(n)
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

func (fm *FileMetric) GenerateHalsteadMetrics(node ast.Node) {
	ast.Walk(&fm.fileHalstead, node)
}

func (fm *FileMetric) GenerateCyclomaticComplexity(node ast.Node) {
	var ccm = CyclomaticComplexityMetric{}
	ast.Walk(&ccm, node)
	fm.cycloCMetric = append(fm.cycloCMetric, ccm)
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
	return fmt.Sprintf("File,\"%s\",%d,%d,%d,%d",
		fm.FileName(), fm.nrOfImports, fm.nrOfFunctionDeclarations, fm.nrOfLines.Go.Code, fm.nrOfStructs)
}
