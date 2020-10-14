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

package cmdutil

import (
	"net/http"

	"github.com/awslabs/amazon-ec2-instance-qualifier/ec2-instance-qualifier-app/pkg/handlers"
)

// RegisterHandlers associates paths with their respective handlers.
func RegisterHandlers() {
	http.HandleFunc("/", handlers.ListRoutesHandler)
	http.HandleFunc("/cpu", handlers.CPULoadHandler)
	http.HandleFunc("/newmem", handlers.NewMemLoadHandler)
	http.HandleFunc("/pet", handlers.PetHandler)
	http.HandleFunc("/pupulate", handlers.PupulateHandler)
	http.HandleFunc("/depupulate", handlers.DepupulateHandler)
}
