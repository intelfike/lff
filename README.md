# lff command

## install

```go get github.com/intelfike/lff```

## Usage

```
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

  -cd string
      change directory (default ".")
  -d  directory
  -f  full path
  -h  display help
  -indent string
      json indent
  -json
      display json
  -n  line number
  -o  open file. (y/[Enter])
  -s  display file with stop

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