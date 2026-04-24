package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"golang.org/x/term"
	"os"
	"path/filepath"
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
	// flags
	hf        = flag.Bool("h", false, "display help")
	ff        = flag.Bool("f", false, "full path")
	Ff        = flag.Bool("F", false, "plain text line search (no comma-split)")
	df        = flag.Bool("d", false, "directory")
	nf        = flag.Bool("n", false, "line number")
	sf        = flag.Bool("s", false, "display file with stop")
	op        = flag.Bool("o", false, "ask to open a file. (y/[Enter])")
	ef        = flag.Bool("e", false, "hiding errors")
	to        = flag.String("to", "", "replacement string for display (does not modify files)")
	wf        = flag.Bool("w", false, "overwrite files with replacement result (requires -to or -remove-line)")
	removeLine = flag.Bool("remove-line", false, "delete matched lines from file (requires -w)")
	limit    = flag.Int("limit", 100, "line size limit")
	cd       = flag.String("cd", ".", "change directory")
	okjson   = flag.Bool("json", false, "printing json")
	indent   = flag.String("indent", "", "json indent")
	nameOnly = flag.Bool("name-only", false, "If not displaing lines.")

	// derived from flags and arguments in init()
	dire   *regexps.RegSet
	file   *regexps.RegSet
	line   *regexps.RegSet
	lineF  []string
	okline bool
	toSet  bool
	jb     = jsonbase.New()
)

func init() {
	flag.Parse()
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "to" {
			toSet = true
		}
	})

	// カレントディレクトリ変更
	err := os.Chdir(*cd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "-cd [path] path is not found.")
		os.Exit(1)
	}

	direlist := strings.Split(flag.Arg(0), ",")
	filelist := strings.Split(flag.Arg(1), ",")
	lineArg := flag.Arg(2)
	var linelist []string
	if *Ff {
		if lineArg != "" {
			lineF = []string{lineArg}
		}
		linelist = lineF
		line = regexps.New(nil, nil)
	} else {
		linelist = strings.Split(lineArg, ",")
		line = regexps.New(distrComp(linelist))
	}
	if term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Println(direlist, filelist, linelist)
	}
	dire = regexps.New(distrComp(direlist))
	file = regexps.New(distrComp(filelist))
	okline = hasLineFilter()

	if *sf {
		fmt.Print(`[commands]
all(a) -> display all.(disable stop)
skip(s) -> skip file.
exit(e) -> end.
`)
	}
	if *hf {
		fmt.Print(`This command is searching dire/file/line tool.


  Usage

 lff [Options] [Directory regexp] [File regexp] [Line regexp]

  Examples

## Display files from current directory.
	lff
## Display files from directory recursive.
	lff .
## Search files from only current direcotry.
	lff "" \.go$
## Recursive search files from all directory.
	lff . \.go$

## Search "func main".
	lff . \.go$ func\smain
## Line contains both of "func" and "main".
	lff . \.go$ func,main
## Line contains "func". But never contains "main".
	lff . \.go$ func,-main
## Plain text search (no regex, no comma-split).
	lff -F . \.go$ "func main"
## Replace matched text for display only.
	lff . \.go$ "func\s+main" -to="func entry"
## Overwrite file with replacement result.
	lff -F . \.txt$ old-text -to=new-text -w


  Options
`)
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *wf && !toSet && !*removeLine {
		fmt.Fprintln(os.Stderr, "error: -w requires -to or -remove-line")
		os.Exit(1)
	}
	if (toSet || *wf) && !okline {
		fmt.Fprintln(os.Stderr, "error: -to and -w require a line search pattern (3rd argument)")
		os.Exit(1)
	}
	if *removeLine && !*wf {
		fmt.Fprintln(os.Stderr, "error: -remove-line requires -w")
		os.Exit(1)
	}
	if *removeLine && !*Ff {
		// パターン "." は全行にマッチするため危険
		for _, reg := range line.OK {
			if reg.String() == "." {
				fmt.Fprintln(os.Stderr, "error: -remove-line does not allow pattern \".\" (matches all lines)")
				os.Exit(1)
			}
		}
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
		if strings.HasPrefix(v, "-") {
			ng = append(ng, v[1:])
		} else {
			if strings.HasPrefix(v, "\\-") {
				v = "-" + v[2:]
			}
			ok = append(ok, v)
		}
	}
	return ok, ng
}

func hasLineFilter() bool {
	if *Ff {
		return len(lineF) > 0
	}
	return !line.IsEmpty()
}

func matchLine(lineStr string) bool {
	if *Ff {
		for _, target := range lineF {
			if !strings.Contains(lineStr, target) {
				return false
			}
		}
		return true
	}
	return line.MatchAll(lineStr)
}

// getReplaceTargets returns the search patterns used for -to replacement.
func getReplaceTargets() []string {
	if *Ff {
		return append([]string{}, lineF...)
	}
	targets := make([]string, 0, len(line.OK))
	for _, reg := range line.OK {
		targets = append(targets, reg.String())
	}
	return targets
}

// displayLine returns the line formatted for display:
// if -to is set the matched text is replaced and the replacement is highlighted,
// otherwise the matched text is highlighted.
func displayLine(lineStr string) string {
	if toSet {
		result, _ := regexps.ReplaceAll(lineStr, getReplaceTargets(), *to, *Ff)
		if *to != "" {
			result = strings.ReplaceAll(result, *to, "\x1b[31m"+*to+"\x1b[m")
			result = strings.Replace(result, "\x1b[m\x1b[31m", "", -1)
		}
		return result
	}
	if *Ff {
		rv := lineStr
		for _, target := range lineF {
			if target == "" {
				continue
			}
			rv = strings.ReplaceAll(rv, target, "\x1b[31m"+target+"\x1b[m")
		}
		return strings.Replace(rv, "\x1b[m\x1b[31m", "", -1)
	}
	return line.OKHightLight(lineStr)
}

// buildReplacedContent reads name, applies replacement to all lines that match,
// and returns the display text (matched lines only), the full file text for writing,
// the number of matched lines, and any error.
func buildReplacedContent(name string) (string, string, int, error) {
	content, err := os.ReadFile(name)
	if err != nil {
		return "", "", 0, err
	}
	if strings.ContainsAny(string(content), "\x00\x01\x02\x03\x04\x05\x06\x07\x08") {
		return "", "", 0, errors.New("This is Binnary File.")
	}
	targets := getReplaceTargets()
	lines := strings.Split(string(content), "\n")
	var display strings.Builder
	var full strings.Builder
	matchedCount := 0
	for i, lineStr := range lines {
		replaced, _ := regexps.ReplaceAll(lineStr, targets, *to, *Ff)
		full.WriteString(replaced)
		if i < len(lines)-1 {
			full.WriteByte('\n')
		}
		if matchLine(lineStr) {
			matchedCount++
			if *nf {
				display.WriteString(strconv.Itoa(i+1) + " ")
			}
			display.WriteString(replaced + "\n")
		}
	}
	return display.String(), full.String(), matchedCount, nil
}

// buildRemovedContent reads name and returns the file content with matched lines removed,
// a display string of the removed lines, the count of removed lines, and any error.
func buildRemovedContent(name string) (string, string, int, error) {
	content, err := os.ReadFile(name)
	if err != nil {
		return "", "", 0, err
	}
	if strings.ContainsAny(string(content), "\x00\x01\x02\x03\x04\x05\x06\x07\x08") {
		return "", "", 0, errors.New("This is Binnary File.")
	}
	lines := strings.Split(string(content), "\n")
	var display strings.Builder
	var full strings.Builder
	removedCount := 0
	for i, lineStr := range lines {
		if matchLine(lineStr) {
			removedCount++
			if *nf {
				display.WriteString(strconv.Itoa(i+1) + " ")
			}
			display.WriteString(lineStr + "\n")
			continue
		}
		full.WriteString(lineStr)
		if i < len(lines)-1 {
			full.WriteByte('\n')
		}
	}
	return display.String(), full.String(), removedCount, nil
}

func main() {
	if runtime.GOOS == "windows" {
		// Buffer size = 100KB
		f := wtof.New(ansicolor.NewAnsiColorWriter(os.Stdout), 100*(1<<10))
		defer f.Close()
		os.Stdout = f.File
	}
	if term.IsTerminal(int(os.Stdout.Fd())) {
		defer fmt.Print("\x1b[m")
	}

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
		fileCount := 0
		lineCount := 0
		for filename := range ch {
			filetext := <-ch
			if okline && len(filetext) == 0 {
				continue
			}
			fileCount++
			d, f := filepath.Split(filename)
			if !okline || *nameOnly {
				if term.IsTerminal(int(os.Stdout.Fd())) {
					fmt.Println(d + file.OKHightLight(f))
				} else {
					fmt.Println(filename)
				}
				lineCount += strings.Count(filetext, "\n")
				openGenFile(filename)
				continue
			}
			if term.IsTerminal(int(os.Stdout.Fd())) {
				fmt.Print(">>> ", d+file.OKHightLight(f), " >>>")
				if *sf {
					// fmt.Print(" >")
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
				if filetext != "" {
					fmt.Print(filetext)
					fmt.Println("[EOF]")
					lineCount += strings.Count(filetext, "\n")
				}
				openGenFile(filename)
				fmt.Println()
			} else {
				// パイプライン
				ss := strings.Split(filetext, "\n")
				for _, s := range ss {
					fmt.Println(filename + ":" + s)
				}
			}
		}
		if term.IsTerminal(int(os.Stdout.Fd())) {
			fmt.Println()
			if *df {
				fmt.Println("", fileCount, "Directories")
			} else {
				fmt.Println("", fileCount, "Files")
			}
			if okline {
				fmt.Println("", lineCount, "Lines")
			}
		}
	}
}

func openGenFile(filename string) {
	// 表示時に開ける-oフラグ
	if !*op {
		return
	}
	fmt.Print("Open?(y/[Enter])")
	yn := ""
	fmt.Scanln(&yn)
	if yn == "y" {
		err := open.Run(filename)
		fmt.Println(filename)
		if err != nil && *ef {
			fmt.Fprintln(os.Stderr, err)
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
	if err != nil && *ef {
		fmt.Fprintln(os.Stderr, err)
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
		} else if *wf {
			if *removeLine {
				displayText, fullText, removedCount, err := buildRemovedContent(fd.Path())
				if err != nil {
					if *ef {
						fmt.Fprintln(os.Stderr, err)
					}
					continue
				}
				if removedCount == 0 {
					continue
				}
				if err := os.WriteFile(fd.Path(), []byte(fullText), fd.Info.Mode().Perm()); err != nil {
					if *ef {
						fmt.Fprintln(os.Stderr, err)
					}
					continue
				}
				_ = displayText
				fmt.Printf("removed %d lines from %s\n", removedCount, fp)
			} else {
				displayText, fullText, matchedCount, err := buildReplacedContent(fd.Path())
				if err != nil {
					if *ef {
						fmt.Fprintln(os.Stderr, err)
					}
					continue
				}
				if displayText == "" {
					continue
				}
				if err := os.WriteFile(fd.Path(), []byte(fullText), fd.Info.Mode().Perm()); err != nil {
					if *ef {
						fmt.Fprintln(os.Stderr, err)
					}
					continue
				}
				fmt.Printf("replaced %d lines from %s\n", matchedCount, fp)
			}
		} else {
			filetext, err := readFile(fd.Path(), fp)
			if err != nil && *ef {
				fmt.Fprintln(os.Stderr, err)
			}
			if !*okjson {
				ch <- fp
				ch <- filetext
			}
		}
	}
}

// ファイルを読み取って表示する
// name=読み取るファイル
// fp=表示するファイル名
func readFile(name, fp string) (string, error) {
	r, err := os.Open(name)
	if err != nil && *ef {
		return "", err
	}
	defer r.Close()
	br := bufio.NewReader(r)
	filetext := ""
	count := 0
	for {
		count++
		lineStr, err := readLine(br, *limit)
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
			filetext += displayLine(lineStr) + "\n"
		}
	}

	return filetext, nil
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
	if !matchLine(lineStr) {
		return "", nil
	}
	// 文字数制限 (0以下なら無制限)
	if linemax > 0 && len(lineStr) > linemax {
		lineStr = lineStr[:linemax]
	}
	return lineStr, nil
}
