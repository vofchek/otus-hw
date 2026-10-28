package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

const (
	validateTag        = "validate"
	validatorSeparator = "|"
	inParamsSeparator  = ","
)

var (
	ErrNotStruct       = errors.New("value must be struct")
	ErrParseError      = errors.New("parse error")
	ErrValidationError = errors.New("validation error")

	ErrBadValidatorTag   = errors.New("validate tag is not properly described")
	ErrEmptyValidatorTag = errors.New("validate tag is empty")
	ErrUnknownValidator  = errors.New("unknown validator")
	ErrWrongValueType    = errors.New("wrong field value type")

	ErrNotIn              = errors.New("value not in allowed values")
	ErrInvalidLength      = errors.New("invalid length")
	ErrValueLessThanMin   = errors.New("value less than min")
	ErrValueMoreThanMax   = errors.New("value more than max")
	ErrNotMatchingPattern = errors.New("value does not match pattern")

	// зарегестрированные валидаторы.
	validators = map[string]validationFn{
		"len":    validateLen,
		"in":     validateIn,
		"max":    validateMax,
		"min":    validateMin,
		"regexp": validateRegexp,
	}
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	s := strings.Builder{}
	for _, v := range v {
		s.WriteString(fmt.Sprintf("{%s: %+v}; ", v.Field, v.Err))
	}

	return s.String()
}

type validationFn func(value reflect.Value, params string) error

func Validate(v interface{}) error {
	rV := reflect.ValueOf(v)
	if rV.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	validationErrors := make(ValidationErrors, 0)

	for i := 0; i < rV.Type().NumField(); i++ {
		field := rV.Type().Field(i)

		if !field.IsExported() {
			continue
		}

		tagValue, tagExists := field.Tag.Lookup(validateTag)

		if !tagExists {
			continue
		}

		if len(strings.TrimSpace(tagValue)) == 0 {
			return fmt.Errorf("field %s, error: %w", field.Name, errors.Join(ErrParseError, ErrEmptyValidatorTag))
		}

		fieldErrors := make([]error, 0, 2)
		fieldValue := rV.FieldByName(field.Name)

		for _, validatorString := range strings.Split(tagValue, validatorSeparator) {
			validationErr := validateField(field, fieldValue, validatorString)

			// если не ошибка валидации, тогда сразу выходим.
			isValidationErr := errors.Is(validationErr, ErrValidationError)

			if isValidationErr {
				fieldErrors = append(fieldErrors, validationErr)
			}

			if validationErr != nil && !isValidationErr {
				return fmt.Errorf("field: %s, err: %w", field.Name, validationErr)
			}
		}

		if len(fieldErrors) != 0 {
			validationErrors = append(validationErrors,
				ValidationError{field.Name, errors.Join(fieldErrors...)},
			)
		}
	}

	if len(validationErrors) != 0 {
		return fmt.Errorf("struct is not valid, errors: %w", validationErrors)
	}

	return nil
}

// валидирует значение поля структуры.
func validateField(field reflect.StructField, fieldValue reflect.Value, validatorString string) error {
	validationFn, validatorParams, err := parseValidator(validatorString)
	if err != nil {
		return err
	}

	var validationErr error
	// если слайс\массив, тогда валидируем каждое значение отдельно.
	if field.Type.Kind() == reflect.Slice || field.Type.Kind() == reflect.Array {
		for i := 0; i < fieldValue.Len(); i++ {
			valueValidationErr := validationFn(fieldValue.Index(i), validatorParams)
			if valueValidationErr != nil {
				validationErr = errors.Join(validationErr, valueValidationErr)
			}
		}
	} else {
		validationErr = validationFn(fieldValue, validatorParams)
	}

	return validationErr
}

// возвращает название и параметры валидатора.
// ошибка возвращается, если тег заполнен некорректно.
func parseValidator(tag string) (validationFn, string, error) {
	validatorName, params, found := strings.Cut(tag, ":")
	if !found {
		err := fmt.Errorf("validatorName: %s, error: %w", validatorName, ErrParseError)
		return nil, "", err
	}

	validationFn, ok := validators[validatorName]
	if !ok {
		err := fmt.Errorf("validatorName: %s, error: %w", validatorName, errors.Join(ErrParseError, ErrUnknownValidator))
		return nil, "", err
	}

	return validationFn, params, nil
}

// Валидаторы.
func validateLen(value reflect.Value, validatorParams string) error {
	maxLen, err := strconv.Atoi(validatorParams)
	if err != nil {
		return errors.Join(ErrParseError, ErrBadValidatorTag)
	}

	correctType := value.Kind() == reflect.String

	if !correctType {
		return errors.Join(ErrParseError, ErrWrongValueType)
	}

	if len(value.String()) != maxLen {
		return errors.Join(ErrValidationError, fmt.Errorf("%w, max=%+v", ErrInvalidLength, maxLen))
	}

	return nil
}

func validateMin(value reflect.Value, validatorParams string) error {
	minValue, err := strconv.Atoi(validatorParams)
	if err != nil {
		return errors.Join(ErrParseError, ErrBadValidatorTag)
	}

	correctType := value.Kind() == reflect.Int

	if !correctType {
		return errors.Join(ErrParseError, ErrWrongValueType)
	}

	if minValue > int(value.Int()) {
		return errors.Join(ErrValidationError, fmt.Errorf("%w, min=%d", ErrValueLessThanMin, minValue))
	}

	return nil
}

func validateMax(value reflect.Value, validatorParams string) error {
	maxValue, err := strconv.Atoi(validatorParams)
	if err != nil {
		return errors.Join(ErrParseError, ErrBadValidatorTag)
	}

	correctType := value.Kind() == reflect.Int

	if !correctType {
		return errors.Join(ErrParseError, ErrWrongValueType)
	}

	if maxValue < int(value.Int()) {
		return errors.Join(ErrValidationError, fmt.Errorf("%w, max=%d", ErrValueMoreThanMax, maxValue))
	}

	return nil
}

func validateRegexp(value reflect.Value, validatorParams string) error {
	re, err := regexp.Compile(validatorParams)
	if err != nil {
		return errors.Join(ErrParseError, ErrBadValidatorTag)
	}

	correctType := value.Kind() == reflect.String

	if !correctType {
		return errors.Join(ErrParseError, ErrWrongValueType)
	}

	if !re.MatchString(value.String()) {
		err := fmt.Errorf("%w, pattern=%s", ErrNotMatchingPattern, validatorParams)
		return errors.Join(ErrValidationError, err)
	}

	return nil
}

func validateIn(value reflect.Value, validatorParams string) error {
	inRaw := strings.Split(validatorParams, inParamsSeparator)

	correctType := value.Kind() == reflect.Int ||
		value.Kind() == reflect.String

	if !correctType {
		return errors.Join(ErrParseError, ErrWrongValueType)
	}

	if value.Kind() == reflect.Int {
		in := make([]int, 0, len(inRaw))
		// конвертация слайса строк в слайс int
		for _, strValue := range inRaw {
			v, err := strconv.Atoi(strValue)
			if err != nil {
				return errors.Join(ErrParseError, ErrBadValidatorTag)
			}
			in = append(in, v)
		}

		if slices.Index(in, int(value.Int())) == -1 {
			return errors.Join(ErrValidationError, fmt.Errorf("%w, in=%+v", ErrNotIn, in))
		}
	}

	if value.Kind() == reflect.String {
		if slices.Index(inRaw, value.String()) == -1 {
			return errors.Join(ErrValidationError, fmt.Errorf("%w, in=%+v", ErrNotIn, inRaw))
		}
	}

	return nil
}
