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

* Executes test suite on a range of EC2 instance types in parallel and persists test results and execution times
* Installs and configures [CloudWatch Agent](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Install-CloudWatch-Agent.html) on each instance type for capturing benchmark data
  * Instance-Qualifier uses the following for benchmarking: `cpu_usage_active` and `mem_used_percent`
  * More information on these metrics can be found [here](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/metrics-collected-by-CloudWatch-agent.html)
* Provides an ingress point for users to add their own logic to be executed in instance user data via `--custom-script` flag
* Supports asynchronous functionality, which means users can exit the CLI after tests begin and resume the session at a later time to fetch the results
* Uses [AWS CloudFormation](https://aws.amazon.com/cloudformation/) to manage all resources
* Creates an S3 bucket to store test results, instance logs, user configuration and CloudFormation template
* Implements mechanisms to ensure infrastructure deletion for various edge cases

## Impact to AWS Account

* The CLI creates a CloudFormation stack with a series of resources during the run and deletes the stack at the end by default. Resources include:
  * A **VPC + Subnet + Internet Gateway**: used to launch instances. Note that they are **only created if you don't specify `vpc`/`subnet` flags or provide invalid ones**
  * A **Security Group**: same as the default security group when you create one using AWS Console.  It has an inbounding rule which opens all ports for all traffic and all protocols, but the source must be within the same security group. With this rule, the instances can access the bucket, but won't be affected by any other traffic coming outside of the security group
  * An **IAM Role**: attached with AmazonS3FullAccess and CloudWatchAgentServerPolicy policies to allow instances to access the bucket and emit CloudWatch metrics, respectively
  * **Launch Templates**: used to launch auto scaling group and instances
  * An **Auto Scaling Group**: the reason we use auto scaling group to manage all instances is that an one-time action can be scheduled to terminate all instances in the group after timeout to ensure the user is not excessively charged
  * **EC2 Instances**
* An **S3 bucket** containing the raw data of an Instance-Qualifier run is also created; however, this artifact is persisted by default
* A sample of this CloudFormation stack can be found [here](https://github.com/awslabs/amazon-ec2-instance-qualifier/blob/master/pkg/templates/master_sample.template) 
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

**Note: the working directory where you execute `ec2-instance-qualifier` must contain the `agent` binary file**

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
        [OPTIONAL] the name of the Bucket created in the last run. When provided with this flag, the CLI won't create new resources, but try to grab test results from the Bucket. If you provide this flag, you don't need to specify any required flags
  -config-file string
        [OPTIONAL] path to config file for cli input parameters in JSON
  -custom-script string
        [OPTIONAL] path to Bash script to be executed on instance-types BEFORE agent runs test-suite and monitoring
  -instance-types string
        [REQUIRED] comma-separated list of instance-types to test
  -persist
        [OPTIONAL] set to true if you'd like the tool to keep the CloudFormation stack after the run. Default is deleting the stack
  -profile string
        [OPTIONAL] AWS CLI Profile to use for credentials and config
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

**Example 1: Test against m4.large and m4.xlarge with a target utilization of 80% in an existing VPC and subnet** *(logs not included in output below)*

```
$ ./ec2-instance-qualifier --instance-types=m4.large,m4.xlarge --test-suite=test-folder --target-utilization=80 --timeout=3600 --vpc=vpc-294b9542 --subnet=subnet-4879bf23
Region Used: us-east-2
Test Run ID: opcfxoss0uyxym4
Bucket Created: qualifier-bucket-opcfxoss0uyxym4
Stack Created: qualifier-stack-opcfxoss0uyxym4
The execution of test suite has been kicked off on all instances. You may quit now and later run the CLI again with the bucket name flag to get the result
+---------------+---------------------------+-------------------------+--------------------------+-----------------+----------------------------+
| INSTANCE TYPE | MEETS TARGET UTILIZATION? | CPU_USAGE_ACTIVE (p100) |  MEM_USED_PERCENT (p100) | ALL TESTS PASS? | TOTAL EXECUTION TIME (sec) |
+---------------+---------------------------+-------------------------+--------------------------+-----------------+----------------------------+
|   m4.large    |           FAIL            |         100.000         |          1.409           |      true       |          130.731           |
+---------------+---------------------------+-------------------------+--------------------------+-----------------+----------------------------+
|   m4.xlarge   |          SUCCESS          |         50.117          |          1.468           |      true       |          130.697           |
+---------------+---------------------------+-------------------------+--------------------------+-----------------+----------------------------+


Detailed test results can be found in s3://qualifier-bucket-opcfxoss0uyxym4/Instance-Qualifier-Run-opcfxoss0uyxym4
User configuration and CloudFormation template are stored in the root directory of the bucket. You may check them if you want
The process of cleaning up stack resources has started. You can quit now
Completed!
```

A unique ID is assigned to each test run and the bucket and stack names also contain the ID. From the results, we know that all instances ran the whole test suite successfully, but only m4.xlarge succeeded to meet the target utilization requirement.



**Example 2: Same as Example 1, but using a config file instead of CLI args**

```
$ cat iq-config.json
{
	"instance-types": "m4.large,m4.xlarge",
	"test-suite": "test-folder",
	"target-utilization": 80,
	"vpc": "vpc-294b9542",
	"subnet": "subnet-4879bf23",
	"ami": "",
	"timeout": 3600,
	"persist": false,
	"profile": "",
	"region": "us-east-2",
	"bucket": "",
	"custom-script": ""
}


$ ./ec2-instance-qualifier --config-file=iq-config.json
(Same output as Example 1)

```

**Example 3: Prompt due to an instance-type not supporting AMI**

```
$ ./ec2-instance-qualifier --instance-types=m4.xlarge,a1.large --test-suite=test-folder --target-utilization=95
Region Used: us-east-2
Test Run ID: n3lytbolzfaq3np
Bucket Created: qualifier-bucket-n3lytbolzfaq3np
Instance types [a1.large] are not supported due to AMI or Availability Zone. Do you want to proceed with the rest instance types [m5n.large] ? y/N
y
Stack Created: qualifier-stack-n3lytbolzfaq3np
The execution of test suite has been kicked off on all instances. You may quit now and later run the CLI again with the bucket name flag to get the result
```
The default AMI (Amazon Linux 2) is not compatible with `a1.large` architecture; therefore, the CLI prompts the user whether to continue the instance-qualifier run with compatible instance types only.


**Example 3.5: Exit CLI after stack creation, then resume**

```
(...continued from above)
The execution of test suite has been kicked off on all instances. You may quit now and later run the CLI again with the bucket name flag to get the result
^C

$ ./ec2-instance-qualifier --bucket=qualifier-bucket-n3lytbolzfaq3np
Region Used: us-east-2
Test Run ID: n3lytbolzfaq3np
Bucket Used: qualifier-bucket-n3lytbolzfaq3np
+---------------+---------------------------+-------------------------+--------------------------+-----------------+----------------------------+
| INSTANCE TYPE | MEETS TARGET UTILIZATION? | CPU_USAGE_ACTIVE (p100) |  MEM_USED_PERCENT (p100) | ALL TESTS PASS? | TOTAL EXECUTION TIME (sec) |
+---------------+---------------------------+-------------------------+--------------------------+-----------------+----------------------------+
|   m4.xlarge   |          SUCCESS          |         50.117          |          1.468           |      true       |          130.697           |
+---------------+---------------------------+-------------------------+--------------------------+-----------------+----------------------------+

Detailed test results can be found in s3://qualifier-bucket-n3lytbolzfaq3np/Instance-Qualifier-Run-n3lytbolzfaq3np
User configuration and CloudFormation template are stored in the root directory of the bucket. You may check them if you want
The process of cleaning up stack resources has started. You can quit now
^C
```
The CLI is interrupted after tests began executing on instances, then resumed by providing the bucket flag. Quitting before the *you may quit now* messaging results in both the CloudFormation stack and S3 bucket getting deleted.

## Interpreting Results

### Table Headers

* `INSTANCE TYPE`: instance type
* `MEETS TARGET UTILIZATION?`: SUCCESS if max CPU and max MEM are less than target utilization; FAIL otherwise
* `CPU_USAGE_ACTIVE`: max `cpu_usage_active` recorded (p100) over the duration of instance-qualifier run
* `MEM_USED_PERCENT`: max `mem_used_percent` recorded (p100) over the duration of instance-qualifier run
* `ALL TESTS PASS?`: true if **all** tests execute successfully (without an error code); false otherwise
* `TOTAL EXECUTION TIME`: how long it took the instance to execute all tests in seconds

## Building
For build instructions please consult [BUILD.md](./BUILD.md).

## Communication
If you've run into a bug or have a new feature request, please open an [issue](https://github.com/awslabs/amazon-ec2-instance-qualifier/issues/new).

##  Contributing
Contributions are welcome! Please read our [guidelines](./CONTRIBUTING.md) and our [Code of Conduct](./CODE_OF_CONDUCT.md).

## License
This project is licensed under the [Apache-2.0](LICENSE) License.