# lff command

## install

```go get github.com/intelfike/lff```

## Usage

```
  Usage

 lff (-cd [directory path]|-d|-f|-n|-s) [Directory regexp] [File regexp] [Line regexp]


  Examples

 lff => Display files from current directory.
 lff . => Display files from directory recursive.
 lff "" \.go$ => Search files from only current direcotry.
 lff . \.go$ => Recursive search files from all directory.

 lff . \.go$ "func\smain" => Search "func main".
 lff . \.go$ "func main" => Line contains both of "func" and "main".
 lff . \.go$ "func \!main" => Line contains "func". But never contains "main".


  Flags

  -cd string
    	change directory (default ".")
  -d	directory
  -f	full path
  -h	display help
  -n	line number
  -s	display file with stop
```

## License
MIT