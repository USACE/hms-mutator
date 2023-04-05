package hms

import (
	"strconv"
	"strings"
)

type Realizations struct {
	Query map[string]int
}

func ReadCsv(csvBytes []byte) (Realizations, error) {
	return Realizations{}, nil

	csvstring := string(csvBytes)
	lines := strings.Split(csvstring, "\r\n") //maybe rn?
	data := make(map[string]int)
	for idx, l := range lines {
		if idx != 0 {
			values := strings.Split(l, ", ")
			reals, err := strconv.Atoi(values[1])
			if err != nil {
				return Realizations{}, err
			}
			data[values[0]] = reals
		}
	}
	return Realizations{Query: data}, nil
}
