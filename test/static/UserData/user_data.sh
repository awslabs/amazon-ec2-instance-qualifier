#!/usr/bin/env bash
aws s3 cp s3://ec2-instance-qualifier-app/ec2-instance-qualifier-app .
chmod u+x ec2-instance-qualifier-app
./ec2-instance-qualifier-app >/dev/null 2>/dev/null &

INSTANCE_TYPE=m4.large
VCPUS_NUM=2
MEM_SIZE=8192
OS_VERSION=Linux/UNIX
ARCHITECTURE=x86_64
BUCKET=
TIMEOUT=0
BUCKET_ROOT_DIR=
TARGET_UTIL=0

adduser qualifier
cd /home/qualifier || :
mkdir instance-qualifier
cd instance-qualifier || :
aws s3 cp s3:///. .
tar -xvf .
cd .
for file in *; do
	if [[ -f "$file" ]]; then
		chmod u+x "$file"
	fi
done
cd ../..
chown -R qualifier instance-qualifier
chmod u+s /sbin/shutdown
sudo -i -u qualifier bash << EOF
cd instance-qualifier/.
./agent "$INSTANCE_TYPE" "$VCPUS_NUM" "$MEM_SIZE" "$OS_VERSION" "$ARCHITECTURE" "$BUCKET" "$TIMEOUT" "$BUCKET_ROOT_DIR" "$TARGET_UTIL" > m4.large.log 2>&1 &
EOF
