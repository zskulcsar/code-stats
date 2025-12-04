package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"sync"

	"github.com/zkulcsar/metrics/exp/metrics"
)

type parseResult struct {
	fm  metrics.FileMetric
	err error
}

func parseConcurrently(paths []string, workers int) ([]metrics.FileMetric, error) {
	jobs := make(chan string)
	results := make(chan parseResult)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range jobs {
				fm, err := parse(p)
				results <- parseResult{fm: fm, err: err}
			}
		}()
	}

	go func() {
		for _, p := range paths {
			jobs <- p
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	fileMetrics := make([]metrics.FileMetric, 0, len(paths))
	for res := range results {
		if res.err != nil {
			return nil, res.err
		}
		fileMetrics = append(fileMetrics, res.fm)
	}
	return fileMetrics, nil
}

func parse(filename string) (fm metrics.FileMetric, err error) {
	fset := token.NewFileSet()
	fmt.Printf("Parsing file: '%s'\n", filename)
	tree, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
	if err != nil {
		return
	}
	//ast.Print(fset, tree)

	fm = metrics.NewFileMetric(filename)
	fm.GenerateMetrics(tree)
	return
}
