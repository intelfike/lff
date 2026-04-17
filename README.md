# lff comm:and

## install

```go get github.com/intelfike/lff```

## Usage

 lff [Options] [Directory regexp] [File regexp] [Line regexp]

###  Examples

```
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
## Replace matched text for display only (file unchanged).
  lff . \.go$ "func\s+main" -to="func entry"
## Overwrite file with replacement result.
  lff -F . \.txt$ old-text -to=new-text -w
```

###  Options

```
  -cd string
      change directory (default ".")
  -d  directory
  -F  plain text line search (no comma-split)
  -f  full path
  -h  display help
  -indent string
      json indent
  -json
      display json
  -n  line number
  -o  open file. (y/[Enter])
  -s  display file with stop
  -to string
      replacement string for display (file unchanged)
  -w  overwrite files with replacement result (requires -to)
```


## Sample

```
~$ lff -cd $GOROOT/src . go$ "func\sDial\("
[./crypto/tls/tls.go]
func Dial(network, addr string, config *Config) (*Conn, error) {

[./log/syslog/syslog.go]
func Dial(network, raddr string, priority Priority, tag string) (*Writer, error) {

[./net/dial.go]
func Dial(network, address string) (Conn, error) {

[./net/rpc/client.go]
func Dial(network, address string) (*Client, error) {

[./net/rpc/jsonrpc/client.go]
func Dial(network, address string) (*rpc.Client, error) {

[./net/smtp/smtp.go]
func Dial(addr string) (*Client, error) {

[./net/textproto/textproto.go]
func Dial(network, addr string) (*Conn, error) {

```

## License
MIT
