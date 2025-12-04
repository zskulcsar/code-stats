package main

import (
	"flag"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/zkulcsar/metrics/exp/metrics"
)

const (
	WORKER_PERCENT float64 = 0.8 // The percentage of the total number of cores to be used
)

func main() {
	dirname := flag.String("d", "", "Directory containing Go files to parse")
	// We default to WORKER_PERCENT (80) percent of the available cores, unless it's explicitly set
	nrOfWorkers := flag.Int("w", int(math.Floor(float64(runtime.NumCPU())*WORKER_PERCENT)), "Nr of workers")
	flag.Parse()

	fileMetrics := make([]metrics.FileMetric, 0)
	paths := make([]string, 0)
	switch {
	// At least one has to be specified
	case *dirname == "":
		flag.Usage()
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
		if err := collectPaths(*dirname, &paths); err != nil {
			fmt.Fprintf(os.Stdout, "walk directory %q: %v\n", *dirname, err)
			os.Exit(1)
		}
	}

	if len(paths) > 0 {
		var err error
		fmt.Printf("Parsing the '%s' folder with %d workers.\n", *dirname, *nrOfWorkers)
		fileMetrics, err = parseConcurrently(paths, *nrOfWorkers)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse files: %v\n", err)
			os.Exit(1)
		}
	}

	var sm = metrics.SummaryMetrics{}
	sm.CalculateMetrics(fileMetrics)
	tmpl, err := template.ParseFiles("exp/templates/summary.md.tmpl")
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse template: %v\n", err)
		os.Exit(1)
	}
	// TODO: do better, SummaryMetrics should handle all of this
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
	// TODO: add the option to generate into a file
	if err := tmpl.Execute(os.Stdout, data); err != nil {
		// TODO: after this all of it should be handled as log rather than \W?[p]rintf()
		fmt.Fprintf(os.Stderr, "execute template: %v\n", err)
	}
}

func collectPaths(root string, paths *[]string) error {
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
		*paths = append(*paths, path)
		return nil
	})
	return nil
}
