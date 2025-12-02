package metrics

import (
	"encoding/json"
	"os/exec"
)

type FileClocStat struct {
	Header struct {
		ClocURL        string  `json:"cloc_url"`
		ClocVersion    string  `json:"cloc_version"`
		ElapsedSeconds float64 `json:"elapsed_seconds"`
		NFiles         int     `json:"n_files"`
		NLines         int     `json:"n_lines"`
		FilesPerSecond float64 `json:"files_per_second"`
		LinesPerSecond float64 `json:"lines_per_second"`
	} `json:"header"`
	Go struct {
		NFiles  int `json:"nFiles"`
		Blank   int `json:"blank"`
		Comment int `json:"comment"`
		Code    int `json:"code"`
	} `json:"Go"`
	Sum struct {
		Blank   int `json:"blank"`
		Comment int `json:"comment"`
		Code    int `json:"code"`
		NFiles  int `json:"nFiles"`
	} `json:"SUM"`
}

func fileCLOC(filename string) (fileCloc FileClocStat, err error) {
	fileCloc = FileClocStat{}
	var cmd = exec.Command("cloc", filename, "--json")
	output, err := cmd.Output()
	if err != nil {
		return
	}
	err = json.Unmarshal(output, &fileCloc)
	return
}
