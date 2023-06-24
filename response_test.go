package goatquery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_AllPropertiesAreReturned(t *testing.T) {
	query := Query{}

	data := []User{
		{
			Base:      Base{Id: 1},
			Firstname: "John",
			Lastname:  "Doe",
			Email:     "John.Doe@email.com",
		},
	}

	res := BuildPagedResponse(data, query, nil)

	assert.Contains(t, res.Value[0], "id")
}
