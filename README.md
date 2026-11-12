# roaring 

[![GoDoc](https://godoc.org/github.com/RoaringBitmap/roaring?status.svg)](https://godoc.org/github.com/RoaringBitmap/roaring)

![Go-CI](https://github.com/RoaringBitmap/roaring/workflows/Go-CI/badge.svg)
![Go-ARM-CI](https://github.com/RoaringBitmap/roaring/workflows/Go-ARM-CI/badge.svg)
![Go-Windows-CI](https://github.com/RoaringBitmap/roaring/workflows/Go-Windows-CI/badge.svg)
=============

This is a go version of the Roaring bitmap data structure. 

Roaring bitmaps are found to work well in many important applications:

> Use Roaring for bitmap compression whenever possible. Do not use other bitmap compression methods ([Wang et al., SIGMOD 2017](http://db.ucsd.edu/wp-content/uploads/2017/03/sidm338-wangA.pdf))

This library is used in production in several systems, and it is part of the [Awesome Go collection](https://awesome-go.com). See [Users](#users).

There are also  [Java](https://github.com/RoaringBitmap/RoaringBitmap) and [C/C++](https://github.com/RoaringBitmap/CRoaring) versions.  The Java, C, C++ and Go version are binary compatible: e.g,  you can save bitmaps
from a Java program and load them back in Go, and vice versa. We have a [format specification](https://github.com/RoaringBitmap/RoaringFormatSpec).


This code is licensed under Apache License, Version 2.0 (ASL2.0).

Copyright 2016-... by the authors.

When should you use a bitmap?
===================================


Sets are a fundamental abstraction in
software. They can be implemented in various
ways, as hash sets, as trees, and so forth.
In databases and search engines, sets are often an integral
part of indexes. For example, we may need to maintain a set
of all documents or rows  (represented by numerical identifier)
that satisfy some property. Besides adding or removing
elements from the set, we need fast functions
to compute the intersection, the union, the difference between sets, and so on.


To implement a set
of integers, a particularly appealing strategy is the
bitmap (also called bitset or bit vector). Using n bits,
we can represent any set made of the integers from the range
[0,n): the ith bit is set to one if integer i is present in the set.
Commodity processors use words of W=32 or W=64 bits. By combining many such words, we can
support large values of n. Intersections, unions and differences can then be implemented
 as bitwise AND, OR and ANDNOT operations.
More complicated set functions can also be implemented as bitwise operations.

When the bitset approach is applicable, it can be orders of
magnitude faster than other possible implementation of a set (e.g., as a hash set)
while using several times less memory.

However, a bitset, even a compressed one is not always applicable. For example, if
you have 1000 random-looking integers, then a simple array might be the best representation.
We refer to this case as the "sparse" scenario.

When should you use compressed bitmaps?
===================================

An uncompressed BitSet can use a lot of memory. For example, if you take a BitSet
and set the bit at position 1,000,000 to true and you have just over 100kB. That is over 100kB
to store the position of one bit. This is wasteful  even if you do not care about memory:
suppose that you need to compute the intersection between this BitSet and another one
that has a bit at position 1,000,001 to true, then you need to go through all these zeroes,
whether you like it or not. That can become very wasteful.

This being said, there are definitively cases where attempting to use compressed bitmaps is wasteful.
For example, if you have a small universe size. E.g., your bitmaps represent sets of integers
from [0,n) where n is small (e.g., n=64 or n=128). If you can use uncompressed BitSet and
it does not blow up your memory usage,  then compressed bitmaps are probably not useful
to you. In fact, if you do not need compression, then a BitSet offers remarkable speed.

The sparse scenario is another use case where compressed bitmaps should not be used.
Keep in mind that random-looking data is usually not compressible. E.g., if you have a small set of
32-bit random integers, it is not mathematically possible to use far less than 32 bits per integer,
and attempts at compression can be counterproductive.

How does Roaring compares with the alternatives?
==================================================


Most alternatives to Roaring are part of a larger family of compressed bitmaps that are run-length-encoded
bitmaps. They identify long runs of 1s or 0s and they represent them with a marker word.
If you have a local mix of 1s and 0, you use an uncompressed word.

There are many formats in this family:

* Oracle's BBC is an obsolete format at this point: though it may provide good compression,
it is likely much slower than more recent alternatives due to excessive branching.
* WAH is a patented variation on BBC that provides better performance.
* Concise is a variation on the patented WAH. It some specific instances, it can compress
much better than WAH (up to 2x better), but it is generally slower.
* EWAH is both free of patent, and it is faster than all the above. On the downside, it
does not compress quite as well. It is faster because it allows some form of "skipping"
over uncompressed words. So though none of these formats are great at random access, EWAH
is better than the alternatives.



There is a big problem with these formats however that can hurt you badly in some cases: there is no random access. If you want to check whether a given value is present in the set, you have to start from the beginning and "uncompress" the whole thing. This means that if you want to intersect a big set with a large set, you still have to uncompress the whole big set in the worst case...

Roaring solves this problem. It works in the following manner. It divides the data into chunks of 2<sup>16</sup> integers
(e.g., [0, 2<sup>16</sup>), [2<sup>16</sup>, 2 x 2<sup>16</sup>), ...). Within a chunk, it can use an uncompressed bitmap, a simple list of integers,
or a list of runs. Whatever format it uses, they all allow you to check for the presence of any one value quickly
(e.g., with a binary search). The net result is that Roaring can compute many operations much faster than run-length-encoded
formats like WAH, EWAH, Concise... Maybe surprisingly, Roaring also generally offers better compression ratios.





### References

- Daniel Lemire, Owen Kaser, Nathan Kurz, Luca Deri, Chris O'Hara, François Saint-Jacques, Gregory Ssi-Yan-Kai, Roaring Bitmaps: Implementation of an Optimized Software Library, Software: Practice and Experience 48 (4), 2018 [arXiv:1709.07821](https://arxiv.org/abs/1709.07821)
-  Samy Chambi, Daniel Lemire, Owen Kaser, Robert Godin,
Better bitmap performance with Roaring bitmaps,
Software: Practice and Experience 46 (5), 2016.[arXiv:1402.6407](http://arxiv.org/abs/1402.6407) This paper used data from http://lemire.me/data/realroaring2014.html
- Daniel Lemire, Gregory Ssi-Yan-Kai, Owen Kaser, Consistently faster and smaller compressed bitmaps with Roaring, Software: Practice and Experience 46 (11), 2016. [arXiv:1603.06549](http://arxiv.org/abs/1603.06549)

### Dependencies

Dependencies are fetched automatically by giving the `-t` flag to `go get`.

they include
  - github.com/bits-and-blooms/bitset
  - github.com/mschoch/smat
  - github.com/glycerine/go-unsnap-stream
  - github.com/philhofer/fwd
  - github.com/jtolds/gls

Note that the smat library requires Go 1.15 or better.

#### Installation

  - go get -t github.com/RoaringBitmap/roaring

### Instructions for contributors

Using bash or other common shells:
```
$ git clone git@github.com:RoaringBitmap/roaring.git
$ export GO111MODULE=on 
$ go mod tidy
$ go test -v
```

### Example

Here is a simplified but complete example:

```go
package main

import (
    "fmt"
    "github.com/RoaringBitmap/roaring/v2"
    "bytes"
)


func main() {
    // example inspired by https://github.com/fzandona/goroar
    fmt.Println("==roaring==")
    rb1 := roaring.BitmapOf(1, 2, 3, 4, 5, 100, 1000)
    fmt.Println(rb1.String())

    rb2 := roaring.BitmapOf(3, 4, 1000)
    fmt.Println(rb2.String())

    rb3 := roaring.New()
    fmt.Println(rb3.String())

    fmt.Println("Cardinality: ", rb1.GetCardinality())

    fmt.Println("Contains 3? ", rb1.Contains(3))

    rb1.And(rb2)

    rb3.Add(1)
    rb3.Add(5)

    rb3.Or(rb1)

    // computes union of the three bitmaps in parallel using 4 workers  
    roaring.ParOr(4, rb1, rb2, rb3)
    // computes intersection of the three bitmaps in parallel using 4 workers  
    roaring.ParAnd(4, rb1, rb2, rb3)


    // prints 1, 3, 4, 5, 1000
    i := rb3.Iterator()
    for i.HasNext() {
        fmt.Println(i.Next())
    }
    fmt.Println()

    // next we include an example of serialization
    buf := new(bytes.Buffer)
    rb1.WriteTo(buf) // we omit error handling
    newrb:= roaring.New()
    newrb.ReadFrom(buf)
    if rb1.Equals(newrb) {
    	fmt.Println("I wrote the content to a byte stream and read it back.")
    }
    // you can iterate over bitmaps using ReverseIterator(), Iterator, ManyIterator()
}
```

If you wish to use serialization and handle errors, you might want to
consider the following sample of code:

```go
	rb := BitmapOf(1, 2, 3, 4, 5, 100, 1000)
	buf := new(bytes.Buffer)
	size,err:=rb.WriteTo(buf)
	if err != nil {
		fmt.Println("Failed writing") // return or panic
	}
	newrb:= New()
	size,err=newrb.ReadFrom(buf)
	if err != nil {
		fmt.Println("Failed reading") // return or panic
	}
	// if buf is an untrusted source, you should validate the result
	// (this adds a bit of complexity but it is necessary for security)
	if newrb.Validate() != nil {
		fmt.Println("Failed validation") // return or panic
	}
	if ! rb.Equals(newrb) {
		fmt.Println("Cannot retrieve serialized version")
	}
```

Given N integers in [0,x), then the serialized size in bytes of
a Roaring bitmap should never exceed this bound:

`` 8 + 9 * ((long)x+65535)/65536 + 2 * N ``

That is, given a fixed overhead for the universe size (x), Roaring
bitmaps never use more than 2 bytes per integer. You can call
``BoundSerializedSizeInBytes`` for a more precise estimate.

Call ``RunOptimize`` to convert containers to their most compact encoding,
including run-length encoding where it helps. Some operations that modify a
bitmap, such as ``AddRange`` and ``RemoveRange``, do not always convert a run
container back to an array or bitmap container when the runs stop being the
most compact encoding. In the worst case, the runs of one 65536-value chunk can
take about 128 KiB where a bitmap container takes 8 KiB. Calling ``RunOptimize``
again after such modifications recovers that space, in memory and once
serialized. This applies to 64-bit bitmaps (``roaring64``) as well.

### 64-bit Roaring

By default, roaring is used to stored unsigned 32-bit integers. However, we also offer
an extension dedicated to 64-bit integers. It supports roughly the same functions:

```go
package main

import (
    "fmt"
    "github.com/RoaringBitmap/roaring/v2/roaring64"
    "bytes"
)


func main() {
    // example inspired by https://github.com/fzandona/goroar
    fmt.Println("==roaring64==")
    rb1 := roaring64.BitmapOf(1, 2, 3, 4, 5, 100, 1000)
    fmt.Println(rb1.String())

    rb2 := roaring64.BitmapOf(3, 4, 1000)
    fmt.Println(rb2.String())

    rb3 := roaring64.New()
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
    newrb:= roaring64.New()
    newrb.ReadFrom(buf)
    if rb1.Equals(newrb) {
    	fmt.Println("I wrote the content to a byte stream and read it back.")
    }
    // you can iterate over bitmaps using ReverseIterator(), Iterator, ManyIterator()
}
```

Only the 32-bit roaring format is standard and cross-operable between Java, C++, C and Go. There is no guarantee that the 64-bit versions are compatible.

### Documentation

Current documentation is available at https://pkg.go.dev/github.com/RoaringBitmap/roaring and https://pkg.go.dev/github.com/RoaringBitmap/roaring/roaring64

### Goroutine safety

In general, it should not generally be considered safe to access
the same bitmaps using different goroutines--they are left
unsynchronized for performance. Should you want to access
a Bitmap from more than one goroutine, you should
provide synchronization. Typically this is done by using channels to pass
the *Bitmap around (in Go style; so there is only ever one owner),
or by using `sync.Mutex` to serialize operations on Bitmaps.

This applies even to operations that appear to be read-only when
copy-on-write is enabled (see `SetCopyOnWrite`). Calling `rb.Clone()`
updates internal bookkeeping in `rb`, and so do `x.Or(rb)` and `x.Xor(rb)`
when `x` also uses copy-on-write. Thus, several goroutines must not call
`Clone()` on the same copy-on-write bitmap at the same time without
synchronization.

### Coverage

We test our software. For a report on our test coverage, see

https://coveralls.io/github/RoaringBitmap/roaring?branch=master

### Benchmark

Type

         go test -bench Benchmark -run -

To run benchmarks on [Real Roaring Datasets](https://github.com/RoaringBitmap/real-roaring-datasets)
run the following:

```sh
go get github.com/RoaringBitmap/real-roaring-datasets
BENCH_REAL_DATA=1 go test -bench BenchmarkRealData -run -
```

### Interactive use

You can use roaring with gore:

- go install github.com/x-motemen/gore/cmd/gore@latest
- Make sure that ``$GOPATH/bin`` is in your ``$PATH``.

```go
$ gore
gore version 0.2.6  :help for help
gore> :import github.com/RoaringBitmap/roaring
gore> x:=roaring.New()
gore> x.Add(1)
gore> x.String()
"{1}"
```


### Fuzzy testing

You can help us test further the library with fuzzy testing:

         go get github.com/dvyukov/go-fuzz/go-fuzz
         go get github.com/dvyukov/go-fuzz/go-fuzz-build
         go test -tags=gofuzz -run=TestGenerateSmatCorpus
         go-fuzz-build github.com/RoaringBitmap/roaring
         go-fuzz -bin=./roaring-fuzz.zip -workdir=workdir/ -timeout=200 -func FuzzSmat

Let it run, and if the # of crashers is > 0, check out the reports in
the workdir where you should be able to find the panic goroutine stack
traces.

You may also replace `-func FuzzSmat`  by `-func FuzzSerializationBuffer` or `-func FuzzSerializationStream`.

### Alternative in Go

There is a Go version wrapping the C/C++ implementation https://github.com/RoaringBitmap/gocroaring

For an alternative implementation in Go, see https://github.com/fzandona/goroar
The two versions were written independently.


### Mailing list/discussion group

https://groups.google.com/g/roaring-bitmaps


### Users

Projects that depend on this Go package are listed first. [Other Roaring users](#other-roaring-users) use the Java library, the C/C++ library, or an implementation of their own.

#### Search

* [Blast](https://github.com/mosuka/blast). Full-text search server built on Bleve.
* [Bleve](https://github.com/blevesearch/bleve). Full-text search library for Go.
* [Gortex](https://github.com/zzet/gortex). Local code-intelligence engine for agents and editors.
* [Milvus](https://github.com/milvus-io/milvus). Cloud-native vector database.
* [Riot](https://github.com/vcaesar/riot). Distributed text and vector search engine.
* [rlama](https://github.com/DonTizi/rlama). Local document question answering on top of Ollama.
* [Sourcegraph](https://sourcegraph.com). Code search and code intelligence platform.
* [Weaviate](https://github.com/weaviate/weaviate). Open-source vector database.
* [ZincSearch](https://github.com/zincsearch/zincsearch). Lightweight search engine.
* [Zoekt](https://github.com/sourcegraph/zoekt). Trigram code search.

#### Databases

* [Apache Doris](https://github.com/apache/doris). Real-time analytics database.
* [Bitalostored](https://github.com/zuoyebang/bitalostored). Redis-compatible distributed storage system.
* [Dgraph](https://github.com/dgraph-io/dgraph). Graph database for real-time queries.
* [Elasticell](https://github.com/deepfabric/elasticell). Distributed key-value database with a Redis-compatible protocol.
* [FrostDB](https://github.com/polarsignals/frostdb). Embeddable columnar database, written for observability workloads.
* [MatrixOne](https://github.com/matrixorigin/matrixone). HTAP database with vector search.
* [Olric](https://github.com/olric-data/olric). Distributed in-memory key-value store and cache.
* [Pilosa](https://github.com/pilosa/pilosa). Distributed bitmap index.
* [Quanta](https://github.com/disney/quanta). Disney's database for mixed transactional and analytical queries, built on Roaring bitmaps.
* [SuperDB](https://github.com/superdb/super). Analytics database for JSON and relational tables.

#### Time series and observability

* [Grafana](https://github.com/grafana/grafana). Observability platform for metrics, logs, and traces.
* [Grafana Loki](https://github.com/grafana/loki). Log aggregation system.
* [Husky](https://www.datadoghq.com/blog/engineering/introducing-husky/). Datadog's columnar event store for observability data.
* [InfluxDB](https://www.influxdata.com). Time series database from InfluxData.
* [Kapacitor](https://github.com/influxdata/kapacitor). Framework for processing, monitoring, and alerting on time series data.
* [Kyanos](https://github.com/hengyoush/kyanos). eBPF tool for analyzing network latency.
* [LinDB](https://github.com/lindb/lindb). Distributed time series database.
* [M3](https://github.com/m3db/m3). Uber's metrics platform, including a distributed time series database and query engine.
* [Netcap](https://github.com/dreadl0ck/netcap). Framework for network traffic analysis.
* [Odigos](https://github.com/odigos-io/odigos). Distributed tracing without application code changes.
* [OpenMeter](https://github.com/openmeterio/openmeter). Usage metering and billing.
* [OpenTelemetry Collector Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib). Community components for the OpenTelemetry Collector.
* [openGemini](https://github.com/openGemini/openGemini). Distributed time series database.
* [Parca](https://github.com/parca-dev/parca). Continuous profiling for CPU and memory.
* [Perforator](https://github.com/yandex/perforator). Cluster-wide continuous profiling system from Yandex.
* [Vince](https://github.com/vinceanalytics/vince). Self-hosted web analytics.

#### Streaming and messaging

* [Apache Pulsar](https://github.com/apache/pulsar). Distributed publish-subscribe messaging system.
* [Bento](https://github.com/warpstreamlabs/bento). Stream processor.
* [Dapr](https://github.com/dapr/dapr). Portable runtime for distributed applications.
* [Mercure](https://github.com/dunglas/mercure). Hub for sending real-time updates to clients.
* [Redpanda Connect](https://github.com/redpanda-data/connect). Stream processor.
* [WuKongIM](https://github.com/WuKongIM/WuKongIM). Instant messaging server.

#### Files, sync, and media

* [AList](https://github.com/AlistGo/alist). File listing and WebDAV server for many storage backends.
* [AList Lite Android](https://github.com/LeoHaoVIP/AListLiteAndroid). Android build of the OpenList file server.
* [anacrolix/torrent](https://github.com/anacrolix/torrent). BitTorrent client package and utilities in Go.
* [autobrr](https://github.com/autobrr/autobrr). Download automation for torrents and Usenet.
* [Bitplay](https://github.com/aculix/bitplay). Browser player for video torrents.
* [Cloud Torrent](https://github.com/jpillora/cloud-torrent). Self-hosted remote torrent client.
* [Distribyted](https://github.com/distribyted/distribyted). Torrent client with HTTP, FUSE, and WebDAV interfaces.
* [fast-note-sync-service](https://github.com/haierkeys/fast-note-sync-service). Note synchronization service.
* [Go2TV](https://github.com/alexballas/go2tv). Casts local media to smart TVs and Chromecast devices.
* [Gopeed](https://github.com/GopeedLab/gopeed). Download manager for HTTP, BitTorrent, and other protocols.
* [Mangayomi](https://github.com/kodjodevf/mangayomi). Reader for manga, novels, and anime.
* [Nyaa](https://github.com/NyaaPantsu/nyaa). Torrent index.
* [OpenCloud](https://github.com/opencloud-eu/opencloud). File management, sharing, and collaboration platform.
* [OpenList](https://github.com/OpenListTeam/OpenList). File listing server, forked from AList.
* [ownCloud Infinite Scale](https://github.com/owncloud/ocis). ownCloud's file platform.
* [Pydio Cells](https://github.com/pydio/cells). Content collaboration platform.
* [Seanime](https://github.com/5rahim/seanime). Media server for anime and manga.
* [SeaweedFS](https://github.com/seaweedfs/seaweedfs). Distributed object and file store.
* [TorrServer](https://github.com/YouROK/TorrServer). Torrent streaming server.

#### Kubernetes, cloud, and infrastructure

* [Cloudpods](https://github.com/yunionio/cloudpods). Cloud platform for virtualization and Kubernetes.
* [E2B](https://github.com/e2b-dev/runtime). Runtime for E2B code sandboxes.
* [Fleet](https://github.com/fleetdm/fleet). Device management.
* [Harness](https://github.com/harness/harness). Developer platform for source control, CI/CD, and artifacts.
* [Kubernetes History Inspector](https://github.com/GoogleCloudPlatform/khi). Log viewer for Kubernetes troubleshooting.
* [Nginx UI](https://github.com/0xJacky/nginx-ui). Web interface for Nginx.
* [Portainer](https://github.com/portainer/portainer). Management interface for Docker and Kubernetes.
* [Semaphore](https://github.com/semaphoreio/semaphore). CI/CD platform.
* [StackRox](https://github.com/stackrox/stackrox). Kubernetes security platform.
* [Sylve](https://github.com/AlchemillaHQ/Sylve). Web interface for Bhyve, jails, and ZFS on FreeBSD.
* [Trident](https://github.com/NetApp/trident). NetApp storage provisioner for Kubernetes.
* [Velero](https://github.com/velero-io/velero). Backup and migration for Kubernetes applications.

#### AI and recommendations

* [Gorse](https://github.com/gorse-io/gorse). Recommender system engine.
* [LocalAGI](https://github.com/mudler/LocalAGI). Self-hosted agent platform.
* [LocalAI](https://github.com/mudler/LocalAI). Local inference server for language, vision, and audio models.
* [Obot](https://github.com/obot-platform/obot). Platform for governing AI agents.

#### Security and networking

* [Apinto](https://github.com/eolinker/apinto). API gateway.
* [BloodHound](https://github.com/SpecterOps/BloodHound). Tool for mapping attack paths in Active Directory.
* [Cerbos](https://github.com/cerbos/cerbos). Authorization for application requests.
* [Keybase](https://github.com/keybase/client). Keybase client.
* [Lantern](https://github.com/getlantern/lantern). VPN client for privacy and censorship circumvention.
* [Souin](https://github.com/darkweak/souin). HTTP cache.
* [The Things Stack](https://github.com/TheThingsNetwork/lorawan-stack). Open-source LoRaWAN network server.
* [Velociraptor](https://github.com/Velocidex/velociraptor). Digital forensics and incident response platform.

#### Blockchains

* [Cosmos SDK](https://github.com/cosmos/cosmos-sdk). Framework for building blockchains.
* [Erigon](https://github.com/erigontech/erigon). Ethereum execution client.
* [Harmony](https://github.com/harmony-one/harmony). Harmony protocol implementation.
* [IBC](https://github.com/cosmos/ibc). Inter-blockchain communication standards for Cosmos.
* [Ignite CLI](https://github.com/ignite/cli). Tooling for building Cosmos SDK chains.
* [IoTeX](https://github.com/iotexproject/iotex-core). IoTeX blockchain implementation.

#### Collaboration and content

* [bbs-go](https://github.com/mlogclub/bbs-go). Forum and question-and-answer platform.
* [Gitea](https://github.com/go-gitea/gitea). Self-hosted Git service.
* [git-bug](https://github.com/git-bug/git-bug). Distributed bug tracker stored in Git.
* [Mattermost](https://github.com/mattermost/mattermost). Team messaging platform.
* [OpenGist](https://github.com/thomiceli/opengist). Self-hosted pastebin.
* [PaoPao](https://github.com/rocboss/paopao-ce). Social community platform.

#### Runtimes, frameworks, and tools

* [Aether](https://github.com/aethereans/aether-app). Peer-to-peer community client.
* [Apache Dubbo Admin](https://github.com/apache/dubbo-admin). Operations console for Apache Dubbo.
* [Apache Dubbo Go](https://github.com/apache/dubbo-go). Go implementation of the Dubbo RPC framework.
* [Daptin](https://github.com/daptin/daptin). Self-hosted server for schema-driven APIs.
* [FrankenPHP](https://github.com/php/frankenphp). PHP application server.
* [go-daily-lib](https://github.com/darjun/go-daily-lib). Notes on one Go library a day.
* [Go-Spring](https://github.com/go-spring/go-spring). Application framework for Go.
* [ingestr](https://github.com/bruin-data/ingestr). Command-line tool for copying data between databases.
* [Inngest](https://github.com/inngest/inngest). Workflow orchestration for step functions.
* [IOC-golang](https://github.com/alibaba/IOC-golang). Dependency injection framework for Go.
* [Nakama](https://github.com/heroiclabs/nakama). Game server for multiplayer, matchmaking, and social features.
* [Nhost](https://github.com/nhost/nhost). Backend platform with GraphQL.
* [pico](https://github.com/picosh/pico). SSH-based services for developers.
* [Qlty](https://github.com/qltysh/qlty). Command-line tool for linting and code quality.
* [RoadRunner](https://github.com/roadrunner-server/roadrunner). PHP application server and process manager.
* [RudderStack](https://github.com/rudderlabs/rudder-server). Customer data pipeline.
* [Seata-go](https://github.com/apache/incubator-seata-go). Go implementation of Seata for distributed transactions.
* [WAL-G](https://github.com/wal-g/wal-g). Backup and restore for databases.

#### Other Roaring users

* [Apache Druid](https://druid.apache.org/). Real-time analytics database.
* [Apache Kylin](https://kylin.apache.org/). Distributed OLAP engine for big data, started at eBay.
* [Apache Lucene](https://lucene.apache.org/). High-performance full-text search library. Lucene ships its own Roaring implementation for sets of document identifiers.
* [Apache Pinot](https://pinot.apache.org/). Real-time distributed OLAP datastore, started at LinkedIn.
* [Apache Solr](https://solr.apache.org/). Search platform built on Lucene.
* [Apache Spark](https://spark.apache.org/). Unified engine for large-scale data processing.
* [Azure DevOps Services](https://azure.microsoft.com/products/devops). Microsoft's hosted DevOps service, formerly Visual Studio Team Services.
* [Elasticsearch](https://www.elastic.co/elasticsearch). Distributed search and analytics engine built on Lucene.
* [Google Procella](https://research.google/pubs/pub48388/). YouTube's SQL engine. It uses Roaring bitmaps for indexing.
* [Netflix Atlas](https://github.com/Netflix/atlas). In-memory dimensional time series database.
* [OpenSearchServer](http://www.opensearchserver.com). Open-source enterprise search engine.
* [Redpanda](https://github.com/redpanda-data/redpanda). Streaming data platform with a Kafka-compatible API. It uses the C/C++ library.
* [Whoosh](https://github.com/mchaput/whoosh). Pure-Python full-text search library, with its own Roaring implementation.


### Further reading

<p>Mastering Programming: From Testing to Performance in Go</p>
<div><a href="https://www.amazon.com/dp/B0FMPGSWR5"><img style="margin-left: auto; margin-right: auto;" src="https://m.media-amazon.com/images/I/61feneHS7kL._SL1499_.jpg" alt="" width="250px" /></a></div>
