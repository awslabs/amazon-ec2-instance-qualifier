// Copyright 2020 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/awslabs/amazon-ec2-instance-qualifier/ec2-instance-qualifier-app/pkg/cmdutil"
)

// Server defaults
const (
	hostname = "0.0.0.0"
	port     = "1738"
)

func main() {
	host := fmt.Sprint(hostname, ":", port)
	cmdutil.RegisterHandlers()
	log.Printf("Server started on: %s\n", host)
	if err := http.ListenAndServe(host, nil); err != nil {
		panic(err)
	}
}
