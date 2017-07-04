package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/intelfike/jsonbase"
	"github.com/intelfike/lff/fileexp"
	"github.com/intelfike/lff/regexps"
	"github.com/intelfike/wtof"
	"github.com/shiena/ansicolor"
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
	okjson   = flag.Bool("json", false, "display json")
	indent   = flag.String("indent", "", "json indent")
	okline   bool
	jb       = jsonbase.New()
)

func init() {
	flag.Parse()

	// カレントディレクトリ変更
	err := os.Chdir(*cd)
	if err != nil {
		fmt.Println("-cd [path] path is not found.")
	}

	// コマンドライン引数の正規表現を入力
	direlist := spaceReg.Split(flag.Arg(0), -1)
	filelist := spaceReg.Split(flag.Arg(1), -1)
	linelist := spaceReg.Split(flag.Arg(2), -1)
	dire = regexps.New(distrComp(direlist))
	file = regexps.New(distrComp(filelist))
	line = regexps.New(distrComp(linelist))
	okline = !line.IsEmpty()

	*sf = *sf || *op

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

 lff [Options] [Directory regexp] [File regexp] [Line regexp]

 Options = (-cd "directory path"|-d|-f|-n|-s|-o|-json|-indent "indent")

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

// \!から始まる文字を否定、そうでなければ肯定として受ける
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
	if runtime.GOOS == "windows" {
		// Buffer size = 100KB
		f := wtof.New(ansicolor.NewAnsiColorWriter(os.Stdout), 100000)
		defer f.Close()
		os.Stdout = f.File
	}
	defer fmt.Print("\x1b[m")

	// 表示と探索の同期
	ch := make(chan string, 1024)

	// ファイル探索
	go run(ch)

	// 表示用
	if *okjson {
		jb.Indent = *indent
		for filename := range ch {
			d, f := filepath.Split(filename)
			d = strings.Trim(d, "/\\")
			if len(d) == 0 {
				d = "."
			}
			jb.ChildPath(d).Push().Value(f)
		}
		fmt.Print(jb)
	} else {
		for filename := range ch {
			filetext := <-ch
			if okline && len(filetext) == 0 {
				continue
			}
			d, f := filepath.Split(filename)
			if !okline {
				fmt.Println(d + file.OKHightLight(f))
				continue
			}
			fmt.Print("[" + d + file.OKHightLight(f) + "]")
			if *sf {
				s := ""
				fmt.Scanln(&s)
				switch s {
				case "e", "exit":
					os.Exit(1)
				case "a", "all":
					*sf = false
				case "s", "skip":
					filetext = ""
				}
			} else {
				fmt.Println()
			}
			fmt.Print(filetext)

			// 表示時に開ける-oフラグ
			if *op && *sf {
				fmt.Print("Open?(y/)")
				yn := ""
				fmt.Scanln(&yn)
				if yn == "y" {
					err := open.Run(filename)
					if err != nil {
						fmt.Println(err)
					}
				}
			}
			fmt.Println()
		}
	}
}

func run(ch chan string) {
	defer close(ch)
	var dir chan fileexp.FileDir
	var err error
	if dire.IsEmpty() {
		dir, err = fileexp.ReadDir(".", 1024)
	} else {
		dir, err = fileexp.ReadDirAll(".", 1024)
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
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
			fp = fd.Rel(".")
		}

		if !okline {
			ch <- fp
			if !*okjson {
				ch <- ""
			}
		} else {
			err := readFile(fd.Path(), fp, ch)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}
}

// ファイルを読み取って表示する
func readFile(name, fp string, ch chan string) error {
	r, err := os.Open(name)
	if err != nil {
		return err
	}
	defer r.Close()
	br := bufio.NewReader(r)
	filetext := ""
	count := 0
	for {
		count++
		lineStr, err := readLine(br, 100)
		if err != nil {
			break
		}
		if len(lineStr) == 0 {
			continue
		}
		if *okjson {
			if *nf {
				jb.ChildPath(fp).Push().Printf(`{"Num":%d,"Text":"%s"}`, count, lineStr)
			} else {
				jb.ChildPath(fp).Push().Value(lineStr)
			}
		} else {
			if *nf {
				filetext += strconv.Itoa(count) + " "
			}
			filetext += line.OKHightLight(lineStr) + "\n"
		}
	}
	if !*okjson {
		ch <- fp
		ch <- filetext
	}
	return nil
}

// チェック済み行を返す
func readLine(br *bufio.Reader, linemax int) (string, error) {
	b, _, err := br.ReadLine()
	if err != nil {
		return "", err
	}
	lineStr := string(b)
	if strings.ContainsAny(lineStr, "\x00\x01\x02\x03\x04\x05\x06\x07\x08") {
		return "", errors.New("This is Binnary File.")
	}
	// 正規表現に一致するか
	if !line.MatchAll(lineStr) {
		return "", nil
	}
	// 文字数制限
	if len(lineStr) > linemax {
		lineStr = lineStr[:linemax]
	}
	return lineStr, nil
}
