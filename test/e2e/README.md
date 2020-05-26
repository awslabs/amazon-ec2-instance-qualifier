# AEIQ End-to-End Tests
This page serves as documentation for AEIQ's end-to-end tests.

## Regions
Tests run the instance-qualifier in 2 regions, `us-east-2` as the default region and `us-west-1` as the non-default region. If you want to run tests in different regions, please change the Region-Specific Variables in [run-tests](./run-tests) and ensure the values meet the requirements in the comments.

## Prerequisites
1. You must have your AWS credentials configured. Take a look at the [AWS CLI configuration documentation](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html#config-settings-and-precedence) for details on the various ways to configure credentials.
2. During the tests run, some resources with low [AWS Service Quota](https://docs.aws.amazon.com/servicequotas/latest/userguide/intro.html) are created simultaneously, including:

| Resource Type                                                     | # in Default Region | # in Non-default Region | Default Quota per Region |
|-------------------------------------------------------------------|---------------------|-------------------------|--------------------------|
| VPC                                                               | 4                   | 1                       | 5                        |
| Internet Gateway                                                  | 3                   | 1                       | 5                        |
| VCPUs of On-Demand Standard (A, C, D, H, I, M, R, T, Z) Instances | 34                  | 8                       | 64 (after validation)    |
To run the tests, please make sure your account is able to create the required number of resources. You can check [AWS Service Quota](https://docs.aws.amazon.com/servicequotas/latest/userguide/intro.html) for detailed information.

## Disclaimer

All associated costs are the user's responsibility.