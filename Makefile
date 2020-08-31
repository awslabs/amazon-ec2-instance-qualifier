MAKEFILE_PATH = $(dir $(realpath -s $(firstword $(MAKEFILE_LIST))))
BUILD_DIR_PATH = ${MAKEFILE_PATH}/build
TEMPLATES_DIR_PATH = ${MAKEFILE_PATH}/pkg/templates
SCRIPTS_DIR_PATH = ${MAKEFILE_PATH}/scripts
CLI_BINARY_NAME=ec2-instance-qualifier
AGENT_BINARY_NAME=agent
APP_BINARY_NAME=ec2-instance-qualifier-app
MASTER_TEMPLATE_VAR=github.com/awslabs/amazon-ec2-instance-qualifier/pkg/template.encodedMasterTemplate
LAUNCH_TEMPLATE_TEMPLATE_VAR=github.com/awslabs/amazon-ec2-instance-qualifier/pkg/template.encodedLaunchTemplateTemplate
AUTO_SCALING_GROUP_TEMPLATE_VAR=github.com/awslabs/amazon-ec2-instance-qualifier/pkg/template.encodedAutoScalingGroupTemplate
INSTANCE_TEMPLATE_VAR=github.com/awslabs/amazon-ec2-instance-qualifier/pkg/template.encodedInstanceTemplate
USER_DATA_TEMPLATE_VAR=github.com/awslabs/amazon-ec2-instance-qualifier/pkg/template.encodedUserData
MONITOR_CPU_SCRIPT_VAR=github.com/awslabs/amazon-ec2-instance-qualifier/pkg/setup.encodedMonitorCpuScript
MONITOR_MEM_SCRIPT_VAR=github.com/awslabs/amazon-ec2-instance-qualifier/pkg/setup.encodedMonitorMemScript
ENCODED_MASTER_TEMPLATE=$(shell cat ${TEMPLATES_DIR_PATH}/master.template | base64 | tr -d '\040\011\012\015')
ENCODED_LAUNCH_TEMPLATE_TEMPLATE=$(shell cat ${TEMPLATES_DIR_PATH}/launch-template.template | base64 | tr -d '\040\011\012\015')
ENCODED_AUTO_SCALING_GROUP_TEMPLATE=$(shell cat ${TEMPLATES_DIR_PATH}/auto-scaling-group.template | base64 | tr -d '\040\011\012\015')
ENCODED_INSTANCE_TEMPLATE=$(shell cat ${TEMPLATES_DIR_PATH}/instance.template | base64 | tr -d '\040\011\012\015')
ENCODED_USER_DATA_TEMPLATE=$(shell cat ${TEMPLATES_DIR_PATH}/user-data.template | base64 | tr -d '\040\011\012\015')
ENCODED_MONITOR_CPU_SCRIPT=$(shell cat ${SCRIPTS_DIR_PATH}/monitor-cpu.sh | base64 | tr -d '\040\011\012\015')
ENCODED_MONITOR_MEM_SCRIPT=$(shell cat ${SCRIPTS_DIR_PATH}/monitor-mem.sh | base64 | tr -d '\040\011\012\015')

$(shell mkdir -p ${BUILD_DIR_PATH} && touch ${BUILD_DIR_PATH}/_go.mod)

clean:
	rm -rf ${BUILD_DIR_PATH}
	rm -f ${MAKEFILE_PATH}/${AGENT_BINARY_NAME}
	rm -f ${MAKEFILE_PATH}/test/e2e/testdata/*.tar.gz
	rm -f ${MAKEFILE_PATH}/instance-qualifier-*.config
	rm -rf ${MAKEFILE_PATH}/test/e2e/tmp

compile:
	@echo ${MAKEFILE_PATH}
	go build -a -ldflags '-X "${MASTER_TEMPLATE_VAR}=${ENCODED_MASTER_TEMPLATE}" -X "${LAUNCH_TEMPLATE_TEMPLATE_VAR}=${ENCODED_LAUNCH_TEMPLATE_TEMPLATE}" -X "${AUTO_SCALING_GROUP_TEMPLATE_VAR}=${ENCODED_AUTO_SCALING_GROUP_TEMPLATE}" -X "${INSTANCE_TEMPLATE_VAR}=${ENCODED_INSTANCE_TEMPLATE}" -X "${MONITOR_CPU_SCRIPT_VAR}=${ENCODED_MONITOR_CPU_SCRIPT}" -X "${MONITOR_MEM_SCRIPT_VAR}=${ENCODED_MONITOR_MEM_SCRIPT}" -X "${USER_DATA_TEMPLATE_VAR}=${ENCODED_USER_DATA_TEMPLATE}"' -o ${BUILD_DIR_PATH}/${CLI_BINARY_NAME} ${MAKEFILE_PATH}/cmd/cli/ec2-instance-qualifier.go
	env GOOS=linux GOARCH=amd64 go build -o ${BUILD_DIR_PATH}/${AGENT_BINARY_NAME} ${MAKEFILE_PATH}/cmd/agent/agent.go
	cp -p ${BUILD_DIR_PATH}/${AGENT_BINARY_NAME} ${MAKEFILE_PATH}/${AGENT_BINARY_NAME}

build: compile

app:
	cd ${MAKEFILE_PATH}/test/app-for-e2e/cmd; \
	env GOOS=linux GOARCH=amd64 go build -gcflags '-N -l' -a -o ${BUILD_DIR_PATH}/${APP_BINARY_NAME} ec2-instance-qualifier-app.go

unit-test:
	go test -bench=. ${MAKEFILE_PATH}/... -v -coverprofile=coverage_aeiq.out -covermode=atomic -outputdir=${BUILD_DIR_PATH}
	cd ${MAKEFILE_PATH}/test/app-for-e2e; \
	go test -bench=. ./... -v -coverprofile=coverage_app.out -covermode=atomic -outputdir=${BUILD_DIR_PATH}

e2e-test: build app
	${MAKEFILE_PATH}/test/e2e/run-tests

readme-test:
	${MAKEFILE_PATH}/test/readme-test/run-readme-spellcheck

license-test:
	${MAKEFILE_PATH}/test/license-test/run-license-test.sh

go-report-card-test:
	${MAKEFILE_PATH}/test/go-report-card-test/run-report-card-test.sh

shellcheck:
	${MAKEFILE_PATH}/test/shellcheck/run-shellcheck

test: unit-test shellcheck readme-test license-test go-report-card-test e2e-test

fmt:
	goimports -w ./ && gofmt -s -w ./

help:
	@grep -E '^[a-zA-Z_-]+:.*$$' $(MAKEFILE_LIST) | sort