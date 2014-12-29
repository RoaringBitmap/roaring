RoaringBitmap [![Build Status](https://travis-ci.org/tgruben/roaring.png)](https://travis-ci.org/tgruben/roaring)[![GoDoc](https://godoc.org/github.com/tgruben/roaring?status.svg)](https://godoc.org/github.com/tgruben/roaring) 
=============

This is a go port of the Roaring bitmap data structure.  The original java version
can be found at https://github.com/lemire/RoaringBitmap and the supporting paper at

http://arxiv.org/abs/1402.6407

The Java and Go version are meant to be binary compatible: you can save bitmaps
from a Java program and load them back in Go, and vice versa.


This code is licensed under Apache License, Version 2.0 (ASL2.0). 


### Dependencies

  - go get github.com/smartystreets/goconvey/convey
  - go get github.com/willf/bitset

Naturally, you also need to grab the roaring code itself:
  - go get github.com/tgruben/roaring


### Example

Here is a simplified but complete example:

```go
package main

import (
    "fmt"
    "github.com/tgruben/roaring"
    "bytes"
)


func main() {
    // example inspired by https://github.com/fzandona/goroa
    fmt.Println("==roaring==")
    rb1 := roaring.BitmapOf(1, 2, 3, 4, 5, 100, 1000)
    fmt.Println(rb1.String())

    rb2 := roaring.BitmapOf(3, 4, 1000)
    fmt.Println(rb2.String())

    rb3 := roaring.NewRoaringBitmap()
    fmt.Println(rb3.String())

    fmt.Println("Cardinality: ", rb1.GetCardinality())

    fmt.Println("Contains 3? ", rb1.Contains(3))

    rb1.And(rb2)

    rb3.Add(1)
    rb3.Add(5)

    rb3.Or(rb1)

    // prints 1, 3, 4, 5, 1000
    i := rb3.Iterator()
    for i.HasNext() {
        fmt.Println(i.Next())
    }
    fmt.Println()

    // next we include an example of serialization
    buf := new(bytes.Buffer)
    rb1.WriteTo(buf) // we omit error handling
    newrb:= roaring.NewRoaringBitmap()
    newrb.ReadFrom(buf)
    if rb1.Equals(newrb) {
    	fmt.Println("I wrote the content to a byte stream and read it back.")
    }
}
```

If you wish to use serialization and handle errors, you might want to 
consider the following sample of code:

```go
	rb := BitmapOf(1, 2, 3, 4, 5, 100, 1000)
	buf := new(bytes.Buffer)
	size,err:=rb.WriteTo(buf)
	if err != nil {
		t.Errorf("Failed writing")
	}
	newrb:= NewRoaringBitmap()
	size,err=newrb.ReadFrom(buf)
	if err != nil {
		t.Errorf("Failed reading")
	}
	if ! rb.Equals(newrb) {
		t.Errorf("Cannot retrieve serialized version")
	}
```




### Documentation

Current documentation is available at http://godoc.org/github.com/tgruben/roaring

### Benchmark

Type

         go test -bench Benchmark -run -

### Alternative

For an alternative implementation in Go, see https://github.com/fzandona/goroar
The two versions were written independently.
