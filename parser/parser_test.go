package parser

import (
	"testing"

	"github.com/goatquery/goatquery-go/ast"
	"github.com/goatquery/goatquery-go/lexer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- OrderBy ---

func Test_ParseOrderBy(t *testing.T) {
	type expected struct {
		segments  []string
		direction ast.OrderByDirection
	}

	tests := []struct {
		name     string
		input    string
		expected []expected
	}{
		{"single descending", "ID desc", []expected{
			{[]string{"ID"}, ast.Descending},
		}},
		{"single ascending", "id asc", []expected{
			{[]string{"id"}, ast.Ascending},
		}},
		{"default ascending", "Name", []expected{
			{[]string{"Name"}, ast.Ascending},
		}},
		{"multiple fields", "id asc, name desc", []expected{
			{[]string{"id"}, ast.Ascending},
			{[]string{"name"}, ast.Descending},
		}},
		{"five fields mixed", "id asc, name desc, age, address asc, postcode desc", []expected{
			{[]string{"id"}, ast.Ascending},
			{[]string{"name"}, ast.Descending},
			{[]string{"age"}, ast.Ascending},
			{[]string{"address"}, ast.Ascending},
			{[]string{"postcode"}, ast.Descending},
		}},
		{"keywords as property names", "address1Line10 asc, asc asc, desc desc", []expected{
			{[]string{"address1Line10"}, ast.Ascending},
			{[]string{"asc"}, ast.Ascending},
			{[]string{"desc"}, ast.Descending},
		}},
		{"empty input", "", []expected{}},
		{"nested property ascending", "company/name asc", []expected{
			{[]string{"company", "name"}, ast.Ascending},
		}},
		{"nested property descending", "company/name desc", []expected{
			{[]string{"company", "name"}, ast.Descending},
		}},
		{"nested and flat mixed", "company/name asc, age desc", []expected{
			{[]string{"company", "name"}, ast.Ascending},
			{[]string{"age"}, ast.Descending},
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := lexer.NewLexer(test.input)
			p := NewParser(l)

			statements, err := p.ParseOrderBy()
			require.NoError(t, err)
			require.Len(t, statements, len(test.expected))

			for i, expected := range test.expected {
				assert.Equal(t, expected.segments, statements[i].Segments, "index: %d", i)
				assert.Equal(t, expected.direction, statements[i].Direction, "index: %d", i)
			}
		})
	}
}

func Test_ParseOrderByErrors(t *testing.T) {
	inputs := []string{
		"123 asc",
		"'name' asc",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			l := lexer.NewLexer(input)
			p := NewParser(l)

			_, err := p.ParseOrderBy()
			assert.Error(t, err)
		})
	}
}

// --- Filter: simple comparisons ---

func Test_ParseFilter(t *testing.T) {
	tests := []struct {
		input            string
		expectedLeft     string
		expectedOperator string
		expectedRight    string
	}{
		{"Name eq 'John'", "Name", "eq", "John"},
		{"Firstname eq 'Jane'", "Firstname", "eq", "Jane"},
		{"Age eq 21", "Age", "eq", "21"},
		{"Age ne 10", "Age", "ne", "10"},
		{"Name contains 'John'", "Name", "contains", "John"},
		{"Id eq e4c7772b-8947-4e46-98ed-644b417d2a08", "Id", "eq", "e4c7772b-8947-4e46-98ed-644b417d2a08"},
		{"Id eq 3.14159265359", "Id", "eq", "3.14159265359"},
		{"Age lt 99", "Age", "lt", "99"},
		{"Age lte 99", "Age", "lte", "99"},
		{"Age gt 99", "Age", "gt", "99"},
		{"Age gte 99", "Age", "gte", "99"},
		{"dateOfBirth eq 2000-01-01", "dateOfBirth", "eq", "2000-01-01"},
		{"dateOfBirth lt 2000-01-01", "dateOfBirth", "lt", "2000-01-01"},
		{"dateOfBirth lte 2000-01-01", "dateOfBirth", "lte", "2000-01-01"},
		{"dateOfBirth gt 2000-01-01", "dateOfBirth", "gt", "2000-01-01"},
		{"dateOfBirth gte 2000-01-01", "dateOfBirth", "gte", "2000-01-01"},
		{"dateOfBirth eq 2023-01-30T09:29:55.1750906Z", "dateOfBirth", "eq", "2023-01-30T09:29:55.1750906Z"},
		{"age eq 100", "age", "eq", "100"},
		{"age eq 99999999999", "age", "eq", "99999999999"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			l := lexer.NewLexer(test.input)
			p := NewParser(l)

			statement, err := p.ParseFilter()
			require.NoError(t, err)
			require.NotNil(t, statement)

			infix, ok := statement.(*ast.InfixExpression)
			require.True(t, ok, "expected InfixExpression")

			assert.Equal(t, test.expectedLeft, infix.Left.TokenLiteral())
			assert.Equal(t, test.expectedOperator, infix.Operator)
			assert.Equal(t, test.expectedRight, infix.Right.TokenLiteral())
		})
	}
}

func Test_ParseFilterEscapedStrings(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedRight string
	}{
		{"escaped single quote", "Name eq 'O\\'Brien'", "O'Brien"},
		{"escaped backslash", "Name eq 'back\\\\slash'", "back\\slash"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := lexer.NewLexer(test.input)
			p := NewParser(l)

			statement, err := p.ParseFilter()
			require.NoError(t, err)

			infix, ok := statement.(*ast.InfixExpression)
			require.True(t, ok)

			str, ok := infix.Right.(*ast.StringLiteral)
			require.True(t, ok)

			assert.Equal(t, test.expectedRight, str.Value)
		})
	}
}

// --- Filter: null and boolean literals ---

func Test_ParseFilterNullLiteral(t *testing.T) {
	tests := []struct {
		input    string
		operator string
	}{
		{"balance eq null", "eq"},
		{"balance ne null", "ne"},
		{"name eq NULL", "eq"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			l := lexer.NewLexer(test.input)
			p := NewParser(l)

			statement, err := p.ParseFilter()
			require.NoError(t, err)

			infix, ok := statement.(*ast.InfixExpression)
			require.True(t, ok)

			assert.Equal(t, test.operator, infix.Operator)

			_, ok = infix.Right.(*ast.NullLiteral)
			assert.True(t, ok, "expected NullLiteral")
		})
	}
}

func Test_ParseFilterBooleanLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"isActive eq true", true},
		{"isActive eq false", false},
		{"isActive ne true", true},
		{"isActive ne false", false},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			l := lexer.NewLexer(test.input)
			p := NewParser(l)

			statement, err := p.ParseFilter()
			require.NoError(t, err)

			infix, ok := statement.(*ast.InfixExpression)
			require.True(t, ok)

			boolLit, ok := infix.Right.(*ast.BooleanLiteral)
			require.True(t, ok, "expected BooleanLiteral")

			assert.Equal(t, test.expected, boolLit.Value)
		})
	}
}

// --- Filter: literal type assertions ---

func Test_ParseFilterLiteralTypes(t *testing.T) {
	t.Run("integer", func(t *testing.T) {
		l := lexer.NewLexer("age eq 21")
		p := NewParser(l)
		stmt, err := p.ParseFilter()
		require.NoError(t, err)
		infix := stmt.(*ast.InfixExpression)
		lit, ok := infix.Right.(*ast.IntegerLiteral)
		require.True(t, ok)
		assert.Equal(t, int64(21), lit.Value)
	})

	t.Run("large integer", func(t *testing.T) {
		l := lexer.NewLexer("age eq 100")
		p := NewParser(l)
		stmt, err := p.ParseFilter()
		require.NoError(t, err)
		infix := stmt.(*ast.InfixExpression)
		lit, ok := infix.Right.(*ast.IntegerLiteral)
		require.True(t, ok)
		assert.Equal(t, int64(100), lit.Value)
	})

	t.Run("double literal", func(t *testing.T) {
		l := lexer.NewLexer("val eq 3.14")
		p := NewParser(l)
		stmt, err := p.ParseFilter()
		require.NoError(t, err)
		infix := stmt.(*ast.InfixExpression)
		lit, ok := infix.Right.(*ast.FloatLiteral)
		require.True(t, ok)
		assert.InDelta(t, 3.14, lit.Value, 0.001)
	})

	t.Run("guid", func(t *testing.T) {
		l := lexer.NewLexer("id eq e4c7772b-8947-4e46-98ed-644b417d2a08")
		p := NewParser(l)
		stmt, err := p.ParseFilter()
		require.NoError(t, err)
		infix := stmt.(*ast.InfixExpression)
		lit, ok := infix.Right.(*ast.UUIDLiteral)
		require.True(t, ok)
		assert.Equal(t, "e4c7772b-8947-4e46-98ed-644b417d2a08", lit.Value.String())
	})

	t.Run("date", func(t *testing.T) {
		l := lexer.NewLexer("dob eq 2000-01-01")
		p := NewParser(l)
		stmt, err := p.ParseFilter()
		require.NoError(t, err)
		infix := stmt.(*ast.InfixExpression)
		_, ok := infix.Right.(*ast.DateLiteral)
		assert.True(t, ok)
	})

	t.Run("datetime", func(t *testing.T) {
		l := lexer.NewLexer("dob eq 2023-01-30T09:29:55.1750906Z")
		p := NewParser(l)
		stmt, err := p.ParseFilter()
		require.NoError(t, err)
		infix := stmt.(*ast.InfixExpression)
		_, ok := infix.Right.(*ast.DateTimeLiteral)
		assert.True(t, ok)
	})

	t.Run("negative integer", func(t *testing.T) {
		l := lexer.NewLexer("age gt -5")
		p := NewParser(l)
		stmt, err := p.ParseFilter()
		require.NoError(t, err)
		infix := stmt.(*ast.InfixExpression)
		lit, ok := infix.Right.(*ast.IntegerLiteral)
		require.True(t, ok)
		assert.Equal(t, int64(-5), lit.Value)
	})

	t.Run("negative double", func(t *testing.T) {
		l := lexer.NewLexer("age eq -0.5")
		p := NewParser(l)
		stmt, err := p.ParseFilter()
		require.NoError(t, err)
		infix := stmt.(*ast.InfixExpression)
		lit, ok := infix.Right.(*ast.FloatLiteral)
		require.True(t, ok)
		assert.InDelta(t, -0.5, lit.Value, 0.001)
	})
}

// --- Filter: logical operators ---

func Test_ParseFilterLogicalOperators(t *testing.T) {
	t.Run("and", func(t *testing.T) {
		l := lexer.NewLexer("Name eq 'John' and Age eq 10")
		p := NewParser(l)

		statement, err := p.ParseFilter()
		require.NoError(t, err)

		outer, ok := statement.(*ast.InfixExpression)
		require.True(t, ok)
		assert.Equal(t, "and", outer.Operator)

		left, ok := outer.Left.(*ast.InfixExpression)
		require.True(t, ok)
		assert.Equal(t, "Name", left.Left.TokenLiteral())
		assert.Equal(t, "eq", left.Operator)
		assert.Equal(t, "John", left.Right.TokenLiteral())

		right, ok := outer.Right.(*ast.InfixExpression)
		require.True(t, ok)
		assert.Equal(t, "Age", right.Left.TokenLiteral())
		assert.Equal(t, "eq", right.Operator)
		assert.Equal(t, "10", right.Right.TokenLiteral())
	})

	t.Run("or", func(t *testing.T) {
		l := lexer.NewLexer("Name eq 'John' or Age eq 10")
		p := NewParser(l)

		statement, err := p.ParseFilter()
		require.NoError(t, err)

		outer, ok := statement.(*ast.InfixExpression)
		require.True(t, ok)
		assert.Equal(t, "or", outer.Operator)

		left, ok := outer.Left.(*ast.InfixExpression)
		require.True(t, ok)
		assert.Equal(t, "Name", left.Left.TokenLiteral())
		assert.Equal(t, "eq", left.Operator)
		assert.Equal(t, "John", left.Right.TokenLiteral())

		right, ok := outer.Right.(*ast.InfixExpression)
		require.True(t, ok)
		assert.Equal(t, "Age", right.Left.TokenLiteral())
		assert.Equal(t, "eq", right.Operator)
		assert.Equal(t, "10", right.Right.TokenLiteral())
	})

	t.Run("and binds tighter than or", func(t *testing.T) {
		l := lexer.NewLexer("Name eq 'John' and Age eq 10 or Id eq 10")
		p := NewParser(l)

		statement, err := p.ParseFilter()
		require.NoError(t, err)

		outer, ok := statement.(*ast.InfixExpression)
		require.True(t, ok)
		assert.Equal(t, "or", outer.Operator)

		left, ok := outer.Left.(*ast.InfixExpression)
		require.True(t, ok)
		assert.Equal(t, "and", left.Operator)

		innerLeft, ok := left.Left.(*ast.InfixExpression)
		require.True(t, ok)
		assert.Equal(t, "Name", innerLeft.Left.TokenLiteral())
		assert.Equal(t, "eq", innerLeft.Operator)
		assert.Equal(t, "John", innerLeft.Right.TokenLiteral())

		innerRight, ok := left.Right.(*ast.InfixExpression)
		require.True(t, ok)
		assert.Equal(t, "Age", innerRight.Left.TokenLiteral())
		assert.Equal(t, "eq", innerRight.Operator)
		assert.Equal(t, "10", innerRight.Right.TokenLiteral())

		right, ok := outer.Right.(*ast.InfixExpression)
		require.True(t, ok)
		assert.Equal(t, "Id", right.Left.TokenLiteral())
		assert.Equal(t, "eq", right.Operator)
		assert.Equal(t, "10", right.Right.TokenLiteral())
	})
}

// --- Filter: grouped expressions ---

func Test_ParseFilterGrouped(t *testing.T) {
	t.Run("parentheses on right", func(t *testing.T) {
		l := lexer.NewLexer("Firstname eq 'User01' and (Age eq 25 or Age eq 30)")
		p := NewParser(l)

		statement, err := p.ParseFilter()
		require.NoError(t, err)

		outer, ok := statement.(*ast.InfixExpression)
		require.True(t, ok)
		assert.Equal(t, "and", outer.Operator)

		left, ok := outer.Left.(*ast.InfixExpression)
		require.True(t, ok)
		assert.Equal(t, "Firstname", left.Left.TokenLiteral())
		assert.Equal(t, "eq", left.Operator)
		assert.Equal(t, "User01", left.Right.TokenLiteral())

		right, ok := outer.Right.(*ast.InfixExpression)
		require.True(t, ok)
		assert.Equal(t, "or", right.Operator)

		rightLeft, ok := right.Left.(*ast.InfixExpression)
		require.True(t, ok)
		assert.Equal(t, "Age", rightLeft.Left.TokenLiteral())
		assert.Equal(t, "eq", rightLeft.Operator)
		assert.Equal(t, "25", rightLeft.Right.TokenLiteral())

		rightRight, ok := right.Right.(*ast.InfixExpression)
		require.True(t, ok)
		assert.Equal(t, "Age", rightRight.Left.TokenLiteral())
		assert.Equal(t, "eq", rightRight.Operator)
		assert.Equal(t, "30", rightRight.Right.TokenLiteral())
	})

	t.Run("parentheses on left", func(t *testing.T) {
		l := lexer.NewLexer("(Firstname eq 'User01' and Age eq 25) or Age eq 30")
		p := NewParser(l)

		statement, err := p.ParseFilter()
		require.NoError(t, err)

		outer, ok := statement.(*ast.InfixExpression)
		require.True(t, ok)
		assert.Equal(t, "or", outer.Operator)

		left, ok := outer.Left.(*ast.InfixExpression)
		require.True(t, ok)
		assert.Equal(t, "and", left.Operator)
	})
}

// --- Filter: nested properties ---

func Test_ParseFilterNestedProperty(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedSegments []string
		expectedOperator string
		expectedRight    string
	}{
		{"single level", "manager/firstName eq 'John'", []string{"manager", "firstName"}, "eq", "John"},
		{"two levels deep", "manager/manager/firstName eq 'John'", []string{"manager", "manager", "firstName"}, "eq", "John"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := lexer.NewLexer(test.input)
			p := NewParser(l)

			statement, err := p.ParseFilter()
			require.NoError(t, err)

			infix, ok := statement.(*ast.InfixExpression)
			require.True(t, ok)

			propPath, ok := infix.Left.(*ast.PropertyPath)
			require.True(t, ok, "expected PropertyPath")

			assert.Equal(t, test.expectedSegments, propPath.Segments)
			assert.Equal(t, test.expectedOperator, infix.Operator)
			assert.Equal(t, test.expectedRight, infix.Right.TokenLiteral())
		})
	}
}

// --- Filter: lambda expressions ---

func Test_ParseFilterLambdaAny(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		property  string
		function  string
		parameter string
	}{
		{"tags", "tags/any(t: t eq 'tag 2')", "tags", "any", "t"},
		{"categories", "categories/any(c: c eq 'electronics')", "categories", "any", "c"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := lexer.NewLexer(test.input)
			p := NewParser(l)

			statement, err := p.ParseFilter()
			require.NoError(t, err)

			lambda, ok := statement.(*ast.LambdaExpression)
			require.True(t, ok, "expected LambdaExpression")

			assert.Equal(t, test.property, lambda.Property.TokenLiteral())
			assert.Equal(t, test.function, lambda.Function)
			assert.Equal(t, test.parameter, lambda.Parameter)

			body, ok := lambda.Body.(*ast.InfixExpression)
			require.True(t, ok, "expected InfixExpression body")
			assert.Equal(t, "eq", body.Operator)
		})
	}
}

func Test_ParseFilterLambdaAll(t *testing.T) {
	l := lexer.NewLexer("items/all(i: i ne null)")
	p := NewParser(l)

	statement, err := p.ParseFilter()
	require.NoError(t, err)

	lambda, ok := statement.(*ast.LambdaExpression)
	require.True(t, ok)

	assert.Equal(t, "items", lambda.Property.TokenLiteral())
	assert.Equal(t, "all", lambda.Function)
	assert.Equal(t, "i", lambda.Parameter)

	body, ok := lambda.Body.(*ast.InfixExpression)
	require.True(t, ok)
	assert.Equal(t, "ne", body.Operator)

	_, ok = body.Right.(*ast.NullLiteral)
	assert.True(t, ok)
}

func Test_ParseFilterLambdaNestedPropertyInBody(t *testing.T) {
	l := lexer.NewLexer("addresses/any(address: address/city eq 'New York')")
	p := NewParser(l)

	statement, err := p.ParseFilter()
	require.NoError(t, err)

	lambda, ok := statement.(*ast.LambdaExpression)
	require.True(t, ok)
	assert.Equal(t, "addresses", lambda.Property.TokenLiteral())
	assert.Equal(t, "any", lambda.Function)
	assert.Equal(t, "address", lambda.Parameter)

	body, ok := lambda.Body.(*ast.InfixExpression)
	require.True(t, ok)

	propPath, ok := body.Left.(*ast.PropertyPath)
	require.True(t, ok)
	assert.Equal(t, []string{"address", "city"}, propPath.Segments)
	assert.Equal(t, "New York", body.Right.TokenLiteral())
}

func Test_ParseFilterLambdaWithLogicalOperators(t *testing.T) {
	t.Run("lambda combined with and", func(t *testing.T) {
		l := lexer.NewLexer("name eq 'John' and tags/any(t: t eq 'important')")
		p := NewParser(l)

		statement, err := p.ParseFilter()
		require.NoError(t, err)

		outer, ok := statement.(*ast.InfixExpression)
		require.True(t, ok)
		assert.Equal(t, "and", outer.Operator)

		left, ok := outer.Left.(*ast.InfixExpression)
		require.True(t, ok)
		assert.Equal(t, "name", left.Left.TokenLiteral())

		_, ok = outer.Right.(*ast.LambdaExpression)
		assert.True(t, ok)
	})

	t.Run("lambda on left with and", func(t *testing.T) {
		l := lexer.NewLexer("tags/any(t: t contains 'work') and status eq 'active'")
		p := NewParser(l)

		statement, err := p.ParseFilter()
		require.NoError(t, err)

		outer, ok := statement.(*ast.InfixExpression)
		require.True(t, ok)
		assert.Equal(t, "and", outer.Operator)

		_, ok = outer.Left.(*ast.LambdaExpression)
		assert.True(t, ok)

		right, ok := outer.Right.(*ast.InfixExpression)
		require.True(t, ok)
		assert.Equal(t, "status", right.Left.TokenLiteral())
	})
}

func Test_ParseFilterLambdaComplexBody(t *testing.T) {
	t.Run("and inside lambda body", func(t *testing.T) {
		l := lexer.NewLexer("tags/any(t: t eq 'tag1' and t ne 'tag2')")
		p := NewParser(l)

		statement, err := p.ParseFilter()
		require.NoError(t, err)

		lambda, ok := statement.(*ast.LambdaExpression)
		require.True(t, ok)

		body, ok := lambda.Body.(*ast.InfixExpression)
		require.True(t, ok)
		assert.Equal(t, "and", body.Operator)
	})

	t.Run("or and nested properties inside lambda body", func(t *testing.T) {
		l := lexer.NewLexer("items/all(i: i/price gt 100 or i/discount lt 0.1)")
		p := NewParser(l)

		statement, err := p.ParseFilter()
		require.NoError(t, err)

		lambda, ok := statement.(*ast.LambdaExpression)
		require.True(t, ok)
		assert.Equal(t, "all", lambda.Function)

		body, ok := lambda.Body.(*ast.InfixExpression)
		require.True(t, ok)
		assert.Equal(t, "or", body.Operator)
	})

	t.Run("nested lambda (any inside any)", func(t *testing.T) {
		l := lexer.NewLexer("orders/any(o: o/items/any(i: i/price gt 1000))")
		p := NewParser(l)

		statement, err := p.ParseFilter()
		require.NoError(t, err)

		outerLambda, ok := statement.(*ast.LambdaExpression)
		require.True(t, ok)
		assert.Equal(t, "any", outerLambda.Function)
		assert.Equal(t, "o", outerLambda.Parameter)

		innerLambda, ok := outerLambda.Body.(*ast.LambdaExpression)
		require.True(t, ok)
		assert.Equal(t, "any", innerLambda.Function)
		assert.Equal(t, "i", innerLambda.Parameter)

		innerProp, ok := innerLambda.Property.(*ast.PropertyPath)
		require.True(t, ok)
		assert.Equal(t, []string{"o", "items"}, innerProp.Segments)

		body, ok := innerLambda.Body.(*ast.InfixExpression)
		require.True(t, ok)
		assert.Equal(t, "gt", body.Operator)
	})
}

// --- Filter: error cases ---

func Test_ParseFilterErrors(t *testing.T) {
	t.Run("syntax", func(t *testing.T) {
		inputs := []string{
			"Name",
			"",
			"eq nee",
			"name nee 10",
			"id contains 10",
			"id contaiins '10'",
			"id eq       John'",
			"name contains null",
			"age lt null",
			"age gt null",
			"age lte null",
			"age gte null",
			"age gt --5",
			"age gt 'hello'",
			"age lt true",
			"age gte e4c7772b-a529-4ed0-b392-1178e3edfa24",
			"name lte false",
		}

		for _, input := range inputs {
			t.Run(input, func(t *testing.T) {
				l := lexer.NewLexer(input)
				p := NewParser(l)

				_, err := p.ParseFilter()
				assert.Error(t, err)
			})
		}
	})

	t.Run("lambda", func(t *testing.T) {
		inputs := []string{
			"tags/any(t: t eq)",
			"tags/any(t t eq 'test')",
			"tags/any( : t eq 'test')",
			"tags/any(t:)",
			"tags/any",
			"tags/any()",
			"tags/invalid(t: t eq 'test')",
		}

		for _, input := range inputs {
			t.Run(input, func(t *testing.T) {
				l := lexer.NewLexer(input)
				p := NewParser(l)

				_, err := p.ParseFilter()
				assert.Error(t, err)
			})
		}
	})
}
