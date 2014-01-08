package main

import (
	"regexp"

	"github.com/russross/blackfriday"
)

var (
	censor   = regexp.MustCompile(`\$\$[^\$]+\$\$|\$[^\$]+\$`)
	uncensor = regexp.MustCompile(`\$+`)
)

func replace(vals [][]byte) func([]byte) []byte {
	i := -1
	return func(b []byte) []byte {
		i++
		return vals[i]
	}
}

func markdown(input []byte) interface{} {
	matches := censor.FindAll(input, -1)
	tex := make([][]byte, len(matches))
	for i, m := range matches {
		tex[i] = make([]byte, len(m))
		for j := range m {
			tex[i][j], m[j] = m[j], '$'
		}
	}
	output := blackfriday.MarkdownCommon(input)
	return uncensor.ReplaceAllFunc(output, replace(tex))
}
