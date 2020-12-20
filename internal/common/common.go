package common

import (
	"bufio"
	"os"
	"strings"
)

func ReadLines(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return []string{""}, err
	}
	defer f.Close()

	var ret []string

	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}

		ret = append(ret, strings.Trim(line, "\n"))
	}

	return ret, nil
}
