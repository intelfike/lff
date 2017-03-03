package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/intelfike/lff/fileexp"
	"github.com/intelfike/lff/regexps"
	"github.com/skratchdot/open-golang/open"
)

var (
	dire     *regexps.RegSet
	file     *regexps.RegSet
	line     *regexps.RegSet
	spaceReg = regexp.MustCompile("\\s+")
	hf       = flag.Bool("h", false, "display help")
	ff       = flag.Bool("f", false, "full path")
	df       = flag.Bool("d", false, "directory")
	nf       = flag.Bool("n", false, "line number")
	sf       = flag.Bool("s", false, "display file with stop")
	op       = flag.Bool("o", false, "open file. (y/[Enter])")
	cd       = flag.String("cd", ".", "change directory")
	arglen   int
	wg       sync.WaitGroup
)

func init() {
	flag.Parse()
	arglen = len(flag.Args())
	direlist := spaceReg.Split(flag.Arg(0), -1)
	filelist := spaceReg.Split(flag.Arg(1), -1)
	linelist := spaceReg.Split(flag.Arg(2), -1)
	dire = regexps.New(distrComp(direlist))
	file = regexps.New(distrComp(filelist))
	line = regexps.New(distrComp(linelist))

	if *sf {
		fmt.Println(
			`all(a) -> display all.(disable stop)
skip(s) -> skip file.
exit(e) -> end.
`)
	}
	if *hf {
		fmt.Println(`This command is searching dire/file/line tool.


  Usage

 lff (-cd [directory path]|-d|-f|-n|-s|-o) [Directory regexp] [File regexp] [Line regexp]


  Examples

 lff => Display files from current directory.
 lff . => Display files from directory recursive.
 lff "" \.go$ => Search files from only current direcotry.
 lff . \.go$ => Recursive search files from all directory.

 lff . \.go$ "func\smain" => Search "func main".
 lff . \.go$ "func main" => Line contains both of "func" and "main".
 lff . \.go$ "func \!main" => Line contains "func". But never contains "main".


  Flags
`)
		flag.PrintDefaults()
		os.Exit(1)
	}
}
func distrComp(s []string) ([]string, []string) {
	ok := []string{}
	ng := []string{}
	for _, v := range s {
		if len(v) == 0 {
			continue
		}
		if strings.HasPrefix(v, "\\!") {
			ng = append(ng, v[2:])
		} else {
			if strings.HasPrefix(v, "\\\\!") {
				v = "\\!" + v[3:]
			}
			ok = append(ok, v)
		}
	}
	return ok, ng
}

func main() {
	ch := make(chan string, 1024)
	if !line.IsEmpty() {
		go lineDisper(ch)
	}
	var dir chan fileexp.FileDir
	var err error
	if dire.IsEmpty() {
		dir, err = fileexp.ReadDir(*cd, 1024)
	} else {
		dir, err = fileexp.ReadDirAll(*cd, 1024)
	}
	if err != nil {
		fmt.Println(err)
		return
	}
	for fd := range dir {
		// disp dir or file.
		if *df {
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

		fp := fd.Abs()
		if !*ff {
			fp = "./" + fd.Rel(*cd)
		}
		directory, filename := filepath.Split(fp)
		disppath := directory + file.OKHightLight(filename)

		if line.IsEmpty() {
			fmt.Println(disppath)
		} else {
			fd.Open()
			filetext := ""
			for v := range fd.ReadChan(1024, 100) {
				if !line.MatchAll(v.Str) {
					continue
				}
				if *nf {
					filetext += strconv.Itoa(v.Num) + " "
				}

				filetext += line.OKHightLight(v.Str) + "\n"
			}
			fd.Close()
			if len(filetext) != 0 {
				wg.Add(1)
				ch <- disppath
				ch <- filetext
			}
		}
	}
	wg.Wait()
}

func lineDisper(ch chan string) {
	for {
		lineDisperLoop(ch)
		wg.Done()
	}
}

func lineDisperLoop(ch chan string) {
	filename := <-ch
	fmt.Print("[", filename, "]")
	filetext := <-ch
	if *sf {
		s := ""
		fmt.Scanln(&s)
		switch s {
		case "e", "exit":
			os.Exit(1)
		case "a", "all":
			*sf = false
		case "s", "skip":
			// wg.Done()
			return
		}
	} else {
		fmt.Println()
	}
	fmt.Println(filetext)
	if *op && *sf {
		fmt.Print("Open file?(y/)")
		yn := ""
		fmt.Scanln(&yn)
		if yn != "y" {
			fmt.Println()
			return
		}
		filename = strings.Replace(filename, "\x1b[31m", "", -1)
		filename = strings.Replace(filename, "\x1b[m", "", -1)

		fmt.Println(open.Run(filename))
	}
}
