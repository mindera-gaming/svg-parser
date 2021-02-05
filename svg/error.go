package svg

import (
	"fmt"
	"strings"
)

type EmptyCoordinateError struct {
	Command string
}

func newEmptyCoordinateError(command string) EmptyCoordinateError {
	return EmptyCoordinateError{
		Command: command,
	}
}

func (e EmptyCoordinateError) Error() string {
	return fmt.Sprintf("%s does not contain coordinate data", e.Command)
}

type InvalidCoordinateError struct {
	Command string
	Data    string
}

func newInvalidCoordinateError(command string, data []string) InvalidCoordinateError {
	return InvalidCoordinateError{
		Command: command,
		Data:    strings.Join(data, " "),
	}
}

func (e InvalidCoordinateError) Error() string {
	return fmt.Sprintf("%s does not contain a valid coordinate or set of coordinates: %s", e.Command, e.Data)
}

type InvalidXError struct {
	Command string
	Data    string
}

func newInvalidXError(command, data string) InvalidXError {
	return InvalidXError{
		Command: command,
		Data:    data,
	}
}

func (e InvalidXError) Error() string {
	return fmt.Sprintf("%s does not contain a valid x: %s", e.Command, e.Data)
}

type InvalidYError struct {
	Command string
	Data    string
}

func newInvalidYError(command, data string) InvalidYError {
	return InvalidYError{
		Command: command,
		Data:    data,
	}
}

func (e InvalidYError) Error() string {
	return fmt.Sprintf("%s does not contain a valid y: %s", e.Command, e.Data)
}

type UnsupportedCommandError struct {
	Command string
}

func newUnsupportedCommandError(command string) UnsupportedCommandError {
	return UnsupportedCommandError{
		Command: command,
	}
}

func (e UnsupportedCommandError) Error() string {
	return fmt.Sprintf("%s is not supported", e.Command)
}
