package cliced

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Name of the tag to parse.
const tagName = "cliced"

// Value to prefix to name value.
const namePrefix = "--"

// Value to prefix to shortName value.
const shortNamePrefix = "-"

// Value of the delimiter between constraint
// key-value pairs.
const constraintValueDelimiter = ":"

// Value of the delimiter between constraints.
const constraintsDelimiter = ";"

// Struct for stroring key-value string pair
type keyValuePair struct {
	key   string
	value string
}

// Split a constraint as key-value constraint
func splitConstraint(constraint string) (keyValuePair, error) {
	parts := strings.Split(constraint, constraintValueDelimiter)
	switch len(parts) {
	case 1:
		return keyValuePair{
			parts[0], "",
		}, nil
	case 2:
		return keyValuePair{
			parts[0], parts[1],
		}, nil
	}
	return keyValuePair{}, fmt.Errorf("syntax error too many characters %s ", constraintValueDelimiter)
}

// Struct defining a parameter from a structField.
type parameter struct {
	// Name of the parameter arguments are tested
	// by appending an underscore to this value.
	name string
	// ShortName of the parameter argument.
	// ShortName matches are evaluated after
	// appending two underscore to this value.
	shortName string
	// Index of the structField this parameter
	// was created from.
	index int
	// Short description of the parameter
	// used for help and error messages.
	description string
	// If true not finding this parameter
	// will result is an error.
	mandatory bool
	// If true not finding this parameter
	// will result is an error.
	used bool
	// Value used to parse array types.
	delimiter string
	// Type of the parameter only types
	// bool,int,string,[]int,[]string are supported.
	tipe reflect.Type
}

// Getter for name.
func (p *parameter) Name() string {
	return p.name
}

// Getter for index.
func (p *parameter) Index() int {
	return p.index
}

// Getter for mandatory.
func (p *parameter) Mandatory() bool {
	return p.mandatory
}
func (p *parameter) CliNames() []string {
	if p.hasShortName() {
		return []string{
			fmt.Sprint(namePrefix, strings.ToLower(p.name)),
			fmt.Sprint(shortNamePrefix, strings.ToLower(p.shortName)),
		}
	}
	return []string{
		fmt.Sprint(namePrefix, strings.ToLower(p.name)),
	}
}

// Getter for delimiter.
func (p *parameter) Delimiter() string {
	return p.delimiter
}

// Splits a string by the delimiter.
func (p *parameter) Split(s string) []string {
	return strings.Split(s, p.delimiter)
}

func (p *parameter) GetHelp() string {
	var buffer bytes.Buffer
	buffer.WriteString(strings.Join(p.CliNames(), " "))
	buffer.WriteString(" ")
	buffer.WriteString(p.tipe.String())
	buffer.WriteString(" ")
	if p.IsArrayType() {
		buffer.WriteString("delimiter ")
		if p.delimiter == " " {
			buffer.WriteString("whitespace ")

		} else {
			buffer.WriteString(p.delimiter)
			buffer.WriteString(" ")
		}
	}
	if p.Mandatory() {
		buffer.WriteString("(mandatory)")
		buffer.WriteString(" ")
	}
	if p.description != "" {
		buffer.WriteString(": ")
		buffer.WriteString(p.description)
	}
	buffer.WriteString("\r\n")
	return buffer.String()
}

// Getter for tipe.
func (p *parameter) Type() reflect.Type {
	return p.tipe
}

// Returns if a shortName has been defined.
func (p *parameter) hasShortName() bool {
	return p.shortName != ""
}

// Getter for description.
func (p *parameter) Description() string {
	return p.description
}

// Returns if the parameter matches the string.
func (p *parameter) Matches(s string) bool {
	// TODO rewrite using the
	// getCliNames method
	for _, name := range p.CliNames() {
		if name == s {
			return true
		}
	}
	return false
}

// Getter for used.
// Serves as a marked for duplicate arguments.
func (p *parameter) Used() bool {
	return p.used
}

// Marks the parameters as used.
func (p *parameter) Use() {
	p.used = true
}

func (p *parameter) IsArrayType() bool {
	stringArrayType, intArrayType := reflect.TypeOf([]string{}), reflect.TypeOf([]int{})
	t := p.tipe
	return t == stringArrayType || t == intArrayType
}

// TODO comment better
// Gets value of the object by reflect
func (p *parameter) getValue(obj interface{}) reflect.Value {
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() == reflect.Ptr {
		objValue = objValue.Elem()
	}
	fieldValue := objValue.FieldByName(p.name)
	return fieldValue
}

// Sets
func (p *parameter) setBool(obj interface{}) func(value string) error {
	p.getValue(obj).SetBool(true)
	return nil
}

func (p *parameter) setInt(obj interface{}) func(value string) error {
	return func(value string) error {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		p.getValue(obj).SetInt(int64(intValue))
		return nil
	}
}
func (p *parameter) setString(obj interface{}) func(value string) error {
	return func(value string) error {
		p.getValue(obj).SetString(value)
		return nil
	}
}
func (p *parameter) setStringArray(obj interface{}) func(value string) error {
	return func(value string) error {
		parts := p.Split(value)
		p.getValue(obj).Set(reflect.ValueOf(parts))
		return nil
	}
}
func (p *parameter) setIntArray(obj interface{}) func(value string) error {
	return func(value string) error {
		parts := p.Split(value)
		intParts := []int{}
		for _, i := range parts {
			j, err := strconv.Atoi(i)
			if err != nil {
				return err
			}
			intParts = append(intParts, j)
		}
		p.getValue(obj).Set(reflect.ValueOf(intParts))
		return nil
	}
}

// fills an object with the desired value
func (p *parameter) SetterCallback(obj interface{}) (func(value string) error, error) {
	// TODO add parameter usage check
	switch p.tipe {
	case reflect.TypeOf(true):
		return p.setBool(obj), nil
	case reflect.TypeOf(1):
		return p.setInt(obj), nil
	case reflect.TypeOf(""):
		return p.setString(obj), nil
	case reflect.TypeOf([]string{}):
		return p.setStringArray(obj), nil
	case reflect.TypeOf([]int{}):
		return p.setIntArray(obj), nil
	}
	return nil, fmt.Errorf("Incompatible type")
}

// Changes the parameter by the value of the constraint.
func (param *parameter) fillParameter(constraint string) error {
	splittedConstraint, err := splitConstraint(constraint)
	key, value := splittedConstraint.key, splittedConstraint.value
	if err != nil {
		return err
	}
	switch key {
	case "description":
		param.description = value
		return nil
	case "shortname":
		param.shortName = value
		return nil
	case "mandatory":
		param.mandatory = true
		return nil
	case "delimiter":
		param.delimiter = value
		return nil
	}
	return fmt.Errorf("unknown key %s", splittedConstraint.value)
}

// Returns a new Paramter from the structField
func newParameter(sf reflect.StructField) (*parameter, error) {
	tag, newParam := sf.Tag.Get(tagName), parameter{
		name:  sf.Name,
		index: sf.Index[0],
		tipe:  sf.Type,
	}
	if newParam.IsArrayType() && newParam.delimiter == "" {
		newParam.delimiter = ","
	}
	if tag == "" {
		return &newParam, nil
	}
	constraints := strings.Split(tag, constraintsDelimiter)
	for _, constraint := range constraints {
		err := newParam.fillParameter(constraint)
		if err != nil {
			return nil, fmt.Errorf(
				"error parsing constraint %s at field %s : %e",
				constraint, newParam.name, err)
		}
	}
	return &newParam, nil
}