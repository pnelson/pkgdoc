# pkgdoc

Package pkgdoc prepares package documentation from source.

## Usage

```go
package main

import (
    "log"
    "github.com/pnelson/pkgdoc"
)

func main() {
    doc, err := pkgdoc.New("net")
    if err != nil {
        log.Fatal(err)
    }
    log.Println(doc.Synopsis)
}
```
