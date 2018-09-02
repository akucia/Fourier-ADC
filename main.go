package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"math"
	"math/cmplx"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/olekukonko/tablewriter"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func makeXYPoints(x []float64, y []float64) plotter.XYs {
	pts := make(plotter.XYs, len(x))
	for i := range pts {
		pts[i].X = x[i]
		pts[i].Y = y[i]
	}
	return pts
}

func plotXY(x []float64, y []float64) *plot.Plot {

	p, err := plot.New()
	check(err)

	err = plotutil.AddLinePoints(p, makeXYPoints(x, y))
	check(err)

	return p

}

func plotDFT(data []complex128, fb float64, db bool, file string) {
	amplitudes := make([]float64, len(data))
	frequencies := make([]float64, len(data))

	for i, value := range data {
		v := cmplx.Abs(value)
		if db {
			v = 20 * math.Log10(math.Max(v, 1e-3))
		}
		amplitudes[i] = v
		frequencies[i] = fb * float64(i)
	}

	p := plotXY(frequencies, amplitudes)

	if db {
		p.Y.Label.Text = "Magnitude^2 [db]"
	} else {
		p.Y.Label.Text = "Magnitude^2"
	}

	p.X.Label.Text = "Frequency [Hz]"

	if err := p.Save(10*vg.Inch, 5*vg.Inch, file); err != nil {
		panic(err)
	}
	fmt.Printf("Plot save in %s\n", file)
}

type DFTResult struct {
	k  int
	xk complex128
}

func dftValue(c chan DFTResult, data []float64, k int, N int) {
	defer wg.Done()
	xk := 0i
	for n, x := range data {
		xValue := complex(0, -2.0*math.Pi*float64(k*n)/float64(N))
		xk += complex(float64(x), 0) * cmplx.Exp(xValue)
	}
	c <- DFTResult{k, xk}
}

func DFT(data []float64, N int) []complex128 {
	queue := make(chan DFTResult, N/2)
	dft := make([]complex128, N/2)
	for k := 0; k < (N/2)-1; k++ {
		wg.Add(1)
		go dftValue(queue, data, k, N)
	}
	wg.Wait()

	close(queue)
	for result := range queue {
		dft[result.k] = result.xk
	}
	return dft
}

func loadData(path string) []float64 {
	fmt.Printf("Reading data from %s \n", path)
	file, err := os.Open(path)
	check(err)
	csvReader := csv.NewReader(file)
	lines, err := csvReader.ReadAll()
	check(err)

	values := make([]float64, len(lines))
	for i, value := range lines {
		values[i], err = strconv.ParseFloat(value[0], 64)
		check(err)
	}
	fmt.Printf("Read %d points\n", len(values))
	return values
}

// Calculates the indices of the first n harmonic frequencies, taking into account the
// aliasing effects.
func aliasedHarmonics(
	signalFreq float64,
	baseFreq float64,
	samplingFreq float64,
	nHarmonics int,
) []int {
	mSig := int(signalFreq / baseFreq)
	fmt.Printf("Signal index: %d\n", mSig)
	N := int(samplingFreq / baseFreq)
	fmt.Printf("DFT len : %d\n", N)

	mh := make([]int, nHarmonics)
	for i := 0; i < nHarmonics; i++ {
		k := i + 1
		m := (k * mSig) % N
		if m < N/2 {
			mh[i] = m

		} else {
			mh[i] = N - m
		}
	}
	fmt.Printf("Aliased harmonics indices %v\n", mh)
	aliasedHarmonicFrequencies := make([]float64, len(mh))
	for i, value := range mh {
		aliasedHarmonicFrequencies[i] = baseFreq * float64(value)
	}

	fmt.Printf("Aliased harmonics freqs %.3g\n", aliasedHarmonicFrequencies)
	return mh

}

func intInSlice(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// THD (Total Harmonic Distorsion). Defined as the ratio of the sum of the powers
// of harmonic components to the power of the signal frequency.
func THD(dft []complex128, signalIndices []int, harmonicIndices []int) float64 {
	dftMagnitudes := convertToMagnitudeSquared(dft)
	harmonicsMagnitude := 0.0
	for idx, i := range harmonicIndices {
		if idx > 0 { // skip first value (base freq term)
			harmonicsMagnitude += dftMagnitudes[i]
		}
	}

	signalMagnitude := 0.0
	for _, idx := range signalIndices {
		signalMagnitude += dftMagnitudes[idx]
	}
	return 20.0 * math.Log10(math.Sqrt(harmonicsMagnitude/signalMagnitude))

}

// SNHR (Signal to Non Harmonic Ratio). Defined as the ratio of power of the signal
// frequency to the sum of the powers of non-harmonic components
func SNHR(dft []complex128, signalIndices []int, harmonicIndices []int) float64 {
	dftMagnitudes := convertToMagnitudeSquared(dft)
	nonHarmonicsMagnitude := 0.0
	for idx, value := range dftMagnitudes {
		if !intInSlice(idx, harmonicIndices) && idx > 0 { // skip harmonics and 0 freq term
			nonHarmonicsMagnitude += value
		}
	}

	signalMagnitude := 0.0
	for _, idx := range signalIndices {
		signalMagnitude += dftMagnitudes[idx]
	}

	return 20.0 * math.Log10(math.Sqrt(signalMagnitude/nonHarmonicsMagnitude))

}

// SFDR (Spurious Free Dynamic Range). Defined as the ratio of power of the signal frequency
// to the power of the of the next largest noise or harmonic distortion.
func SFDR(dft []complex128, signalIndices []int) float64 {
	dftMagnitudes := convertToMagnitudeSquared(dft)
	nonSignalMaxMagnitude := 0.0
	for idx, value := range dftMagnitudes {
		if !intInSlice(idx, signalIndices) && idx > 0 { // skip harmonics
			if value > nonSignalMaxMagnitude {
				nonSignalMaxMagnitude = value
			}
		}
	}

	signalMagnitude := 0.0
	for _, idx := range signalIndices {
		signalMagnitude += dftMagnitudes[idx]
	}

	return 20.0 * math.Log10(math.Sqrt(signalMagnitude/nonSignalMaxMagnitude))

}

// SINAD (Signal to Noise And Distortion). Defined as the ratio of the power of the signal
// to the total power of the noise and harmonic distortions.
func SINAD(dft []complex128, signalIndices []int) float64 {
	dftMagnitudes := convertToMagnitudeSquared(dft)
	nonSignalMagnitude := 0.0
	for idx, value := range dftMagnitudes {
		if !intInSlice(idx, signalIndices) && idx > 0 { // skip harmonics
			nonSignalMagnitude += value
		}
	}

	signalMagnitude := 0.0
	for _, idx := range signalIndices {
		signalMagnitude += dftMagnitudes[idx]
	}

	return 20.0 * math.Log10(math.Sqrt(signalMagnitude/nonSignalMagnitude))

}

// ENOB (Effective Number Of Bits). ENOB specifies the number of bits of an ideal ADC
// that would have the same resolution as the tested, real ADC.
func ENOB(sinad float64) float64 {
	return (sinad - 1.76) / 6.02
}

func convertToMagnitudeSquared(data []complex128) []float64 {
	x := make([]float64, len(data))

	for i, value := range data {
		a := cmplx.Abs(value)
		x[i] = a * a
	}
	return x
}

var wg sync.WaitGroup

var (
	input  = flag.String("input", "", "Input file path.")
	fsig   = flag.Float64("fsig", 0.0, "Original signal frequency.")
	fsam   = flag.Float64("fsam", 0.0, "Sampling frequency.")
	dftlen = flag.Int("dftlen", 1024, "Length of the DFT.")
)

func main() {
	fmt.Println("Starting")
	flag.Parse()

	data := loadData(*input)
	fs := *fsam
	freq := *fsig
	fb := fs / float64(*dftlen)

	table := tablewriter.NewWriter(os.Stdout)

	tableData := [][]string{
		{
			strconv.FormatFloat(freq, 'f', 3, 64),
			strconv.FormatFloat(fs, 'f', 3, 64),
			strconv.FormatInt(int64(*dftlen), 10),
			strconv.FormatFloat(fb, 'f', 3, 64),
		},
	}

	table.SetHeader([]string{"Fsig [Hz]", "Fs [Hz]", "DFT len", "Fb [Hz]"})

	for _, v := range tableData {
		table.Append(v)
	}
	table.Render()

	started := time.Now()
	dft := DFT(data, *dftlen)
	elapsed := time.Since(started)
	fmt.Printf("Calculated DFT in %s.\n", elapsed)

	plotFileName := strings.Split(*input, ".")[0] + ".png"
	plotDFT(dft, fb, true, plotFileName)

	mSig := int(freq / fb)
	aliasedHarmonicIndices := aliasedHarmonics(freq, fb, fs, 10)
	signalIndices := []int{mSig}
	thd := THD(dft, signalIndices, aliasedHarmonicIndices)
	snhr := SNHR(dft, signalIndices, aliasedHarmonicIndices)
	sfdr := SFDR(dft, signalIndices)
	sinad := SINAD(dft, signalIndices)
	enob := ENOB(sinad)

	table = tablewriter.NewWriter(os.Stdout)

	tableData = [][]string{
		{
			strconv.FormatFloat(thd, 'f', 3, 64),
			strconv.FormatFloat(snhr, 'f', 3, 64),
			strconv.FormatFloat(sfdr, 'f', 3, 64),
			strconv.FormatFloat(sinad, 'f', 3, 64),
			strconv.FormatFloat(enob, 'f', 3, 64),
		},
	}

	table.SetHeader([]string{"THD [db]", "SNHR [dB]", "SFDR [dB]", "SINAD [dB]", "Enob [bits]"})

	for _, v := range tableData {
		table.Append(v)
	}
	table.Render()
}
