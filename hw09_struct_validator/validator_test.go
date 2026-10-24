package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int      `validate:"min:18|max:50"`
		Email  string   `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole `validate:"in:admin,stuff"`
		Phones []string `validate:"len:11"`
		meta   json.RawMessage
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}

	MyTest struct {
		MyField    []string `validate:"len:10|regexp:(?i)^[a-z]+$"`
		SliceField []int    `validate:"in:10,20,30"`
		notPublic  string   `validate:"in:test1,test2"`
	}
)

func TestValidateNotStruct(t *testing.T) {
	t.Run("value is not struct", func(t *testing.T) {
		err := Validate(42)

		require.ErrorIs(t, ErrNotStruct, err)
	})

	t.Run("value is nil", func(t *testing.T) {
		err := Validate(nil)

		require.ErrorIs(t, ErrNotStruct, err)
	})
}

func TestValidate(t *testing.T) {
	tests := []struct {
		in          interface{}
		expectedErr error
	}{
		{
			Response{
				Code: 302,
				Body: "test",
			},
			ValidationErrors{
				ValidationError{
					Field: "Code",
					Err:   ErrNotIn,
				},
			},
		},

		{
			App{
				Version: strings.Repeat("a", 10),
			},
			ValidationErrors{
				ValidationError{
					Field: "Version",
					Err:   ErrInvalidLength,
				},
			},
		},

		{
			User{
				ID:     strings.Repeat("i", 36),
				Name:   "test",
				Age:    25,
				Email:  "email@email.com",
				Role:   UserRole("admin"),
				Phones: []string{strings.Repeat("P", 12)},
				meta:   nil,
			},
			ValidationErrors{
				ValidationError{
					Field: "Phones",
					Err:   ErrInvalidLength,
				},
			},
		},

		{
			User{
				ID:     strings.Repeat("i", 36),
				Name:   "test",
				Age:    10,
				Email:  "email@email.com",
				Role:   UserRole("admin"),
				Phones: []string{strings.Repeat("P", 11)},
				meta:   nil,
			},
			ValidationErrors{
				ValidationError{
					Field: "Age",
					Err:   ErrValueLessThanMin,
				},
			},
		},

		{
			User{
				ID:     strings.Repeat("i", 36),
				Name:   "test",
				Age:    150,
				Email:  "email@email.com",
				Role:   UserRole("admin"),
				Phones: []string{strings.Repeat("P", 11)},
				meta:   nil,
			},
			ValidationErrors{
				ValidationError{
					Field: "Age",
					Err:   ErrValueMoreThanMax,
				},
			},
		},

		{
			User{
				ID:     strings.Repeat("i", 36),
				Name:   "test",
				Age:    25,
				Email:  "this_is_not_an_email",
				Role:   UserRole("admin"),
				Phones: []string{strings.Repeat("P", 11)},
				meta:   nil,
			},
			ValidationErrors{
				ValidationError{
					Field: "Email",
					Err:   ErrNotMatchingPattern,
				},
			},
		},

		{
			User{
				ID:     strings.Repeat("i", 36),
				Name:   "test",
				Age:    25,
				Email:  "email@email.com",
				Role:   UserRole("invalid role"),
				Phones: []string{strings.Repeat("P", 11)},
				meta:   nil,
			},
			ValidationErrors{
				ValidationError{
					Field: "Role",
					Err:   ErrNotIn,
				},
			},
		},

		{
			MyTest{
				MyField:    []string{"1"},
				SliceField: []int{10},
			},
			ValidationErrors{
				ValidationError{
					Field: "MyField",
					Err:   ErrInvalidLength,
				},
			},
		},
		{
			MyTest{
				MyField:    []string{"1"},
				SliceField: []int{10},
			},
			ValidationErrors{
				ValidationError{
					Field: "MyField",
					Err:   ErrNotMatchingPattern,
				},
			},
		},
		{
			MyTest{
				MyField:    []string{"1"},
				SliceField: []int{777},
			},
			ValidationErrors{
				ValidationError{
					Field: "SliceField",
					Err:   ErrNotIn,
				},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()

			// _ = tt
			validationError := Validate(tt.in)
			require.Error(t, validationError, "not validation error")

			var e ValidationErrors
			require.ErrorAs(t, validationError, &e, "%+v", validationError)

			var expected ValidationErrors
			require.ErrorAs(t, tt.expectedErr, &expected, "test expected validation error wrong format")

			errMap := make(map[string]error, 0)
			for _, vErr := range expected {
				errMap[vErr.Field] = vErr.Err
			}

			errExists := false
			for _, err := range e {
				validationErr, inMap := errMap[err.Field]
				if inMap && errors.Is(err.Err, validationErr) {
					errExists = true
					break
				}
			}
			assert.True(t, errExists, "no expected validation error for field %v", tt.expectedErr)
		})
	}
}

func TestValidateAllValid(t *testing.T) {
	tests := []interface{}{
		Response{
			Code: 200,
			Body: "test",
		},
		App{
			Version: strings.Repeat("a", 5),
		},
		User{
			ID:     strings.Repeat("i", 36),
			Name:   "test",
			Age:    25,
			Email:  "email@email.com",
			Role:   UserRole("admin"),
			Phones: []string{strings.Repeat("P", 11)},
			meta:   nil,
		},
		MyTest{
			MyField:    []string{"ThisIsTest"},
			SliceField: []int{10},
			notPublic:  "bad value but without validation",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()

			// _ = tt
			validationError := Validate(tt)
			require.Nil(t, validationError, "some error: %+v", validationError)
		})
	}
}

func TestValidateParsingErorrs(t *testing.T) {
	tests := []struct {
		in interface{}
		e  error
	}{
		{
			in: struct {
				A int `validate:""`
			}{
				A: 10,
			},
			e: ErrEmptyValidatorTag,
		},
		{
			in: struct {
				A int `validate:"UNKNOWN_VALIDATOR:0"`
			}{
				A: -1000,
			},
			e: ErrUnknownValidator,
		},
		{
			in: struct {
				A int `validate:"len:999"`
			}{
				A: 10000,
			},
			e: ErrWrongValueType,
		},
		{
			in: struct {
				A int `validate:"in:A,B,C"`
			}{
				A: 10000,
			},
			e: ErrBadValidatorTag,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()

			// _ = tt
			validationError := Validate(tt.in)
			require.Error(t, validationError, "struct cant be valid")
		})
	}
}
