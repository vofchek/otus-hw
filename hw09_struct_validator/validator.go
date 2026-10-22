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

	validatorRe = regexp.MustCompile(`^(?<Name>[a-zA-Z_]+):(?<Params>.*)$`)
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	s := strings.Builder{}
	for _, v := range v {
		_, err := fmt.Fprintf(&s, "{%s: %+v}; ", v.Field, v.Err)
		if err != nil {
			panic(fmt.Errorf("cant assemble error text %w", err))
		}
	}

	return s.String()
}

func Validate(v interface{}) error {
	if v == nil {
		return nil
	}

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

		if strings.TrimSpace(tagValue) == "" {
			return fmt.Errorf("field %s, error: %w", field.Name, errors.Join(ErrParseError, ErrEmptyValidatorTag))
		}

		fieldErrors := make([]error, 0, 2)

		fieldValue := rV.FieldByName(field.Name)

		for _, vS := range strings.Split(tagValue, validatorSeparator) {
			validatorName, validatorParams, err := parseValidator(vS)
			if err != nil {
				return fmt.Errorf("field %s, error: %w", field.Name, errors.Join(ErrParseError, err))
			}

			var validationErr error
			switch validatorName {
			case "len":
				validationErr = validateLen(field, fieldValue, validatorParams)
			case "min":
				validationErr = validateMin(field, fieldValue, validatorParams)
			case "max":
				validationErr = validateMax(field, fieldValue, validatorParams)
			case "regexp":
				validationErr = validateRegex(field, fieldValue, validatorParams)
			case "in":
				validationErr = validateIn(field, fieldValue, validatorParams)
			default:
				err := errors.Join(ErrParseError, ErrUnknownValidator)
				return fmt.Errorf("field %s, validatorName: %s, error: %w", field.Name, validatorName, err)
			}

			if validationErr != nil && !errors.Is(validationErr, ErrValidationError) {
				return fmt.Errorf("field: %s, err: %w", field.Name, err)
			} else if errors.Is(validationErr, ErrValidationError) {
				fieldErrors = append(fieldErrors, validationErr)
			}
		}

		if len(fieldErrors) != 0 {
			e := ValidationError{field.Name, errors.Join(fieldErrors...)}
			validationErrors = append(validationErrors, e)
		}
	}

	if len(validationErrors) != 0 {
		return fmt.Errorf("struct is not valid, errors: %w", validationErrors)
	}

	return nil
}

func parseValidator(tag string) (string, string, error) {
	matches := validatorRe.FindStringSubmatch(tag)
	if len(matches) == 0 {
		return "", "", fmt.Errorf("cant find validator string in %s, err: %w", tag, ErrBadValidatorTag)
	}

	validatorIdx := validatorRe.SubexpIndex("Name")
	if validatorIdx == -1 {
		return "", "", ErrBadValidatorTag
	}

	paramsIdx := validatorRe.SubexpIndex("Params")
	if paramsIdx == -1 {
		return "", "", ErrBadValidatorTag
	}

	return matches[validatorIdx], matches[paramsIdx], nil
}

func validateLen(field reflect.StructField, value reflect.Value, validatorParams string) error {
	pMax, err := strconv.Atoi(validatorParams)
	if err != nil {
		return errors.Join(ErrParseError, ErrBadValidatorTag)
	}

	correctType := field.Type.Kind() == reflect.String ||
		field.Type.Kind() == reflect.Slice ||
		field.Type.Kind() == reflect.Array

	if !correctType {
		return errors.Join(ErrParseError, ErrWrongValueType)
	}

	if value.Len() != pMax {
		return errors.Join(ErrValidationError, fmt.Errorf("%w, max=%+v", ErrInvalidLength, pMax))
	}

	return nil
}

func validateMin(field reflect.StructField, value reflect.Value, validatorParams string) error {
	pMin, err := strconv.Atoi(validatorParams)
	if err != nil {
		return errors.Join(ErrParseError, ErrBadValidatorTag)
	}

	correctType := field.Type.Kind() == reflect.Int

	if !correctType {
		return errors.Join(ErrParseError, ErrWrongValueType)
	}

	if pMin > int(value.Int()) {
		return errors.Join(ErrValidationError, fmt.Errorf("%w, min=%d", ErrValueLessThanMin, pMin))
	}

	return nil
}

func validateMax(field reflect.StructField, value reflect.Value, validatorParams string) error {
	pMax, err := strconv.Atoi(validatorParams)
	if err != nil {
		return errors.Join(ErrParseError, ErrBadValidatorTag)
	}

	correctType := field.Type.Kind() == reflect.Int

	if !correctType {
		return errors.Join(ErrParseError, ErrWrongValueType)
	}

	if pMax < int(value.Int()) {
		return errors.Join(ErrValidationError, fmt.Errorf("%w, max=%d", ErrValueMoreThanMax, pMax))
	}

	return nil
}

func validateRegex(field reflect.StructField, value reflect.Value, validatorParams string) error {
	re, err := regexp.Compile(validatorParams)
	if err != nil {
		return errors.Join(ErrParseError, ErrBadValidatorTag)
	}

	correctType := field.Type.Kind() == reflect.String

	if !correctType {
		return errors.Join(ErrParseError, ErrWrongValueType)
	}

	if !re.MatchString(value.String()) {
		err := fmt.Errorf("%w, pattern=%s", ErrNotMatchingPattern, validatorParams)
		return errors.Join(ErrValidationError, err)
	}

	return nil
}

func validateIn(field reflect.StructField, value reflect.Value, validatorParams string) error {
	inRaw := strings.Split(validatorParams, inParamsSeparator)

	correctType := field.Type.Kind() == reflect.Int ||
		field.Type.Kind() == reflect.String

	if !correctType {
		return errors.Join(ErrParseError, ErrWrongValueType)
	}

	if field.Type.Kind() == reflect.Int {
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

	if field.Type.Kind() == reflect.String {
		if slices.Index(inRaw, value.String()) == -1 {
			return errors.Join(ErrValidationError, fmt.Errorf("%w, in=%+v", ErrNotIn, inRaw))
		}
	}

	return nil
}
