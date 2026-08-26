package utils

import (
	"math/rand"
	"strconv"
)

type StormSampler interface {
	SampleNames(event int64, seeds []SeedSet) error
	SampleName(rng *rand.Rand) string
}

type BootstrapSampler struct {
	StormNames   []string
	yearStormMap map[int][]string
	sampleNames  []string
	minYear      int32
	maxYear      int32
}

func InitBootstrapSampler(Names []string) (*BootstrapSampler, error) {
	//collect names into common years and sample as year groups
	b := BootstrapSampler{}
	b.yearStormMap = make(map[int][]string)
	minYear := 9999
	maxYear := 0
	for _, n := range Names {
		//extract year from name.
		//yyyymmdd_xxhr_storm-type_storm-rank
		yyyy := n[0:4]
		year, err := strconv.Atoi(yyyy)
		if err != nil {
			return &b, err
		}
		storms, ok := b.yearStormMap[year]
		if ok {
			storms = append(storms, n)
			b.yearStormMap[year] = storms
		} else {
			b.yearStormMap[year] = []string{n}
		}
		if year > maxYear {
			maxYear = year
		}
		if year < minYear {
			minYear = year
		}
	}
	b.maxYear = int32(maxYear)
	b.minYear = int32(minYear)
	b.StormNames = Names
	return &b, nil
}
func (b *BootstrapSampler) SampleNames(event int64, seeds []SeedSet) error {

	stormCount := len(b.StormNames)
	sample := make([]string, stormCount)
	rng := rand.New(rand.NewSource(seeds[event].EventSeed)) //check with haden.
	delta := b.maxYear - b.minYear
	i := 0
	for {
		tmpStorms := b.yearStormMap[int(b.minYear+rng.Int31n(delta))]
		for _, s := range tmpStorms {
			sample = append(sample, s)
			i++
			if i >= stormCount {
				break

			}
		}
		if i >= stormCount {
			break

		}
	}
	b.sampleNames = sample
	return nil
}
func (b *BootstrapSampler) SampleName(rng *rand.Rand) string {
	stormCount := len(b.sampleNames)
	return b.sampleNames[rng.Int31n(int32(stormCount))]
}

type JackknifeSampler struct {
	StormNames   []string
	yearStormMap map[int][]string
	sampleNames  []string
	minYear      int32
	maxYear      int32
}

func InitJackknifeSampler(Names []string) (*JackknifeSampler, error) {
	//collect names into common years and sample as year groups
	b := JackknifeSampler{}
	b.yearStormMap = make(map[int][]string)
	minYear := 9999
	maxYear := 0
	for _, n := range Names {
		//extract year from name.
		//yyyymmdd_xxhr_storm-type_storm-rank
		yyyy := n[0:4]
		year, err := strconv.Atoi(yyyy)
		if err != nil {
			return &b, err
		}
		storms, ok := b.yearStormMap[year]
		if ok {
			storms = append(storms, n)
			b.yearStormMap[year] = storms
		} else {
			b.yearStormMap[year] = []string{n}
		}
		if year > maxYear {
			maxYear = year
		}
		if year < minYear {
			minYear = year
		}
	}
	b.maxYear = int32(maxYear)
	b.minYear = int32(minYear)
	b.StormNames = Names
	return &b, nil
}
func (b *JackknifeSampler) SampleNames(event int64, seeds []SeedSet) error {

	stormCount := len(b.StormNames)
	// Initialize with 0 length but pre-allocate capacity for performance
	sample := make([]string, 0, stormCount)
	rng := rand.New(rand.NewSource(seeds[event].EventSeed)) // check with haden.
	delta := b.maxYear - b.minYear - 5
	skipYearMin := b.minYear + rng.Int31n(delta) //need to make this a min skip year and a max skip year...
	skipYearMax := skipYearMin + 5

	for year := b.minYear; year <= b.maxYear; year++ {
		if year >= skipYearMin && year <= skipYearMax {
			continue
		}
		tmpStorms := b.yearStormMap[int(year)]
		for _, s := range tmpStorms {
			sample = append(sample, s)
		}
	}

	b.sampleNames = sample
	return nil
}
func (b *JackknifeSampler) SampleName(rng *rand.Rand) string {
	stormCount := len(b.sampleNames)
	return b.sampleNames[rng.Int31n(int32(stormCount))]
}

type BestEstimateSampler struct {
	StormNames []string
}

func InitBestEstimateSampler(Names []string) (*BestEstimateSampler, error) {
	b := BestEstimateSampler{}
	b.StormNames = Names
	return &b, nil
}
func (b *BestEstimateSampler) SampleNames(event int64, seeds []SeedSet) error {
	return nil
}
func (b *BestEstimateSampler) SampleName(rng *rand.Rand) string {
	stormCount := len(b.StormNames)
	return b.StormNames[rng.Int31n(int32(stormCount))]
}
