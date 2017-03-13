package regexps

import (
	"regexp"
	"strings"
)

type RegSet struct {
	OK          []*regexp.Regexp
	NG          []*regexp.Regexp
	hightLights *regexp.Regexp
}

func New(ok, ng []string) *RegSet {
	rs := new(RegSet)
	rs.OK = AllCompile(ok)
	rs.NG = AllCompile(ng)
	ss := make([]string, 0, len(rs.OK))
	for _, reg := range rs.OK {
		if reg.String() == "." {
			continue
		}
		ss = append(ss, reg.String())
	}
	s := strings.Join(ss, "|")
	rs.hightLights = regexp.MustCompile("(" + s + ")")
	return rs
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
func (rs *RegSet) IsAcceptAll() bool {
	return rs.hightLights.String() == "()"
}
func (rs *RegSet) OKHightLight(text string) string {
	if rs.hightLights.String() == "()" {
		return text
	}
	rv := rs.hightLights.ReplaceAllString(text, "\x1b[31m$1\x1b[m")
	rv = strings.Replace(rv, "\x1b[m\x1b[31m", "", -1)
	return rv
}
