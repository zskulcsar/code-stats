package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"

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
		fmt.Fprintf(os.Stdout, "*** walking directory %s\n", *dirname)
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
		fmt.Fprintf(os.Stdout, "*** parsing file %s\n", *filename)
		fm, err := parse(*filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse %q: %v\n", *filename, err)
			os.Exit(1)
		}
		fileMetrics = append(fileMetrics, fm)
	}

	// Stats:
	fmt.Printf("\n*** Printing stats.\n\n")
	for _, fm := range fileMetrics {
		fmt.Printf("%s\n", fm.String())
		var fileABCMetric = fm.FileABCMetric()
		fmt.Printf("%s\n", fileABCMetric.String())
		var fileHalsteadMetric = fm.FileHalstead()
		fmt.Printf("%s\n", fileHalsteadMetric.String())
		for _, abcm := range fm.ABCMetrics() {
			fmt.Printf("\t%s\n", abcm.String())
		}
	}
}

func walkDir(root string, fileMetrics *[]metrics.FileMetric) error {
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		//fmt.Printf("---WalkDir('%s', '%s',  '%s')\n", path, d, err)
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".go" {
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
