package goatquery

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_AllPropertiesAreReturned(t *testing.T) {
	query := Query{}

	data := []User{
		{
			Base:      Base{Id: uuid.New()},
			Firstname: "John",
			Lastname:  "Doe",
			Email:     "John.Doe@email.com",
		},
	}

	res := BuildPagedResponse(data, query, nil)

	assert.Contains(t, res.Value[0], "id")
}

func Test_SelectUUIDTypeReturnsCorrectValue(t *testing.T) {
	field := "id"
	query := Query{Select: field}
	uuid := uuid.New()

	data := []User{
		{
			Base:      Base{Id: uuid},
			Firstname: "John",
			Lastname:  "Doe",
			Email:     "John.Doe@email.com",
		},
	}

	res := BuildPagedResponse(data, query, nil)

	assert.Equal(t, uuid, res.Value[0][field])
}

func Test_SelectIntTypeReturnsCorrectValue(t *testing.T) {
	field := "age"
	query := Query{Select: field}
	age := uint(21)

	data := []User{
		{
			Age:       age,
			Base:      Base{Id: uuid.New()},
			Firstname: "John",
			Lastname:  "Doe",
			Email:     "John.Doe@email.com",
		},
	}

	res := BuildPagedResponse(data, query, nil)

	assert.Equal(t, age, res.Value[0][field])
}
