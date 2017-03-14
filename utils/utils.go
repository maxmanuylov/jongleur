package utils

import (
    "fmt"
    "reflect"
    "strconv"
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

func ParsePort(portStr string) (int, error) {
    port, err := strconv.Atoi(portStr)
    if err != nil {
        return 0, fmt.Errorf("Invalid port value \"%s\": %v", portStr, err)
    }

    if port < 0 || port > 0xFFFF {
        return 0, fmt.Errorf("Port value out of range: %d", port)
    }

    return port, nil
}

type UsageError struct {
    message string
}

func (err UsageError) Error() string {
    return err.message
}
