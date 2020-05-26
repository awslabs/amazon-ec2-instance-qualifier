# Amazon EC2 Instance Qualifier: Build Instructions

## Install Go version 1.14+

There are several options for installing go:

1. If you're on mac, you can simply `brew install go`
2. If you'd like a flexible go installation manager consider using gvm https://github.com/moovweb/gvm
3. For all other situations use the official go getting started guide: https://golang.org/doc/install

## Compile

This project uses `make` to organize compilation, build, and test targets.

To compile cmd/cli/ec2-instance-qualifier.go and cmd/agent/agent.go, which will build the full static binary and pull in dependent packages, run:

```
$ make build
```

The resulting binary will be in the generated `build/` dir

```
$ make build

$ ls build/
agent
ec2-instance-qualifier
```

## Test

You can execute the unit tests for the instance qualifier with `make`:

```
$ make unit-test
```


### Run All Tests

The full suite includes license-test, go-report-card, integration tests, and more. See the full list in the [makefile](./Makefile). NOTE: some tests require AWS Credentials to be configured on the system:

```
$ make test
```

## Format

To keep our code readable with go conventions, we use `goimports` to format the source code.
Make sure to run `goimports` before you submit a PR or you'll be caught by our tests! 

You can use the `make fmt` target as a convenience

```
$ make fmt
```

## See All Make Targets

To see all possible make targets and dependent targets, run:

```
$ make help
app:
build: compile
clean:
compile:
fmt:
go-report-card-test:
help:
license-test:
readme-test:
test: unit-test readme-test license-test go-report-card-test e2e-test
unit-test:
```
