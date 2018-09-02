package main

import (
	"math"
	"testing"

	log "github.com/sirupsen/logrus"
)

const testDataFile = "./data/example_data.csv"
const samplingFreq = 4000
const dftLen = 1024
const signalFreq = 402.34375

func TestDFT(t *testing.T) {
	// test setup
	eps := 1e-6
	data, err := loadData(testDataFile)
	if err != nil {
		log.Fatalf("could not load data from the file %s: err", *input, err)
	}
	dft := DFT(data, dftLen)
	freq := signalFreq
	baseFreq := samplingFreq / float64(dftLen)
	mSig := int(freq / baseFreq)
	signalIndices := []int{mSig}

	expectedAliasedHarmonicIndices := []int{103, 206, 309, 412, 509, 406, 303, 200, 97, 6}
	aliasedHarmonicIndices := aliasedHarmonics(signalFreq, baseFreq, samplingFreq, 10)
	for i := range expectedAliasedHarmonicIndices {
		if aliasedHarmonicIndices[i] != expectedAliasedHarmonicIndices[i] {
			t.Fatalf(
				"harmonic index %d should be: %d, got: %d",
				i,
				expectedAliasedHarmonicIndices[i],
				aliasedHarmonicIndices[i],
			)
		}
	}

	expectedThd := -72.79221360760943
	thd := THD(dft, signalIndices, aliasedHarmonicIndices)
	if math.Abs(thd-expectedThd) > eps {
		t.Fatalf("THD value should be: %f, got: %f", expectedThd, thd)
	}

	expectedSnhr := 50.01075896705737
	snhr := SNHR(dft, signalIndices, aliasedHarmonicIndices)
	if math.Abs(snhr-expectedSnhr) > eps {
		t.Fatalf("SNHR value should be: %f, got: %f", expectedSnhr, snhr)
	}

	expectedSfdr := 65.76798708208678
	sfdr := SFDR(dft, signalIndices)
	if math.Abs(sfdr-expectedSfdr) > eps {
		t.Fatalf("SFDR value should be: %f, got: %f", expectedSfdr, sfdr)
	}

	expectedSinad := 49.9879294422944
	sinad := SINAD(dft, signalIndices)
	if math.Abs(sinad-expectedSinad) > eps {
		t.Fatalf("SINAD value should be: %f, got: %f", expectedSinad, sinad)
	}

	expectedEnob := 8.011283960514021
	enob := ENOB(sinad)
	if math.Abs(enob-expectedEnob) > eps {
		t.Fatalf("ENOB value should be: %f, got: %f", expectedEnob, enob)
	}

}
