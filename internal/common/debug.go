package common

import (
	"fmt"
	"reflect"
)

func PrintStruct(v interface{}, printFunc func(string)) {
	// Check if the printFunc is nil
	if printFunc == nil {
		printFunc = func(s string) {
			fmt.Println(s)
		}
	}

	// Print the struct with an empty indent
	printStruct(v, "", printFunc)
}

// printStruct prints the fields of a struct in a readable format.
func printStruct(v interface{}, indent string, printFunc func(string)) {
	val := reflect.ValueOf(v)

	// If the value is a pointer, dereference it
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			printFunc(fmt.Sprintf("%snil", indent))
			return
		}
		val = val.Elem()
	}

	// Check if the value is a struct
	if val.Kind() == reflect.Struct {
		for i := 0; i < val.NumField(); i++ {
			field := val.Type().Field(i)
			fieldValue := val.Field(i)
			if !field.IsExported() {
				continue
			}

			// Construct the line for the field name and value
			if fieldValue.Kind() == reflect.Ptr {
				if fieldValue.IsNil() {
					printFunc(fmt.Sprintf("%s%s: nil", indent, field.Name))
				} else {
					printFunc(fmt.Sprintf("%s%s:", indent, field.Name))
					printStruct(fieldValue.Interface(), indent+"  ", printFunc)
				}
			} else if fieldValue.Kind() == reflect.Struct {
				printFunc(fmt.Sprintf("%s%s:", indent, field.Name))
				printStruct(fieldValue.Interface(), indent+"  ", printFunc)
			} else {
				printFunc(fmt.Sprintf("%s%s: %v", indent, field.Name, fieldValue))
			}
		}
	} else {
		// TODO: Handle other types (e.g., map, slice, etc.)
		printFunc(fmt.Sprintf("%v", val.Interface()))
	}
}
