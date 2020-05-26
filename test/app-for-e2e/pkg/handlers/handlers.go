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

package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"syscall"
	"time"
)

// CPULoadHandler creates 2 goroutines and waits on "endless" loop until sleep duration passes.
// Duration is provided as a URL parameter or defaulted to 10.
// ex: localhost:1738/cpu?seconds=5
func CPULoadHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("Entered CPULoadHandler")
	flusher, ok := res.(http.Flusher)
	if !ok {
		http.Error(res, "Server does not support flusher", http.StatusInternalServerError)
		return
	}

	sleepTime := parseOrDefault(req, "seconds", 10)
	fmt.Fprintf(res, "Starting to sleep CPUs for %d seconds 💤\n", sleepTime)
	flusher.Flush()

	done := make(chan int)
	for i := 0; i < 2; i++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
				}
			}
		}()
	}

	time.Sleep(time.Duration(sleepTime) * time.Second)
	close(done)

	fmt.Fprint(res, "CPUs awake!! 🌞️\n")
	log.Println("Leaving CPULoadHandler")
}

// MemLoadHandler launches X goroutines concurrently, independent of one another.
// X is provided as a URL parameter or defaulted to 100,000.
// A single Goroutine is ~8KB; therefore, with the default value ~800MB of memory is expected to be used.
// ex: localhost:1738/mem?routines=120000
func MemLoadHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("Entered MemLoadHandler")
	flusher, ok := res.(http.Flusher)
	if !ok {
		http.Error(res, "Server does not support flusher", http.StatusInternalServerError)
		return
	}

	numGoRoutines := parseOrDefault(req, "routines", 100000)

	fmt.Fprintf(res, "Starting mem test with %d goroutines ۞\n", numGoRoutines)
	flusher.Flush()

	for i := 0; i < numGoRoutines; i++ {
		go func() {
			log.Println("goroutine-ing")
		}()
	}

	fmt.Fprint(res, "All goroutines finished!! 🏁\n")
	log.Println("Leaving MemLoadHandler")
}

// NewMemLoadHandler allocates X MiB physical memory and waits for 10 seconds before ending.
// X is provided as a URL parameter or defaulted to 1,000.
// ex: localhost:1738/newmem?mib=2000
func NewMemLoadHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("Entered NewMemLoadHandler")
	flusher, ok := res.(http.Flusher)
	if !ok {
		http.Error(res, "Server does not support flusher", http.StatusInternalServerError)
		return
	}

	memUsed := parseOrDefault(req, "mib", 1000)

	fmt.Fprintf(res, "Starting new mem test with %d MiB memory used ۞\n", memUsed)
	flusher.Flush()

	memory := make([]byte, memUsed*1024*1024)
	err := syscall.Mlock(memory)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	time.Sleep(10 * time.Second)

	fmt.Fprint(res, "New memory test finished!! 🏁\n")
	log.Println("Leaving NewMemLoadHandler")
}

// ListRoutesHandler returns the valid paths on this server.
func ListRoutesHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("Entered ListRoutesHandler")
	paths := []string{"cpu", "mem"}
	for _, path := range paths {
		fmt.Fprintln(res, path)
	}
	log.Println("Leaving ListRoutesHandler")
}

// parseOrDefault takes a requestParam key and parses into an int if parsing fails, then assign to defaultValue.
func parseOrDefault(req *http.Request, requestParam string, defaultValue int) int {
	result := defaultValue
	paramAsInt, ok := req.URL.Query()[requestParam]
	if !ok || len(paramAsInt[0]) < 1 {
		log.Printf("no parameter provided in request --> defaulting to %d \n", defaultValue)
	} else {
		intValue, err := strconv.Atoi(paramAsInt[0])
		if err != nil {
			log.Printf("something went wrong parsing the parameter --> defaulting to %d \n", defaultValue)
			log.Println(err)
		} else {
			result = intValue
		}
	}

	return result
}
