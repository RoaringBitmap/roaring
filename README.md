RoaringBitmap [![Build Status](https://travis-ci.org/tgruben/roaring.png)](https://travis-ci.org/tgruben/roaring) 
=============

This is a go port of the Roaring bitmap data structured.  The origin java version can be found at https://github.com/lemire/RoaringBitmap and the supporting paper at

http://arxiv.org/abs/1402.6407

For an alternative implementation in Go, see https://github.com/fzandona/goroar
The two versions were written independently.

This code is licensed under Apache License, Version 2.0 (ASL2.0). 

### Dependencies

  - go get github.com/smartystreets/goconvey/convey
  - go get github.com/willf/bitset

Naturally, you also need to grab the roaring code itself:
  - go get github.com/tgruben/roaring


### Example



```go
package main

import (
    "fmt"
    "github.com/tgruben/roaring"
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
}
```



### Documentation

Current documentation is available at http://godoc.org/github.com/tgruben/roaring

### To do

  - Implement a Remove function
  - Write performance benchmarks
  - Important: accelerate operations with assembly instructions (e.g., POPCNT)
  - Implement fast aggregation techniques (see https://github.com/lemire/RoaringBitmap/blob/master/src/main/java/org/roaringbitmap/FastAggregation.java)
