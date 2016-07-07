package util

import (
    "fmt"
    "os"
    "os/signal"
    "reflect"
    "strings"
    "syscall"
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

func WaitForTermination() {
    signalsChannel := make(chan os.Signal, 1)
    signal.Notify(signalsChannel, syscall.SIGINT, syscall.SIGTERM)
    <-signalsChannel
}

func EtcdItemsKey(itemType string) string {
    return fmt.Sprintf("/jongleur/items/%s", itemType)
}