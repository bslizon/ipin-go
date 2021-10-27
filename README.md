# ipin-go
- iphone png normalizer

- transferred from Axel E. Brzostowski iPhone PNG Images Normalizer in Python
  - https://axelbrz.com/?mod=iphone-png-images-normalizer

# How to use
```
$ go get -u -v github.com/bslizon/ipin-go
```

```go
package main

import (
	"os"

	ipingo "github.com/bslizon/ipin-go"
)

func main() {
	b, err := ipingo.GetNormalizedPNG("/tmp/from.png")
	if err != nil {
		panic(err)
	}

	f, err := os.Create("/tmp/to.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err := f.Write(b); err != nil {
		panic(err)
	}
}


```
