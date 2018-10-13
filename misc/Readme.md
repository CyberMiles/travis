# Generate genesis

we can use genesis related tools to generate genesis.josn or genesis allocates for dev and mainnet.

### 1. generate genesis.json
use `gen_genesis/main.go` to generate genesis output, then we can redirect the output to somewhere you want.

* development genesis output

```bash
$ cd gen_genesis
$ go run ./main.go dev > ../genesis/dev_genesis.json
```

* mainnet genesis output

```bash
$ cd gen_genesis
$ go run ./main.go mainnet > ../genesis/mainnet_genesis.json
```

the genesis.josn can be used to `--vm-genesis string   VM genesis file` when init the node.
if we omit the `--vm-genesis`, the app will use default config.
the default config come from the genesis.go and *_alloc.go.

### 2. how to make allocates
use `genesis/mkalloc.go` to make allocates, example as follow

```apple js
$ cd genesis
$ go run ./mkalloc.go devAllocData ./dev_genesis.go > dev_alloc.go
```