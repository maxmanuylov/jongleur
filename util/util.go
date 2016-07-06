package util

import (
    "reflect"
    "fmt"
    "strings"
)

func Check(objectPointer interface{}) error {
    value := reflect.ValueOf(objectPointer).Elem()
    for i := 0; i < value.NumField(); i++ {
        field := value.Field(i)
        if reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()) {
            return UsageError{fmt.Sprintf("\"%s\" option is not specified", strings.ToLower(reflect.TypeOf(objectPointer).Elem().Field(i).Name))}
        }
    }
    return nil
}

type UsageError struct {
    message string
}

func (err UsageError) Error() string {
    return err.message
}


