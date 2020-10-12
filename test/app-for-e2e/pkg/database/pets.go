package database

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/awslabs/amazon-ec2-instance-qualifier/ec2-instance-qualifier-app/pkg/crypto"
)

const (
	tableName     = "Pets"
	charsetString = "abcdefghijklmnopqrstuvwxyz0123456789"
)

var (
	petCache = make(map[string]Pet)
	charset  = []rune(charsetString)
)

// Pet is the code representation of a pet entry in the database
type Pet struct {
	PetId  string `json:"PetId"`
	Name   string `json:"Name"`
	Breed  string `json:"Breed"`
	Status string `json:"Status"`
}

// GetPetByID looks up and returns a pet by its petId
func GetPetByID(petId string) (Pet, error) {
	if cachedPet, ok := petCache[petId]; ok {
		return cachedPet, nil
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-2"),
	}))
	svc := dynamodb.New(sess)

	petResult := Pet{}
	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"PetId": {
				S: aws.String(petId),
			},
		},
	})
	if err != nil {
		fmt.Println(err)
		return petResult, err
	}
	if result.Item == nil {
		msg := "Could not find pet with Id " + petId
		return petResult, errors.New(msg)
	}

	err = dynamodbattribute.UnmarshalMap(result.Item, &petResult)
	if err != nil {
		msg := "Failed to unmarshal Pet " + petId + " " + err.Error()
		return petResult, errors.New(msg)
	}

	// decrypt
	nameBytes, err := hex.DecodeString(petResult.Name)
	if err != nil {
		return petResult, err
	}
	plainName, err := crypto.Decrypt(nameBytes, crypto.SecureCryptoKey)
	if err != nil {
		return petResult, err
	}

	petResult.Name = string(plainName)
	fmt.Printf("Retrieved %v\n", petResult)
	petCache[petId] = petResult
	return petResult, nil
}

// GetPetCount returns the total number of pets in the table
func GetPetCount() (int64, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-2"),
	}))
	svc := dynamodb.New(sess)
	input := &dynamodb.ScanInput{
		Select:    aws.String(dynamodb.SelectCount),
		TableName: aws.String(tableName),
	}
	output, err := svc.Scan(input)
	if err != nil {
		return 0, err
	}
	fmt.Printf("Pets total table count: %v\n", *output.Count)
	return *output.Count, nil
}

// AddPet adds a pet to the table after encrypting its name
func AddPet(pet Pet) (string, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-2"),
	}))
	svc := dynamodb.New(sess)

	// encrypt dog's name for privacy
	cipherName, err := crypto.Encrypt([]byte(pet.Name), crypto.SecureCryptoKey)
	if err != nil {
		fmt.Printf("There was an error with encryption: %s\n", err.Error())
		return "", err
	}
	pet.Name = hex.EncodeToString(cipherName)
	pet.PetId = createPetId()

	attrValue, err := dynamodbattribute.MarshalMap(pet)
	if err != nil {
		fmt.Println("error marshalling map: ", err.Error())
		return "", err
	}

	input := &dynamodb.PutItemInput{
		Item:      attrValue,
		TableName: aws.String(tableName),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		fmt.Println("error calling PutItem: ", err.Error())
		return "", err
	}

	fmt.Printf("Added %v\n", pet)
	return pet.PetId, nil
}

// DeletePet removes a pet from the table by its petId
func DeletePet(petId string) error {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-2"),
	}))
	svc := dynamodb.New(sess)

	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"PetId": {
				S: aws.String(petId),
			},
		},
		TableName: aws.String(tableName),
	}

	_, err := svc.DeleteItem(input)
	if err != nil {
		fmt.Println("error calling DeleteItem: ", err.Error())
		return err
	}

	fmt.Printf("Deleted %v\n", petId)
	return nil
}

// Helpers

func createPetId() string {
	res := make([]rune, 10)
	for i := range res {
		res[i] = charset[rand.Intn(len(charset))]
	}
	return string(res)
}
