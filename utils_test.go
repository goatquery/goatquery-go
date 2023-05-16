package goatquery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_splitStringOnMultipleDelimiters(t *testing.T) {
	filter := "firstname eq 'goat'"

	result := splitString(filter)

	assert.Equal(t, 1, len(result))
	assert.Equal(t, []string{"firstname eq 'goat'"}, result)
}

func Test_splitStringOnMultipleDelimitersAnd(t *testing.T) {
	filter := "firstname eq 'goat' and lastname eq 'query'"

	result := splitString(filter)

	assert.Equal(t, 3, len(result))
	assert.Equal(t, []string{"firstname eq 'goat'", "and", "lastname eq 'query'"}, result)
}

func Test_splitStringOnMultipleDelimiterOr(t *testing.T) {
	filter := "firstname eq 'and' or lastname eq 'and'"

	result := splitString(filter)

	assert.Equal(t, 3, len(result))
	assert.Equal(t, []string{"firstname eq 'and'", "or", "lastname eq 'and'"}, result)
}

func Test_splitStringOnMultipleDelimiterOrWithSpace(t *testing.T) {
	filter := "firstname eq ' and ' or lastname eq ' and or '"

	result := splitString(filter)

	assert.Equal(t, 3, len(result))
	assert.Equal(t, []string{"firstname eq ' and '", "or", "lastname eq ' and or '"}, result)
}
