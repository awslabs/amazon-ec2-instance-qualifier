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
	"archive/tar"
	"bufio"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

const (
	randStringLen                   = 15
	charsetString                   = "abcdefghijklmnopqrstuvwxyz0123456789"
	boolPromptAnswerYes             = "y"
	boolPromptAnswerNo              = "N"
	boolPromptInvalidAnswerPrompt   = "Invalid input. Please type"
	optionPromptInvalidAnswerPrompt = "Invalid option. Please choose again:"
)

var charset = []rune(charsetString)

// GetRandomString returns a string of length 15 consisting of lower-case characters and numbers.
func GetRandomString() string {
	res := make([]rune, randStringLen)
	for i := range res {
		res[i] = charset[rand.Intn(len(charset))]
	}
	return string(res)
}

// Compress creates a gzipped archive (.tar.gz) for a folder.
func Compress(filename string, compressedName string) error {
	archiveName := filename + ".tar"
	defer os.Remove(archiveName)
	if err := tarFolder(filename, archiveName); err != nil {
		return err
	}
	if err := gzipFile(archiveName, compressedName); err != nil {
		return err
	}

	log.Printf("%s successfully compressed\n", filepath.Base(filename))

	return nil
}

// BoolPrompt generates a bool prompt in the output stream. It will only return the answer when getting
// "y" or "N", otherwise each time an invalid input is given, the prompt will be outputted again.
func BoolPrompt(prompt string, inputStream *os.File, outputStream *os.File) (bool, error) {
	reader := bufio.NewReader(inputStream)
	validAnswer := boolPromptAnswerYes + "/" + boolPromptAnswerNo
	fmt.Fprintln(outputStream, prompt, validAnswer)

	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			return false, err
		}
		text = strings.TrimSuffix(text, "\n")
		if text == boolPromptAnswerYes {
			return true, nil
		} else if text == boolPromptAnswerNo {
			return false, nil
		} else {
			fmt.Fprintln(outputStream, boolPromptInvalidAnswerPrompt, validAnswer)
		}
	}
}

// OptionPrompt generates an option prompt in the output stream. It will only return the choice when getting
// a valid answer (an integer between [0,optionNum-1]), otherwise each time an invalid input is given, the prompt
// will be outputted again.
func OptionPrompt(prompt string, optionNum int, inputStream *os.File, outputStream *os.File) (int, error) {
	reader := bufio.NewReader(inputStream)
	fmt.Fprintln(outputStream, prompt)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			return 0, err
		}
		text = strings.TrimSuffix(text, "\n")
		option, err := strconv.Atoi(text)
		if err != nil || option < 0 || option >= optionNum {
			fmt.Fprintln(outputStream, optionPromptInvalidAnswerPrompt)
		} else {
			return option, nil
		}
	}
}

// RenderTable renders the 2D string array to the output stream in table format with the given header.
func RenderTable(data [][]string, header []string, outputStream *os.File) {
	table := tablewriter.NewWriter(outputStream)
	table.SetHeader(header)
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.SetAutoFormatHeaders(false)
	table.SetRowLine(true)
	for _, row := range data {
		table.Append(row)
	}
	table.Render()
}

// MarshalToFile marshals an object to a json string and writes it to a file.
func MarshalToFile(v interface{}, filename string) error {
	jsonData, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filename, jsonData, 0644); err != nil {
		return err
	}

	return nil
}

// DecodeBase64 decodes a base64 string.
func DecodeBase64(s string) (string, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// tarFolder creates an archive file for a folder.
func tarFolder(src string, dest string) error {
	writer, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer writer.Close()
	tarWriter := tar.NewWriter(writer)
	defer tarWriter.Close()

	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(src)
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, src))
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		reader, err := os.Open(path)
		if err != nil {
			return err
		}
		defer reader.Close()

		_, err = io.Copy(tarWriter, reader)
		if err != nil {
			return err
		}

		return nil
	})
}

// gzipFile compresses a file using gzip.
func gzipFile(src string, dest string) error {
	reader, err := os.Open(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	writer, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer writer.Close()
	gzipWriter := gzip.NewWriter(writer)
	defer gzipWriter.Close()

	_, err = io.Copy(gzipWriter, reader)
	if err != nil {
		return err
	}

	return nil
}
