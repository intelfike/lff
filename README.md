# lff command

## install

```go get github.com/intelfike/lff```

## Usage

```
This command is searching dire/file/line tool.


  Usage

 lff (-cd [directory path]|-d|-f|-n|-s) [Directory regexp] [File regexp] [Line regexp]


  Examples

 lff "" \.go$ => Search only current direcotry.
 lff . \.go$ => Recursive search directory.

 lff . \.go$ "func\smain" => Search "func main".
 lff . \.go$ "func main" => Line contains both "func" and "main".
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