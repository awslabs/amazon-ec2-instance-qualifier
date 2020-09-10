<h1>Amazon EC2 Instance Qualifier</h1>

<h4>A CLI tool that automates benchmarking on a range of EC2 instance types.</h4>

<p>
	<a href="https://golang.org/doc/go1.14">
	<img src="https://img.shields.io/github/go-mod/go-version/awslabs/amazon-ec2-instance-qualifier?color=blueviolet" alt="go-version">
	</a>
	<a href="https://opensource.org/licenses/Apache-2.0">
	<img src="https://img.shields.io/badge/License-Apache%202.0-ff69b4.svg?color=orange" alt="license">
	</a>
	<a href="https://goreportcard.com/report/github.com/awslabs/amazon-ec2-instance-qualifier">
	<img src="https://goreportcard.com/badge/github.com/awslabs/amazon-ec2-instance-qualifier" alt="go-report-card">
	</a>
  <a href="https://travis-ci.com/awslabs/amazon-ec2-instance-qualifier">
	<img src="https://travis-ci.com/awslabs/amazon-ec2-instance-qualifier.svg?branch=master" alt="build-status">
  </a>
</p>




<div>
  <hr>
</div>

## Summary

How do users know which EC2 instance types are compatible with their application? Currently, there exists no tooling or baselining of any kind provided by AWS. If a user wants to see which of the 250+ different instance types are acceptable, then the user must spin up each instance type individually and test their application’s performance. Spot users often find themselves asking this question when they are told to utilize as many different instance types as possible in order to reduce the chance of spot interruptions. Still, most users will only ever choose a small subset of what could be acceptable due to the pain and cost of manual testing.

The instance qualifier is an open source command line tool that automates benchmarking on a range of EC2 instance types. The user will use the CLI to provide a test suite and a list of EC2 instance types. Instance qualifier will then run the input on all designated types, test against multiple metrics, and output the results in a user friendly format. In this way, instance qualifier will automate testing across instance types and address a severe pain point for spot users and EC2 users looking to venture into other instance type families.

## Major Features

* Executes test suite on a range of EC2 instance types in parallel while monitoring and collecting benchmark data. Supported benchmarks:
  * CPU load: average queue length of processes waiting to be executed or being running (given by `uptime`)
  * Memory used: amount of used memory (given by `free -m`)
* Supports asynchronous functionality, which means users can exit the CLI after tests begin and resume the session at a later time to fetch the results
* Uses [AWS CloudFormation](https://aws.amazon.com/cloudformation/) to manage all resources
* Creates an S3 bucket to store test results, instance logs, user configuration and CloudFormation template
* Implements mechanisms to ensure infrastructure deletion for various edge cases

## Impact to AWS Account

* The CLI creates a CloudFormation stack with a series of resources during the run and deletes the stack at the end by default. Resources include:
  * **A VPC + Subnet + Internet Gateway**: used to launch instances. Note that they are **only created if you don't specify `vpc`/`subnet` flags or provide invalid ones**
  * **A Security Group**: same as the default security group when you create one using AWS Console.  It has an inbounding rule which opens all ports for all traffic and all protocols, but the source must be within the same security group. With this rule, the instances can access the bucket, but won't be affected by any other traffic coming outside of the security group
  * **An IAM Role**: attached with AmazonS3FullAccess policy to allow instances to access the bucket without AWS credential configuration
  * **Launch Templates**: used to launch auto scaling group and instances
  * **An Auto Scaling Group**: the reason we use auto scaling group to manage all instances is that an one-time action can be scheduled to terminate all instances in the group after timeout to ensure the user is not excessively charged
  * **EC2 Instances**
* An S3 bucket is also created and won't be deleted after the run, which contains the raw data if deep dive is needed
* If a fatal error occurs or the user presses Ctrl-C during the run, the CLI deletes the resources appropriately. Note that if the CLI is interrupted when the tests have begun on all instances, it thinks that the user may resume the session at a later time, thus won't delete any resources
* No impact to any original resources or settings of the AWS account

**Disclaimer: All associated costs are the user's responsibility.**

## Configuration

To execute the CLI, you will need AWS credentials configured. Take a look at the [AWS CLI configuration documentation](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html#config-settings-and-precedence) for details on the various ways to configure credentials. An easy way to try out the ec2-instance-qualifier CLI is to populate the following environment variables with your AWS API credentials.

```
export AWS_ACCESS_KEY_ID="..."
export AWS_SECRET_ACCESS_KEY="..."
```

If you already have an AWS CLI profile setup, you can pass that directly into ec2-instance-qualifier:

```
$ ./ec2-instance-qualifier --instance-types=m4.large --test-suite=test-folder --target-utilization=30 --profile=my-aws-cli-profile --region=us-east-1
```

You can set the `AWS_REGION` environment variable if you don't want to pass in `--region` on each run.

```
export AWS_REGION="us-east-1"
```

## Examples

**Note: the working directory where you execute `ec2-instance-qualifier` must contains the `agent` binary file**

**All CLI Options**

```bash#help
$ ./ec2-instance-qualifier --help
ec2-instance-qualifier is a CLI tool that automates testing on a range of EC2 instance types.
Provided with a test suite and a list of EC2 instance types, ec2-instance-qualifier will then
run the input on all designated types, test against multiple metrics, and output the results
in a user friendly format

Usage:
  ec2-instance-qualifier [flags]

Examples:
./ec2-instance-qualifier --instance-types=m4.large,c5.large,m4.xlarge --test-suite=path/to/test-folder --target-utilization=30 --vpc=vpc-294b9542 --subnet=subnet-4879bf23 --timeout=2400
./ec2-instance-qualifier --instance-types=m4.xlarge,c1.large,c5.large --test-suite=path/to/test-folder --target-utilization=50 --profile=default
./ec2-instance-qualifier --bucket=qualifier-bucket-123456789abcdef

Flags:
  -ami string
        [OPTIONAL] ami id
  -bucket string
        [OPTIONAL] the name of the bucket created in the last run. When provided with this flag, the CLI won't create new resources, but try to grab test results from the bucket. If you provide this flag, you don't need to specify any required flags
  -custom-script string
        [OPTIONAL] path to Bash script to be executed on instance-types BEFORE agent runs test-suite and monitoring
  -instance-types string
        [REQUIRED] comma-separated list of instance-types to test
  -persist
        [OPTIONAL] set to true if you'd like the tool to keep the CloudFormation stack after the run. Default is deleting the stack
  -profile string
        [OPTIONAL] AWS CLI profile to use for credentials and config
  -region string
        [OPTIONAL] AWS Region to use for API requests
  -subnet string
        [OPTIONAL] subnet id
  -target-utilization int
        [REQUIRED] % of total resource used (CPU, Mem) benchmark (must be an integer). ex: 30 means instances using less than 30% CPU and Mem SUCCEED
  -test-suite string
        [REQUIRED] folder containing test files to execute
  -timeout int
        [OPTIONAL] max seconds for test-suite execution on instances (default 3600)
  -vpc string
        [OPTIONAL] vpc id
```

**Example 1: test files in test-folder against m4.large, c5.large and m4.xlarge with a target utilization of 50% in a specified VPC and subnet** *(logs not included in output below)*

```
$ ./ec2-instance-qualifier --instance-types=m4.large,c5.large,m4.xlarge --test-suite=test-folder --target-utilization=50 --vpc=vpc-294b9542 --subnet=subnet-4879bf23
Region Used: us-east-2
Test Run ID: opcfxoss0uyxym4
Bucket Created: qualifier-bucket-opcfxoss0uyxym4
Stack Created: qualifier-stack-opcfxoss0uyxym4
The execution of test suite has been kicked off on all instances. You may quit now and later run the CLI again with the bucket name flag to get the result
+---------------+---------------------------+-------------+----+---------------+----+-----------------+----------------------------+
| INSTANCE TYPE | MEETS TARGET UTILIZATION? | MAX CPU (n) | %  | MAX MEM (MiB) | %  | ALL TESTS PASS? | TOTAL EXECUTION TIME (sec) |
+---------------+---------------------------+-------------+----+---------------+----+-----------------+----------------------------+
|   c5.large    |           FAIL            |    1.667    | 83 |     3188      | 78 |      true       |          130.590           |
+---------------+---------------------------+-------------+----+---------------+----+-----------------+----------------------------+
|   m4.xlarge   |          SUCCESS          |    1.677    | 42 |     3211      | 20 |      true       |          130.722           |
+---------------+---------------------------+-------------+----+---------------+----+-----------------+----------------------------+
|   m4.large    |           FAIL            |    1.687    | 84 |     3193      | 39 |      true       |          130.703           |
+---------------+---------------------------+-------------+----+---------------+----+-----------------+----------------------------+

Detailed test results can be found in s3://qualifier-bucket-opcfxoss0uyxym4/Instance-Qualifier-Run-opcfxoss0uyxym4
User configuration and CloudFormation template are stored in the root directory of the bucket. You may check them if you want
The process of cleaning up stack resources has started. You can quit now
Completed!
```

**Example 2: test in a new VPC infrastructure with a timeout of 125 seconds** *(logs not included in output below)*

```
$ ./ec2-instance-qualifier --instance-types=c4.large,m5.large,m4.xlarge --test-suite=test-folder --target-utilization=50 --timeout=125
Region Used: us-east-2
Test Run ID: 72suqu0ra0t7t1a
Bucket Created: qualifier-bucket-72suqu0ra0t7t1a
Stack Created: qualifier-stack-72suqu0ra0t7t1a
The execution of test suite has been kicked off on all instances. You may quit now and later run the CLI again with the bucket name flag to get the result
+---------------+---------------------------+-------------+----+---------------+---+-----------------+----------------------------+
| INSTANCE TYPE | MEETS TARGET UTILIZATION? | MAX CPU (n) | %  | MAX MEM (MiB) | % | ALL TESTS PASS? | TOTAL EXECUTION TIME (sec) |
+---------------+---------------------------+-------------+----+---------------+---+-----------------+----------------------------+
|   c4.large    |           FAIL            |    1.217    | 61 |      83       | 2 |      false      |          120.010           |
+---------------+---------------------------+-------------+----+---------------+---+-----------------+----------------------------+
|   m5.large    |           FAIL            |    1.262    | 63 |      92       | 1 |      false      |          120.027           |
+---------------+---------------------------+-------------+----+---------------+---+-----------------+----------------------------+
|   m4.xlarge   |          SUCCESS          |    1.389    | 35 |      109      | 1 |      false      |          120.020           |
+---------------+---------------------------+-------------+----+---------------+---+-----------------+----------------------------+

Detailed test results can be found in s3://qualifier-bucket-72suqu0ra0t7t1a/Instance-Qualifier-Run-72suqu0ra0t7t1a
User configuration and CloudFormation template are stored in the root directory of the bucket. You may check them if you want
The process of cleaning up stack resources has started. You can quit now
Completed!
```

**Example 3: test against unsupported instance types and exit the CLI after tests begin, then resume the session at a later time to fetch the results** *(logs not included in output below)*

```
$ ./ec2-instance-qualifier --instance-types=m5n.large,a1.large --test-suite=test-folder --target-utilization=95
Region Used: us-east-2
Test Run ID: n3lytbolzfaq3np
Bucket Created: qualifier-bucket-n3lytbolzfaq3np
Instance types [a1.large] are not supported due to AMI or Availability Zone. Do you want to proceed with the rest instance types [m5n.large] ? y/N
y
Stack Created: qualifier-stack-n3lytbolzfaq3np
The execution of test suite has been kicked off on all instances. You may quit now and later run the CLI again with the bucket name flag to get the result
^C
```

```
$ ./ec2-instance-qualifier --bucket=qualifier-bucket-n3lytbolzfaq3np
Region Used: us-east-2
Test Run ID: n3lytbolzfaq3np
Bucket Used: qualifier-bucket-n3lytbolzfaq3np
+---------------+---------------------------+-------------+----+---------------+----+-----------------+----------------------------+
| INSTANCE TYPE | MEETS TARGET UTILIZATION? | MAX CPU (n) | %  | MAX MEM (MiB) | %  | ALL TESTS PASS? | TOTAL EXECUTION TIME (sec) |
+---------------+---------------------------+-------------+----+---------------+----+-----------------+----------------------------+
|   m5n.large   |          SUCCESS          |    1.697    | 85 |     3195      | 39 |      true       |          130.613           |
+---------------+---------------------------+-------------+----+---------------+----+-----------------+----------------------------+

Detailed test results can be found in s3://qualifier-bucket-n3lytbolzfaq3np/Instance-Qualifier-Run-n3lytbolzfaq3np
User configuration and CloudFormation template are stored in the root directory of the bucket. You may check them if you want
The process of cleaning up stack resources has started. You can quit now
^C
```

## Interpreting Results

### Table Headers

* `INSTANCE TYPE`: instance type
* `MEETS TARGET UTILIZATION?`: SUCCESS if max CPU and max MEM are less than target utilization; FAIL otherwise
* `MAX CPU`: highest CPU load average recorded out of all completed test runs
* `%`: MAX CPU as a percentage of number of VCPUs
* `MAX MEM`: highest memory used average recorded out of all completed test runs
* `%`: MAX MEM as percentage of physical memory size
* `ALL TESTS PASS?`: true if **all** tests execute successfully (without an error code); false otherwise
* `TOTAL EXECUTION TIME`: how long it took the instance to execute all tests in seconds

### Interpretation of Example Results

**Example 1**

A unique ID is assigned to each test run and the bucket and stack names also contain the ID. From the results, we know that all instances ran the whole test suite successfully, but only m4.xlarge succeeded to meet the target utilization requirement.

**Example 2**

All instances failed to pass all tests due to timeout. And since partial results are not recorded, the values of `MAX CPU` + `%`, `MAX MEM` + `%`, `TOTAL EXECUTION TIME` and `MEETS TARGET UTILIZATION?` were calculated based on only the finished tests. 

**Example 3**

First, `a1.large` doesn't support x86_64, which is the architecture of the default AMI (Amazon Linux 2), so the CLI asked the user whether to continue the tests with the rest instance types. Then the CLI was interrupted when the tests began on instances, and later resumed by providing the bucket flag. Finally when the CLI said "You can quit now", the user pressed Ctrl-C again, which wouldn't interrupt the asynchronous deletion of resources.

## Building
For build instructions please consult [BUILD.md](./BUILD.md).

## Communication
If you've run into a bug or have a new feature request, please open an [issue](https://github.com/awslabs/amazon-ec2-instance-qualifier/issues/new).

##  Contributing
Contributions are welcome! Please read our [guidelines](./CONTRIBUTING.md) and our [Code of Conduct](./CODE_OF_CONDUCT.md).

## License
This project is licensed under the [Apache-2.0](LICENSE) License.