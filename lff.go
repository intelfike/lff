package main

import (
	"flag"
	"fmt"
	"regexp"

	"github.com/intelfike/lff/fileexp"
	"github.com/intelfike/lff/regexps"
)

var (
	dire     regexps.RegSet
	file     regexps.RegSet
	line     regexps.RegSet
	spaceReg = regexp.MustCompile("\\s+")
	f        = flag.Bool("f", false, "full path")
	d        = flag.Bool("d", false, "directory")
	n        = flag.Bool("n", false, "line number")
	cd       = flag.String("cd", ".", "change directory")
	arglen   int
)

func init() {
	flag.Parse()
	arglen = len(flag.Args())
	direlist := spaceReg.Split(flag.Arg(0), -1)
	filelist := spaceReg.Split(flag.Arg(1), -1)
	linelist := spaceReg.Split(flag.Arg(2), -1)

	// for n, v := range linelist {
	// 	if len(v) == 0 {
	// 		continue
	// 	}
	// 	linelist[n] = "(" + v + ")"
	// }

	dire.OK = regexps.AllCompile(direlist)
	file.OK = regexps.AllCompile(filelist)
	line.OK = regexps.AllCompile(linelist)
}

func main() {
	dir, err := fileexp.ReadDirAll(*cd, 1024)
	if err != nil {
		fmt.Println(err)
		return
	}
	for fd := range dir {
		// disp dir or file.
		if *d {
			if !fd.Info.IsDir() {
				continue
			}
		} else {
			if fd.Info.IsDir() {
				continue
			}
		}

		// regexp switch
		if !dire.MatchAll(fd.Dir) {
			continue
		}
		if !file.MatchAll(fd.Info.Name()) {
			continue
		}
		fmt.Print(*cd + "/")
		if *f {
			fmt.Println(fd.Abs())
		} else {
			fmt.Println(fd.Rel(*cd))
		}

		// line
		if !line.IsEmpty() {
			for v := range fd.ReadChan(1024, 100) {
				if !line.MatchAll(v.Str) {
					continue
				}
				if *n {
					fmt.Print(v.Num, " ")
				}

				fmt.Println(line.OKHightLight(v.Str))
			}
			fmt.Println()
		}
	}
}
