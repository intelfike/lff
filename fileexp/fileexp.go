package fileexp

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type FileDir struct {
	Dir    string
	Info   os.FileInfo
	Reader *os.File
}

func ReadDirAll(dir string, bufsize int) (chan FileDir, error) {
	_, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	ch := make(chan FileDir, bufsize)
	go recReadDir(dir, ch, 0)
	return ch, nil
}

func recReadDir(dir string, ch chan FileDir, depth int) {
	Infos, _ := ioutil.ReadDir(dir)
	for _, v := range Infos {
		ch <- FileDir{Dir: dir, Info: v}
		if v.IsDir() {
			recReadDir(filepath.Join(dir, v.Name()), ch, depth+1)
		}
	}
	if depth == 0 {
		close(ch)
	}
}

func ReadDir(dir string, bufsize int) (chan FileDir, error) {
	d, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	ch := make(chan FileDir, bufsize)
	go func() {
		for _, v := range d {
			ch <- FileDir{Dir: dir, Info: v}
		}
		close(ch)
	}()
	return ch, nil
}

func (fd *FileDir) Abs() string {
	f, _ := filepath.Abs(fd.Path())
	return f
}
func (fd *FileDir) Rel(curd string) string {
	f, _ := filepath.Rel(curd, fd.Path())
	return f
}
func (fd *FileDir) Path() string {
	return filepath.Join(fd.Dir, fd.Info.Name())
}

type Line struct {
	Str string
	Num int
}

func (fd *FileDir) Open() {
	r, _ := os.Open(fd.Path())
	fd.Reader = r
}
func (fd *FileDir) Close() {
	fd.Reader.Close()
}
func (fd *FileDir) ReadChan(bufsize, linelength int) chan Line {
	ch := make(chan Line, bufsize)
	go func() {
		r := bufio.NewReader(fd.Reader)
		count := 0
		for {
			b, _, err := r.ReadLine()
			str := string(b)
			if err != nil {
				break
			}
			if strings.ContainsAny(str, "\x00\x01\x02\x03\x04\x05\x06\x07\x08") {
				break
			}
			if len(str) > linelength {
				str = str[:linelength]
			}
			count++
			ch <- Line{Str: str, Num: count}
		}
		close(ch)
	}()
	return ch
}
