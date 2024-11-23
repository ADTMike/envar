package envar

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// convertToType converts a string value to the specified field type.
func convert(value string, field reflect.Value) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			durationValue, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("failed to convert %s to time.Duration: %v", value, err)
			}
			field.Set(reflect.ValueOf(durationValue))
			break
		}
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to convert %s to int: %v", value, err)
		}
		field.SetInt(intValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:

		uintValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to convert %s to uint: %v", value, err)
		}
		field.SetUint(uintValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("failed to convert %s to bool: %v", value, err)
		}
		field.SetBool(boolValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("failed to convert %s to float: %v", value, err)
		}
		field.SetFloat(floatValue)
	case reflect.Complex64, reflect.Complex128:
		complexValue, err := strconv.ParseComplex(value, 128)
		if err != nil {
			return fmt.Errorf("failed to convert %s to complex: %v", value, err)
		}
		field.SetComplex(complexValue)
	case reflect.Slice:
		field.Set(reflect.ValueOf(strings.Split(value, ",")))
	default:
		// Handle unsupported types
		return fmt.Errorf("unsupported field type %s for value %s", field.Kind(), value)
	}
	return nil
}
