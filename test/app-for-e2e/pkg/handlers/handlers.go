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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"syscall"
	"time"

	db "github.com/awslabs/amazon-ec2-instance-qualifier/ec2-instance-qualifier-app/pkg/database"
)

// E2E Testing

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

// ListRoutesHandler returns available paths
func ListRoutesHandler(res http.ResponseWriter, req *http.Request) {
	paths := "paths: / , cpu , newmem , pet , pupulate , depupulate"
	messageResponseJSON(res, http.StatusOK, paths)
	return
}

// DEMO APP

// MessageResponse contains any additional information required in the server response
type MessageResponse struct {
	Message string `json:"message"`
}

const (
	petIdParam = "petId"
	numParam   = "num"
)

// PetHandler handles pets and pet accessories
// ex: ex: localhost:1738/pet , ex: localhost:1738/pet?petId=4
func PetHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		handlePetGet(res, req)
	case http.MethodPost:
		handlePetPost(res, req)
	case http.MethodDelete:
		handlePetDelete(res, req)
	default:
		message := "Malformed request; unsupported http method. Supported methods: Get, Post, Delete"
		messageResponseJSON(res, http.StatusBadRequest, message)
	}
}

func handlePetGet(res http.ResponseWriter, req *http.Request) {
	petIdParameter, _ := req.URL.Query()[petIdParam]
	// if GET and no petId, then treat as a request for total table count
	if len(petIdParameter) < 1 {
		count, err := db.GetPetCount()
		if err != nil {
			message := "Could not fetch total number of pets: " + err.Error()
			messageResponseJSON(res, http.StatusBadRequest, message)
			return
		}
		messageResponseJSON(res, http.StatusOK, strconv.FormatInt(count, 10))
		return
	}
	petId := petIdParameter[0]
	pet, err := db.GetPetByID(petId)
	if err != nil {
		message := "Pet not found: " + err.Error()
		messageResponseJSON(res, http.StatusNotFound, message)
		return
	}
	jsonResponse(res, http.StatusOK, pet)
}

func handlePetPost(res http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		message := "Invalid request; body required"
		messageResponseJSON(res, http.StatusMethodNotAllowed, message)
		return
	}
	var pet db.Pet
	err = json.Unmarshal(body, &pet)
	if err != nil {
		log.Println(err)
		message := "Invalid request; can't unmarshal"
		messageResponseJSON(res, http.StatusMethodNotAllowed, message)
		return
	}
	addedId, err := db.AddPet(pet)
	if err != nil {
		message := "Invalid request; can't add pet to DB"
		messageResponseJSON(res, http.StatusMethodNotAllowed, message)
		return
	}
	jsonResponse(res, http.StatusOK, addedId)
}

func handlePetDelete(res http.ResponseWriter, req *http.Request) {
	petIdParameter, _ := req.URL.Query()[petIdParam]
	if len(petIdParameter) < 1 {
		message := "Pet ID required"
		messageResponseJSON(res, http.StatusBadRequest, message)
		return
	}
	petId := petIdParameter[0]
	err := db.DeletePet(petId)
	if err != nil {
		message := "Could remove pet from DB"
		messageResponseJSON(res, http.StatusNotFound, message)
		return
	}
	jsonResponse(res, http.StatusOK, "Pet Deleted")
}

// PupulateHandler populates the Pets table with pups
// ex: localhost:1738/pupulate?num=1000
func PupulateHandler(res http.ResponseWriter, req *http.Request) {
	numPups := parseOrDefault(req, numParam, 100)
	pupsAdded, err := db.PopulateTable(numPups)
	if err != nil {
		message := "Could not populate Pets table"
		messageResponseJSON(res, http.StatusBadRequest, message)
		return
	}
	jsonResponse(res, http.StatusOK, pupsAdded)
	return
}

// DepupulateHandler deletes a number of pups from the Pets table
// ex: localhost:1738/depupulate?num=1000
func DepupulateHandler(res http.ResponseWriter, req *http.Request) {
	numPups := parseOrDefault(req, numParam, 100)
	err := db.DeleteEntries(numPups)
	if err != nil {
		message := "Could not delete from Pets table " + err.Error()
		messageResponseJSON(res, http.StatusBadRequest, message)
		return
	}
	jsonResponse(res, http.StatusOK, "Entries from pets table deleted")
	return
}

// Helpers

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

func messageResponseJSON(res http.ResponseWriter, status int, message string) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(status)
	msg := MessageResponse{Message: message}
	mJSON, err := json.Marshal(msg)
	if err != nil {
		log.Println(err)
	}
	fmt.Fprint(res, string(mJSON))
}

func jsonResponse(res http.ResponseWriter, status int, result interface{}) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.WriteHeader(status)
	payload, err := json.Marshal(result)
	if err != nil {
		log.Println(err)
	}
	fmt.Fprint(res, string(payload))
}
