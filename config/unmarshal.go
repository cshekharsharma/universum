package config

import (
	"fmt"
	"reflect"
	"strconv"
)

func Unmarshal(data TOMLTable, result interface{}) error {
	v := reflect.ValueOf(result)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("result argument must be a non-nil pointer to a struct")
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("result argument must be a pointer to a struct")
	}

	return unmarshalStruct(data, v)
}

func unmarshalStruct(data TOMLTable, v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		tag := fieldType.Tag.Get("toml")

		key := tag
		if key == "" {
			key = fieldType.Name
		}

		tomlValue, exists := data[key]
		if !exists {
			continue // Skip if TOML key doesn't exist
		}

		if field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct {
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			nestedTable, ok := tomlValue.(TOMLTable)
			if !ok {
				return fmt.Errorf("expected nested table for struct field %s", key)
			}
			if err := unmarshalStruct(nestedTable, field.Elem()); err != nil {
				return err
			}
			continue
		}

		if field.Kind() == reflect.Struct {
			nestedTable, ok := tomlValue.(TOMLTable)
			if !ok {
				return fmt.Errorf("expected nested table for struct field %s", key)
			}
			if err := unmarshalStruct(nestedTable, field); err != nil {
				return err
			}
			continue
		}

		if err := setFieldValue(field, tomlValue); err != nil {
			return fmt.Errorf("failed to set field %s: %v", key, err)
		}
	}

	return nil
}

func setFieldValue(field reflect.Value, value TOMLValue) error {
	if !field.CanSet() {
		return fmt.Errorf("cannot set field")
	}

	switch field.Kind() {
	case reflect.Bool:
		if v, ok := value.(bool); ok {
			field.SetBool(v)
		} else {
			return fmt.Errorf("expected bool, got %T", value)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v, ok := value.(int); ok {
			field.SetInt(int64(v))
		} else if vStr, ok := value.(string); ok {
			intValue, err := strconv.Atoi(vStr)
			if err != nil {
				return fmt.Errorf("expected int, got %s", vStr)
			}
			field.SetInt(int64(intValue))
		} else {
			return fmt.Errorf("expected int, got %T", value)
		}
	case reflect.Float32, reflect.Float64:
		if v, ok := value.(float64); ok {
			field.SetFloat(v)
		} else {
			return fmt.Errorf("expected float, got %T", value)
		}
	case reflect.String:
		if v, ok := value.(string); ok {
			field.SetString(v)
		} else {
			return fmt.Errorf("expected string, got %T", value)
		}
	case reflect.Slice:
		// Handle array (slice) types
		return setSliceValue(field, value)
	default:
		return fmt.Errorf("unsupported kind %s", field.Kind())
	}

	return nil
}

func setSliceValue(field reflect.Value, value TOMLValue) error {
	if arr, ok := value.([]TOMLValue); ok {
		slice := reflect.MakeSlice(field.Type(), len(arr), len(arr))

		for i, elem := range arr {
			err := setFieldValue(slice.Index(i), elem)
			if err != nil {
				return err
			}
		}

		field.Set(slice)
		return nil
	}

	return fmt.Errorf("expected array, got %T", value)
}
