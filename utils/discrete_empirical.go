package utils

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/usace/cc-go-sdk"
)

type DiscreteEmpiricalDistribution struct {
	bin_starts             []int
	cumulative_probability []float64
}
type StormTypeSeasonalityDistributionMap map[string]DiscreteEmpiricalDistribution //storm type DiscreteEmpiricalDistribution.
func NewDescreteEmpiricalDistribution(bin_starts []int, cumuatlive_probs []float64) DiscreteEmpiricalDistribution {
	return DiscreteEmpiricalDistribution{bin_starts: bin_starts, cumulative_probability: cumuatlive_probs}
}
func DescreteEmpiricalDistributionFromBytes(data []byte) DiscreteEmpiricalDistribution {
	stringbytes := string(data)
	lines := strings.Split(stringbytes, "\r\n")
	starts := make([]int, 0)
	probs := make([]float64, 0)
	var dist DiscreteEmpiricalDistribution
	for i, line := range lines {
		if i > 0 {
			if len(line) > 0 {
				vals := strings.Split(line, ",")
				binstart, err := strconv.Atoi(vals[0])
				if err != nil {
					return dist
				}
				starts = append(starts, binstart)
				prob, err := strconv.ParseFloat(vals[1], 64) //if incrimental, update this val with time.
				if err != nil {
					return dist
				}
				probs = append(probs, prob)
			}
		}
	}
	return NewDescreteEmpiricalDistribution(starts, probs)
}
func (ded DiscreteEmpiricalDistribution) Sample(probability float64) int {
	if ded.cumulative_probability[0] < probability {
		for i, p := range ded.cumulative_probability {
			if p >= probability {
				return ded.bin_starts[i]
			}
		}
	} else {
		return ded.bin_starts[0]
	}
	return int(ded.bin_starts[len(ded.bin_starts)-1])
}
func ReadStormDistributions(iomanager cc.IOManager, storeKey string, filePaths []string) (StormTypeSeasonalityDistributionMap, error) {
	StormTypeSeasonalityDistributionMap := make(map[string]DiscreteEmpiricalDistribution)
	store, err := iomanager.GetStore(storeKey)
	if err != nil {
		return StormTypeSeasonalityDistributionMap, err
	}
	session, ok := store.Session.(cc.S3DataStore)
	if !ok {
		return StormTypeSeasonalityDistributionMap, fmt.Errorf("%v was not an s3datastore type", storeKey)
	}
	root := store.Parameters.GetStringOrFail("root")
	for _, path := range filePaths {
		pathpart := strings.Replace(path, fmt.Sprintf("%v/", root), "", -1)
		reader, err := session.Get(pathpart, "")
		if err != nil {
			return StormTypeSeasonalityDistributionMap, err
		}
		bytes, err := io.ReadAll(reader)
		if err != nil {
			return StormTypeSeasonalityDistributionMap, err
		}
		dist := DescreteEmpiricalDistributionFromBytes(bytes)
		if err != nil {
			return StormTypeSeasonalityDistributionMap, err
		}
		StormTypeSeasonalityDistributionMap[path] = dist
	}
	return StormTypeSeasonalityDistributionMap, nil
}
