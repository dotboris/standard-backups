package testutils

import (
	"strings"

	"github.com/lithammer/dedent"
)

func Dedent(in string) string {
	res := in
	res = dedent.Dedent(res)
	res = strings.TrimLeft(res, "\n")
	res = strings.TrimRight(res, "\n")
	return res
}

func DedentYaml(in string) string {
	res := in
	res = Dedent(in)
	res = strings.ReplaceAll(res, "\t", "  ")
	return res
}
