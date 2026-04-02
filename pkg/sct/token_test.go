package sct

import (
	"errors"
	"testing"
)

const validId = "12345678901234567890"
const invalidId = "1"
const validSecret = "12345678901234567890123456789012345678901234567890"
const invalidSecret = "1"
const validName = "some-name"

func TestValidateHappyPath(t *testing.T) {
	token := &Token[interface{}]{
		Name: validName,
		Id:   validId,
		Secrets: map[string]string{
			"key": validSecret,
		},
	}
	err := token.Validate()
	if err != nil {
		t.Error(err)
		panic(err)
	}
}

func TestValidateMissingName(t *testing.T) {
	token := &Token[interface{}]{
		Id: validId,
		Secrets: map[string]string{
			"key": validSecret,
		},
	}
	err := token.Validate()
	if err == nil {
		err = errors.New("name should be required")
		t.Error(err)
		panic(err)
	}
}

func TestValidateMissingSecrets(t *testing.T) {
	token := &Token[interface{}]{
		Name: validName,
		Id:   validId,
	}
	err := token.Validate()
	if err == nil {
		err = errors.New("secrets should be required")
		t.Error(err)
		panic(err)
	}
}

func TestValidateIdIsTooShort(t *testing.T) {
	token := &Token[interface{}]{
		Name: validName,
		Id:   invalidId,
		Secrets: map[string]string{
			"key": validSecret,
		},
	}
	err := token.Validate()
	if err == nil {
		err = errors.New("id should be too short")
		t.Error(err)
		panic(err)
	}
}

func TestValidateSecretIsTooShort(t *testing.T) {
	token := &Token[interface{}]{
		Name: validName,
		Id:   validId,
		Secrets: map[string]string{
			"key": invalidSecret,
		},
	}
	err := token.Validate()
	if err == nil {
		err = errors.New("secret should be too short")
		t.Error(err)
		panic(err)
	}
}
