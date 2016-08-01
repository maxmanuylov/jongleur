package utils

import (
    "fmt"
    "reflect"
    "strings"
)

func Check(objectPointer interface{}) error {
    value := reflect.ValueOf(objectPointer).Elem()
    _type := reflect.TypeOf(objectPointer).Elem()

    for i := 0; i < value.NumField(); i++ {
        if isZero(value.Field(i)) {
            return UsageError{message: fmt.Sprintf("\"%s\" option is not specified", strings.ToLower(_type.Field(i).Name))}
        }
    }

    return nil
}

func isZero(value reflect.Value) bool {
    valueType := value.Type()
    if valueType.Kind() == reflect.Bool {
        return false
    }
    return reflect.DeepEqual(value.Interface(), reflect.Zero(valueType).Interface())
}

type UsageError struct {
    message string
}

func (err UsageError) Error() string {
    return err.message
}

func EtcdItemsKey(itemType string) string {
    return fmt.Sprintf("/jongleur/items/%s", itemType)
}