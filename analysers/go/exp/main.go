package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/zkulcsar/metrics/exp/metrics"
)

func main() {
	filename := flag.String("f", "", "Go source file to parse")
	dirname := flag.String("d", "", "Directory containing Go files to parse")
	flag.Parse()

	fileMetrics := make([]metrics.FileMetric, 0)
	switch {
	// At least one has to be specified
	case *filename == "" && *dirname == "":
		flag.Usage()
		os.Exit(1)
	// Both can't be specified
	case *filename != "" && *dirname != "":
		fmt.Fprintln(os.Stderr, "please specify only one of -f or -d")
		os.Exit(1)
	// Check directory & walk
	case *dirname != "":
		info, err := os.Stat(*dirname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "directory %q not found: %v\n", *dirname, err)
			os.Exit(1)
		}
		if !info.IsDir() {
			fmt.Fprintf(os.Stderr, "%q is not a directory\n", *dirname)
			os.Exit(1)
		}
		//fmt.Fprintf(os.Stdout, "*** walking directory %s\n", *dirname)
		if err := walkDir(*dirname, &fileMetrics); err != nil {
			fmt.Fprintf(os.Stdout, "walk directory %q: %v\n", *dirname, err)
			os.Exit(1)
		}
	// Check the file
	case *filename != "":
		if _, err := os.Stat(*filename); err != nil {
			fmt.Fprintf(os.Stderr, "file %q not found: %v\n", *filename, err)
			os.Exit(1)
		}
		//fmt.Fprintf(os.Stdout, "*** parsing file %s\n", *filename)
		fm, err := parse(*filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse %q: %v\n", *filename, err)
			os.Exit(1)
		}
		fileMetrics = append(fileMetrics, fm)
	}

	// Stats:
	// fmt.Printf("\n*** Printing stats.\n\n")
	// for _, fm := range fileMetrics {
	// 	fmt.Printf("%s\n", fm.String())
	// 	var fileABCMetric = fm.FileABCMetric()
	// 	fmt.Printf("%s\n", fileABCMetric.String())
	// 	var fileHalsteadMetric = fm.FileHalstead()
	// 	fmt.Printf("%s\n", fileHalsteadMetric.String())
	// 	// Kinda ugly, but for testing it's fine
	// 	ccms := fm.FileCyclomaticComplexity()
	// 	for i, abcm := range fm.ABCMetrics() {
	// 		fmt.Printf("\t%s\n", abcm.String())
	// 		fmt.Printf("\t%s\n", ccms[i].String())
	// 	}
	// }

	var sm = metrics.SummaryMetrics{}
	sm.CalculateMetrics(fileMetrics)
	tmpl, err := template.ParseFiles("exp/templates/summary.md.tmpl")
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse template: %v\n", err)
		os.Exit(1)
	}
	data := struct {
		Project            string
		CompositeScore     float64
		TotalNrOfFiles     int
		TotalCodeLOC       int
		TotalCommentLOC    int
		NrOfDImports       int
		NrOfStructs        int
		NrOfFunctions      int
		NrOfComplexFuncs   int
		FunPerFMedian      float64
		StrucPerFMedian    float64
		LocPerFMedian      float64
		CommentDensity     float64
		CyclDestinyPerkLOC float64
		CyclCAverage       float64
		CyclCMedian        float64
		CyclCP95           float64
		CyclCHighRate      float64
		CyclCConcentration float64
		HalVolumePerkLOC   float64
		HalEffortPerkLOC   float64
		HalDifMedian       float64
		ABCCodeSizePerFun  float64
		ABCBranCondRatio   float64
		ABCHighRate        float64
	}{
		Project:            filepath.Base(*dirname),
		CompositeScore:     sm.CompositeScore(),
		TotalNrOfFiles:     sm.TotalNrOfFiles(),
		TotalCodeLOC:       sm.TotalCodeLOC(),
		TotalCommentLOC:    sm.TotalCommentLOC(),
		NrOfDImports:       sm.NrOfDImports(),
		NrOfStructs:        sm.NrOfStructs(),
		NrOfFunctions:      sm.NrOfFunctions(),
		NrOfComplexFuncs:   sm.NrOfComplexFuncs(),
		FunPerFMedian:      sm.FunPerFMedian(),
		StrucPerFMedian:    sm.StrucPerFMedian(),
		LocPerFMedian:      sm.LocPerFMedian(),
		CommentDensity:     sm.CommentDensity(),
		CyclDestinyPerkLOC: sm.CyclDestinyPerkLOC(),
		CyclCAverage:       sm.CyclCAverage(),
		CyclCMedian:        sm.CyclCMedian(),
		CyclCP95:           sm.CyclCP95(),
		CyclCHighRate:      sm.CyclCHighRate(),
		CyclCConcentration: sm.CyclCConcentration(),
		HalVolumePerkLOC:   sm.HalVolumePerkLOC(),
		HalEffortPerkLOC:   sm.HalEffortPerkLOC(),
		HalDifMedian:       sm.HalDifMedian(),
		ABCCodeSizePerFun:  sm.ABCCodeSizePerFun(),
		ABCBranCondRatio:   sm.ABCBranCondRatio(),
		ABCHighRate:        sm.ABCHighRate(),
	}
	if err := tmpl.Execute(os.Stdout, data); err != nil {
		fmt.Fprintf(os.Stderr, "execute template: %v\n", err)
	}
}

func walkDir(root string, fileMetrics *[]metrics.FileMetric) error {
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".go" {
			// We can't measure but go source code only
			return nil
		}
		if strings.Contains(path, "_test.go") {
			// Skipping over "test" files
			return nil
		}
		fm, err := parse(path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		*fileMetrics = append(*fileMetrics, fm)
		return nil
	})
	return nil
}

func parse(filename string) (fm metrics.FileMetric, err error) {
	fset := token.NewFileSet()
	tree, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
	if err != nil {
		return
	}
	//ast.Print(fset, tree)

	fm = metrics.NewFileMetric(filename)
	fm.GenerateMetrics(tree)
	return
}
