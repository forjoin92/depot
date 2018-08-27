# depot
a store by hashicorp/raft for demonstrate

## Getting Started

### Installing

To start using depot, install Go and run `go get -u`:

```sh
$ go get -u github.com/forjoin92/depot
```

This will retrieve the library and install the `depot` command line utility into
your `$GOBIN` path.

### Starting one member of depot

```sh
./depot -cluster 127.0.0.1:30401 -id 127.0.0.1:30401 -testAddr 127.0.0.1 -testPort 9001
```

Store a value ("value1") to a key ("key1"):

```sh
curl -L http://127.0.0.1:9001/setKV -XPUT -d {"key1":"value1"}
```

Retrieve the stored key:

```sh
curl -L http://127.0.0.1:9001/getKV/key1
```

Delete the stored key:

```sh
curl -L http://127.0.0.1:9001/deleteKV/key1 -XDELETE
```

Add one member of depot:

```sh
curl -L http://127.0.0.1:9001/addNode -XPOST -d 127.0.0.1:30402

./depot -cluster 127.0.0.1:30402 -id 127.0.0.1:30402 -testAddr 127.0.0.1 -testPort 9002
```

### Starting a cluster of depot

```sh
./depot -cluster 127.0.0.1:30401,127.0.0.1:30402 -id 127.0.0.1:30401 -testAddr 127.0.0.1 -testPort 9001

./depot -cluster 127.0.0.1:30401,127.0.0.1:30402 -id 127.0.0.1:30402 -testAddr 127.0.0.1 -testPort 9002
```

Delete one member of depot's cluster:

```sh
curl -L http://127.0.0.1:9001/removeNode -XDELETE 127.0.0.1:30402
```
