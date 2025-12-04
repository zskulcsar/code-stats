package metrics

import (
	"fmt"
	"math"
	"math/big"
	"sort"
)

// Cyclomatic Complexity thresholds
const (
	CC_LOW      int     = 10 // Low Cyclomatic Complexity
	CC_MODERATE int     = 20 // Moderate Cyclomatic Complexity
	CC_HIGH     int     = 50 // High Cyclomatic Complexity
	CC_TOP_N    int     = 10 // Top N functions to check for Cyclomatic Complexity concentration
	ABC_T_HIGH  float64 = 15 // Threshold for a high ABC Code Size (suggested)
	KLOC_MAGN   int     = 1  // The magnitude for kLOC
)

// Weights
const (
	W_CC_MEDIAN int = 1 // Weight for median Cyclomatic Complexity
	W_CC_P95    int = 1 // Weight for P95 Cyclomatic Complexity
	W_ABC_FUN   int = 1 // Weight for ABC code size per functions
	W_HAL_EFF   int = 1 // Weight for Halstead effort per kLOC
	W_COM_DEN   int = 1 // Weight for comment density
)

type SummaryMetrics struct {
	// Calculated metrics
	cyclDestinyPerkLOC float64 // (sum of CC over functions with ABC code size > 0) / (total code LOC / 1000)
	cyclCAverage       float64 // average CC over functions with ABC code size > 0
	cyclCMedian        float64 // median CC over functions with ABC code size > 0
	cyclCP95           float64 // 95th percentile CC over functions with ABC code size > 0
	cyclCHighRate      float64 // fraction of functions with CC > threshold (ABC code size > 0)
	cyclCConcentration float64 // (sum of CC in top N functions) / (sum of CC in all functions) with ABC code size > 0
	halVolumePerkLOC   float64 // (sum of Halstead Volume over files) / (total code LOC / 1000)
	halEffortPerkLOC   float64 // (sum of Halstead Effort over files) / (total code LOC / 1000)
	halDifMedian       float64 // median CC over functions with ABC code size > 0 (per instruction)
	abcCodeSizePerFun  float64 // median ABC code size per function (ABC code size > 0)
	abcBranCondRatio   float64 // (sum of ABC code size over functions with ABC code size > 0) / (number of such functions)
	abcHighRate        float64 // fraction of functions with ABC code size above ABC_T_HIGH
	// Simple(ish) metrics
	totalNrOfFiles   int // Total nr of files
	totalCodeLOC     int // Total nr of LoC
	totalCommentLOC  int // Total nr of Comment lines
	nrOfDImports     int // Nr of distinc imports across all files
	nrOfStructs      int // Nr of strcuctures across all files
	nrOfFunctions    int // Nr of functions across all files
	nrOfComplexFuncs int // Nr of functions that are not simple (ie.: ABC > 0)
	// Calculated simple metrics
	funPerFMedian   float64 // median(number of functions over all_files)
	strucPerFMedian float64 // median(number of structs over all_files)
	locPerFMedian   float64 // median((total code LOC) / (number of functions).filtered_on(function.ABCMetric.codeSize > 0))
	commentDensity  float64 // (total comment LOC) / (total code LOC + total comment LOC)
	// Composite
	compositeScore float64 // W_CC_MEDIAN*z(median CC) + W_CC_P95*z(P95 CC) + W_ABC_FUN*z(ABC per function) + W_HAL_EFF*z(Halstead effort per kLOC) – W_COM_DEN*z(comment density)
}

func (sm *SummaryMetrics) CalculateMetrics(fileMetrics []FileMetric) {
	// Reset
	*sm = SummaryMetrics{}
	var kLocMagnitude float64 = float64(KLOC_MAGN * 1000)

	var (
		totalCodeLOC    int
		totalCommentLOC int
		distinctImports map[string]int
		totalStructs    int
		totalFunctions  int
		nrOfComplFuncs  int

		ccValues         []float64
		abcValues        []float64
		halVolumeValues  []float64
		halEffortValues  []float64
		halEffortPerK    []float64
		commentDensities []float64
		funPerFile       []float64
		structsPerFile   []float64
		locPerFunction   []float64
	)
	distinctImports = map[string]int{}

	for _, fm := range fileMetrics {
		for imp := range fm.imports {
			distinctImports[imp]++
		}
		totalStructs += fm.nrOfStructs

		funsInFile := len(fm.abcMetrics)
		totalFunctions += funsInFile
		funPerFile = append(funPerFile, float64(funsInFile))
		structsPerFile = append(structsPerFile, float64(fm.nrOfStructs))

		codeLOC := fm.nrOfLines.Go.Code
		commentLOC := fm.nrOfLines.Go.Comment
		totalCodeLOC += codeLOC
		totalCommentLOC += commentLOC

		funsWithMetrics := 0
		for i := 0; i < minInt(len(fm.abcMetrics), len(fm.cycloCMetric)); i++ {
			abcm := fm.abcMetrics[i]
			if abcm.CodeSize() == 0 {
				continue
			}
			cc := fm.cycloCMetric[i].ccm
			ccValues = append(ccValues, float64(cc))
			abcValues = append(abcValues, float64(abcm.CodeSize()))
			funsWithMetrics++
		}
		if funsWithMetrics > 0 {
			locPerFunction = append(locPerFunction, float64(codeLOC)/float64(funsWithMetrics))
		}
		nrOfComplFuncs += funsWithMetrics

		vol := fm.fileHalstead.Volume()
		halVolumeValues = append(halVolumeValues, vol)
		eff := fm.fileHalstead.Effort()
		halEffortValues = append(halEffortValues, eff)
		if codeLOC > 0 {
			halEffortPerK = append(halEffortPerK, eff/(float64(codeLOC)/kLocMagnitude))
		}

		if codeLOC+commentLOC > 0 {
			cd := float64(commentLOC) / float64(codeLOC+commentLOC)
			commentDensities = append(commentDensities, cd)
		}
	}

	kLOC := float64(totalCodeLOC) / kLocMagnitude
	totalCC := sumFloatBig(ccValues)

	// Cyclomatic complexity metrics
	if kLOC > 0 {
		sm.cyclDestinyPerkLOC = div(totalCC, kLOC)
	}
	sm.cyclCAverage = meanFloat64(ccValues)
	sm.cyclCMedian = medianFloat64(ccValues)
	sm.cyclCP95 = percentileFloat64(ccValues, 95)
	if len(ccValues) > 0 {
		sm.cyclCHighRate = float64(countAbove(ccValues, float64(CC_HIGH))) / float64(len(ccValues))
	}
	sm.cyclCConcentration = ccConcentration(ccValues, CC_TOP_N)

	// Halstead metrics
	if kLOC > 0 {
		sm.halVolumePerkLOC = div(sumFloatBig(halVolumeValues), kLOC)
		fmt.Printf("--- sm.halVolumePerkLOC %f\n", sm.halVolumePerkLOC)
		sm.halEffortPerkLOC = div(sumFloatBig(halEffortValues), kLOC)
		fmt.Printf("--- sm.halEffortPerkLOC %f\n", sm.halEffortPerkLOC)
	} else {
		fmt.Printf("!!! halVolumePerkLOC & halEffortPerkLOC is not calculated !!!")
	}
	// As requested: median CC over functions with ABC code size > 0
	sm.halDifMedian = medianFloat64(ccValues)

	// ABC metrics
	sm.abcCodeSizePerFun = medianFloat64(abcValues)
	if len(abcValues) > 0 {
		sm.abcBranCondRatio = div(sumFloatBig(abcValues), float64(len(abcValues)))
		sm.abcHighRate = float64(countAbove(abcValues, ABC_T_HIGH)) / float64(len(abcValues))
	}

	// Simple metrics
	sm.totalNrOfFiles = len(fileMetrics)
	sm.totalCodeLOC = totalCodeLOC
	sm.totalCommentLOC = totalCommentLOC
	sm.nrOfDImports = len(distinctImports)
	sm.nrOfStructs = totalStructs
	sm.nrOfFunctions = totalFunctions
	sm.nrOfComplexFuncs = nrOfComplFuncs
	sm.funPerFMedian = medianFloat64(funPerFile)
	sm.strucPerFMedian = medianFloat64(structsPerFile)
	sm.locPerFMedian = medianFloat64(locPerFunction)
	if totalCodeLOC+totalCommentLOC > 0 {
		sm.commentDensity = float64(totalCommentLOC) / float64(totalCodeLOC+totalCommentLOC)
	}

	// Composite score (z-scores within this project)
	zMedianCC := zScore(sm.cyclCMedian, ccValues)
	zP95CC := zScore(sm.cyclCP95, ccValues)
	zABC := zScore(sm.abcCodeSizePerFun, abcValues)
	zHalEff := zScore(sm.halEffortPerkLOC, halEffortPerK)
	zComment := zScore(sm.commentDensity, commentDensities)

	// TODO: can return NaN, should use "math/big"
	sm.compositeScore = float64(W_CC_MEDIAN)*zMedianCC +
		float64(W_CC_P95)*zP95CC +
		float64(W_ABC_FUN)*zABC +
		float64(W_HAL_EFF)*zHalEff -
		float64(W_COM_DEN)*zComment
}

func zScore(value float64, values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	mean := meanFloat64(values)
	std := stddevFloat64(values, mean)
	if std == 0 {
		return 0
	}
	return (value - mean) / std
}

func ccConcentration(values []float64, topN int) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]float64(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] > sorted[j] })

	limit := minInt(topN, len(sorted))
	topSum := sumFloatBig(sorted[:limit])
	total := sumFloatBig(sorted)

	var bconc = topSum.Quo(topSum, total)
	var fconc, acc = bconc.Float64()
	if acc == big.Exact {
		return fconc
	} else {
		fmt.Printf("--- CommentConcentration can't be calcuated (yet).")
		return math.NaN()
	}
}

func countAbove(values []float64, threshold float64) int {
	count := 0
	for _, v := range values {
		if v > threshold {
			count++
		}
	}
	return count
}

func sumFloatBig(values []float64) (sum *big.Float) {
	sum = new(big.Float)
	for _, v := range values {
		if math.IsNaN(v) {
			// TODO: apparently volume can be NaN, should bring the big (guns)
			v = 0
		}
		var bv = new(big.Float).SetFloat64(v)
		sum.Add(sum, bv)
	}
	return sum
}

func div(x *big.Float, y float64) float64 {
	var by = new(big.Float).SetFloat64(y)
	var mean, acc = x.Quo(x, by).Float64()

	if acc == big.Exact {
		// The result can be represented as a float64
		return mean
	} else {
		fmt.Printf("--- Value can't be calcuated (yet).")
		return math.NaN()
	}
}

func meanFloat64(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum = sumFloatBig(values)
	return div(sum, float64(len(values)))
}

func stddevFloat64(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		diff := v - mean
		sum += diff * diff
	}
	return math.Sqrt(sum / float64(len(values)))
}

func medianFloat64(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

func percentileFloat64(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	if p <= 0 {
		return minFloat(values)
	}
	if p >= 100 {
		return maxFloat(values)
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	idx := (p / 100) * float64(len(sorted)-1)
	lower := int(idx)
	upper := int(math.Ceil(idx))
	if lower == upper {
		return sorted[lower]
	}
	weight := idx - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

func minFloat(values []float64) float64 {
	min := values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min
}

func maxFloat(values []float64) float64 {
	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (sm *SummaryMetrics) CyclDestinyPerkLOC() float64 {
	return sm.cyclDestinyPerkLOC
}

func (sm *SummaryMetrics) CyclCAverage() float64 {
	return sm.cyclCAverage
}

func (sm *SummaryMetrics) CyclCMedian() float64 {
	return sm.cyclCMedian
}

func (sm *SummaryMetrics) CyclCP95() float64 {
	return sm.cyclCP95
}

func (sm *SummaryMetrics) CyclCHighRate() float64 {
	return sm.cyclCHighRate
}

func (sm *SummaryMetrics) CyclCConcentration() float64 {
	return sm.cyclCConcentration
}

func (sm *SummaryMetrics) HalVolumePerkLOC() float64 {
	return sm.halVolumePerkLOC
}

func (sm *SummaryMetrics) HalEffortPerkLOC() float64 {
	return sm.halEffortPerkLOC
}

func (sm *SummaryMetrics) HalDifMedian() float64 {
	return sm.halDifMedian
}

func (sm *SummaryMetrics) ABCCodeSizePerFun() float64 {
	return sm.abcCodeSizePerFun
}

func (sm *SummaryMetrics) ABCBranCondRatio() float64 {
	return sm.abcBranCondRatio
}

func (sm *SummaryMetrics) ABCHighRate() float64 {
	return sm.abcHighRate
}

func (sm *SummaryMetrics) TotalNrOfFiles() int {
	return sm.totalNrOfFiles
}

func (sm *SummaryMetrics) TotalCodeLOC() int {
	return sm.totalCodeLOC
}
func (sm *SummaryMetrics) TotalCommentLOC() int {
	return sm.totalCommentLOC
}

func (sm *SummaryMetrics) NrOfDImports() int {
	return sm.nrOfDImports
}

func (sm *SummaryMetrics) NrOfStructs() int {
	return sm.nrOfStructs
}

func (sm *SummaryMetrics) NrOfFunctions() int {
	return sm.nrOfFunctions
}

func (sm *SummaryMetrics) NrOfComplexFuncs() int {
	return sm.nrOfComplexFuncs
}

func (sm *SummaryMetrics) FunPerFMedian() float64 {
	return sm.funPerFMedian
}

func (sm *SummaryMetrics) StrucPerFMedian() float64 {
	return sm.strucPerFMedian
}

func (sm *SummaryMetrics) LocPerFMedian() float64 {
	return sm.locPerFMedian
}

func (sm *SummaryMetrics) CommentDensity() float64 {
	return sm.commentDensity
}

func (sm *SummaryMetrics) CompositeScore() float64 {
	return sm.compositeScore
}

func (sm *SummaryMetrics) String() string {
	return fmt.Sprintf(
		"%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%d,%d,%d,%.2f,%.2f,%.2f,%.2f,%.2f",
		sm.cyclDestinyPerkLOC,
		sm.cyclCAverage,
		sm.cyclCMedian,
		sm.cyclCP95,
		sm.cyclCHighRate,
		sm.cyclCConcentration,
		sm.halVolumePerkLOC,
		sm.halEffortPerkLOC,
		sm.halDifMedian,
		sm.abcCodeSizePerFun,
		sm.abcBranCondRatio,
		sm.abcHighRate,
		sm.nrOfDImports,
		sm.nrOfStructs,
		sm.nrOfFunctions,
		sm.funPerFMedian,
		sm.strucPerFMedian,
		sm.locPerFMedian,
		sm.commentDensity,
		sm.compositeScore,
	)
}
