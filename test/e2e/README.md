# AEIQ End-to-End Tests
This page serves as documentation for AEIQ's end-to-end tests.

## Regions
Instance-Qualifier e2e tests run in `us-east-2` region. To change test regions, update the Region-Specific Variables [here](https://github.com/awslabs/amazon-ec2-instance-qualifier/blob/main/test/e2e/run-tests#L34). Ensure the values meet the requirements in the comments.

## Prerequisites
1. You must have your AWS credentials configured. Take a look at the [AWS CLI configuration documentation](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html#config-settings-and-precedence) for details on the various ways to configure credentials.
2. To run the tests, please make sure your account is able to create the required number of resources. You can check [AWS Service Quota](https://docs.aws.amazon.com/servicequotas/latest/userguide/intro.html) for detailed information.

An overview of the resources created during e2e testing:

| Resource Type                                                     | # in us-east-2 Region |  Default Quota per Region|
|-------------------------------------------------------------------|---------------------|--------------------------|
| VPC                                                               | 4                   | 5                        |
| Internet Gateway                                                  | 3                   | 5                        |
| VCPUs of On-Demand Standard (A, C, D, H, I, M, R, T, Z) Instances | 34                  | 64 (after validation)    |

## Troubleshooting
By default, the e2e-tests destroy **all** artifacts including the CloudFormation stacks and associated S3 buckets. This paired with S3's [eventual consistency](https://aws.amazon.com/premiumsupport/knowledge-center/s3-listing-deleted-objects/) for
item deletion can make re-running the tests difficult due to S3 failing to create a new bucket thinking it already exists. Workarounds include creating a unique Bucket name for each run or commenting out the S3 creation/deletion code in the tests.
For users wanting to update the Bucket name, change the following:
* [run-tests](https://github.com/awslabs/amazon-ec2-instance-qualifier/blob/main/test/e2e/run-tests#L21)
  * `APP_BUCKET="ec2-instance-qualifier-app"` --> `APP_BUCKET="ec2-instance-qualifier-app-01"`
* [custom-script_sample](https://github.com/awslabs/amazon-ec2-instance-qualifier/blob/main/test/templates/custom-script_sample.template#L1)
  * `aws s3 cp s3://ec2-instance-qualifier-app/ec2-instance-qualifier-app .` --> `aws s3 cp s3://ec2-instance-qualifier-app-01/ec2-instance-qualifier-app .`
    * Note the application name does **not** change

## Disclaimer
All associated costs are the user's responsibility.
