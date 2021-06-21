# conf

[![Go Reference](https://pkg.go.dev/badge/github.com/ucarion/conf.svg)](https://pkg.go.dev/github.com/ucarion/conf)

`conf` is a small utility for configuring Golang programs. It addresses three
things the stdlib `flag` package makes a bit difficult:

1. Defining large numbers of flags is tedious.
2. You can't "parse" flags from env vars, only argv.
3. Printing out your config on startup, with secrets redacted, is also tedious.

With `conf`,

1. You can create flags using a struct with tagged fields.
2. You can parse flag values from both env vars and argv.
3. You can get a copy of your struct from (1) with secrets zeroed out.

Here's a working example:

```go
// examples/hello-world/main.go:
package main

import (
   "fmt"

   "github.com/ucarion/conf"
)

func main() {
   config := struct {
      Username string `conf:"name,noredact" usage:"who to log in as"`
      Password string `conf:"password"`
   }{
      Username: "jdoe",
   }

   conf.Load(&config)
   fmt.Println("raw config", config)
   fmt.Println("redacted config", conf.Redact(config))
}
```

```console
$ ./hello-world -h
Usage of ./hello-world:
  -name string
    	who to log in as (env var HELLO_WORLD_NAME) (default "jdoe")
  -password string
    	(env var HELLO_WORLD_PASSWORD)

$ HELLO_WORLD_PASSWORD=letmein ./hello-world -name alan
raw config {alan letmein}
redacted config {alan }
```

## Installation

You can start using `conf` by running:

```bash
go get github.com/ucarion/conf
```

## Usage

You define your struct-of-flags like this:

```go
config := struct {
	Username string `conf:"username,noredact"`
	Password string `conf:"password"`
}{
    Username: "jdoe", // a default value for --username
}
```

Unexported and non-`conf`-tagged fields are always ignored. You can optionally
add a `usage` field.

You parse flag values like this:

```go
conf.Load(&config)
```

Assuming your program's name is `./hello-world`, then `config.Username` gets
populated from `--username` or `HELLO_WORLD_USERNAME`, and `config.Password`
gets populated from `--password` or `HELLO_WORLD_PASSWORD`.

You can print out your struct-of-flags, with `Password` zeroed out, like this:

```go
fmt.Println(conf.Redact(config))
```

Everything is assumed to be secret unless marked with `noredact`.

## FAQs

### What types are supported?

`conf` supports the same types `flag` supports: `bool`, `int`, `string`,
`time.Duration`, `float64`, `int64` `uint`, `uint64`.

### Can you load structs recursively?

Yes. You need to tag the struct with `conf`:

```go
// examples/substruct/main.go:
package main

import (
   "fmt"
   "time"

   "github.com/ucarion/conf"
)

func main() {
   type dbConfig struct {
      DSN     string        `conf:"dsn"`
      Timeout time.Duration `conf:"timeout,noredact"`
   }

   config := struct {
      PrimaryDB   dbConfig `conf:"primary-db,noredact"`
      SecondaryDB dbConfig `conf:"secondary-db,noredact"`
   }{}

   conf.Load(&config)
   fmt.Println("raw config", config)
   fmt.Println("redacted config", conf.Redact(config))
}
```

```console
$ ./substruct -h
Usage of ./substruct:
  -primary-db-dsn string
    	(env var SUBSTRUCT_PRIMARY_DB_DSN)
  -primary-db-timeout duration
    	(env var SUBSTRUCT_PRIMARY_DB_TIMEOUT)
  -secondary-db-dsn string
    	(env var SUBSTRUCT_SECONDARY_DB_DSN)
  -secondary-db-timeout duration
    	(env var SUBSTRUCT_SECONDARY_DB_TIMEOUT)
```

`conf.Redact` will zero out the entire sub-struct unless you explicitly mark the
relevant field with `noredact`.
