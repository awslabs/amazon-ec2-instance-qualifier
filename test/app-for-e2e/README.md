# EC2-Instance-Qualifier-App

**EC2-Instance-Qualifier-App** is a basic web server used to simulate work on ec2 instances to collect benchmark data.

The server's purpose is to simulate a real, working application for the end-to-end tests in Amazon-ec2-instance-qualifier project.

## Supported Paths

* **/**: lists the available tests
* **/cpu**: cpu stress test that creates 2 "endless" loop goroutines until it's time to abort
  * supported parameters: seconds
    * number of seconds until endless loops are aborted
    * default value is *10 seconds*
* **/mem**: mem stress test that creates simple goroutines running in parallel
  * supported parameters: routines
    * number of goroutines to create and run in parallel
    * default value is *100,000 goroutines*
* **/newmem**: mem stress test that allocates the specified amount of physical memory
  * supported parameters: mib
    * amount of physical memory that is allocated
    * default value is *1,000 MiB*

## Usage

First start the server

```
$ ./path/to/ec2-instance-qualifier-app
```

**Execute the cpu stress test with a duration of 12 seconds**

```
$ curl "localhost:1738/cpu?seconds=12"
```

**Execute the mem stress test that creates 200,000 goroutines**

```
$ curl "localhost:1738/mem?routines=200000"
```

**Execute the mem stress test that allocates 3,000 MiB physical memory**

```
$ curl "localhost:1738/newmem?mib=3000"
```
