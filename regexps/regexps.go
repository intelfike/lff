package regexps

import (
	"regexp"
	"strings"
)

type RegSet struct {
	OK []*regexp.Regexp
	NG []*regexp.Regexp
}

func AllCompile(ss []string) []*regexp.Regexp {
	regs := []*regexp.Regexp{}
	for _, v := range ss {
		if v == "" {
			continue
		}
		c, err := regexp.Compile(v)
		if err != nil {
			continue
		}
		regs = append(regs, c)
	}
	return regs
}

// func AppendOK(){
// 	r := ss

// 	AllCompile()
// }
// func AppendNG(){

// }
func (rs *RegSet) MatchAll(str string) bool {
	for _, v := range rs.NG {
		if v.MatchString(str) {
			return false
		}
	}
	for _, v := range rs.OK {
		if !v.MatchString(str) {
			return false
		}
	}
	return true
}
func (rs *RegSet) IsEmpty() bool {
	return len(rs.OK) == 0 && len(rs.NG) == 0
}
func (rs *RegSet) OKHightLight(text string) string {
	ss := make([]string, len(rs.OK))
	for _, reg := range rs.OK {
		ss = append(ss, reg.String())

	}
	s := strings.Join(ss, "|")
	reg := regexp.MustCompile("(" + s + ")")
	str := reg.ReplaceAllString(v.Str, "\x1b[31m$1\x1b[m")
	return reg.ReplaceAllString(text, "[$1]")
}
