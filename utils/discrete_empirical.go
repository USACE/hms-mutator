package utils

import (
	"strconv"
	"strings"
)

type DiscreteEmpiricalDistribution struct {
	bin_starts             []int
	cumulative_probability []float64
}
func NewDescreteEmpiricalDistribution(bin_starts int[], cumuatlive_probs []float64) DiscreteEmpiricalDistribution{
	return DiscreteEmpiricalDistribution{bin_starts: bin_starts, cumulative_probability: cumuatlive_probs}
}
func DescreteEmpiricalDistributionFromBytes(data []byte) DiscreteEmpiricalDistribution{
	stringbytes := string(data)
	lines := strings.Split(data,"\r\n")
	starts := make([]int)
	probs := make([]probs)
	var dist DiscreteEmpiricalDistribution
	for i, line := range lines{
		if i>0{
			if len(line)>0{
				vals := strings.Split(line, ",")
				binstart, err := strconv.Atoi(vals[0])
				if err!=nil{
					return dist
				}
				prob, err := strconv.ParseFloat(vals[1], 64)//if incrimental, update this val with time.
				if err!=nil{
					return dist
				}
			}
		}
	}
	return NewDescreteEmpiricalDistribution(starts,probs)
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
