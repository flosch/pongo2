## How to format time.Time

Display time with format

```golang
package test

import (
	"github.com/flosch/pongo2/v4"
	"testing"
	"time"
)


func TestDate(t *testing.T) {
	// here is the template
	tpl, err := pongo2.FromString("Time now： {{ dateNow|date: \"2006-01-02 15:04:05\" }}")
	if err != nil {
		panic(err)
	}
	//  eval the time.Time for now
	out, err := tpl.Execute(pongo2.Context{"dateNow": time.Now()})
	if err != nil {
		panic(err)
	}
	t.Logf("pOut: %+v", out)
}

```

you will see

```
=== RUN   TestDate
    pongo2gin_test.go:31: pOut: Time now： 2022-03-15 09:53:48
--- PASS: TestDate (0.00s)
PASS

```
