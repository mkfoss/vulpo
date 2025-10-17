package vulpo

/*
#cgo CFLAGS: -I./mkfdbflib
#cgo LDFLAGS: -L./mkfdbflib -lmkfdbf
#include "d4all.h"
#include <stdlib.h>
*/
import "C"
import (
	"unsafe"
)

// ExprFilter represents a compiled dBASE expression for filtering records
type ExprFilter struct {
	expr     *C.EXPR4
	vulpo    *Vulpo
	exprText string
}

// NewExprFilter creates a new expression filter from a dBASE expression string
func (v *Vulpo) NewExprFilter(expression string) (*ExprFilter, error) {
	if !v.Active() {
		return nil, NewError("database not open")
	}

	// Convert Go string to C string
	cExpr := C.CString(expression)
	defer C.free(unsafe.Pointer(cExpr))

	// Parse the expression using CodeBase (use the low-level function directly)
	expr := C.expr4parseLow(v.data, cExpr, nil)
	if expr == nil {
		// Get error information from CodeBase
		return nil, NewErrorf("failed to parse expression: %s", expression)
	}

	return &ExprFilter{
		expr:     expr,
		vulpo:    v,
		exprText: expression,
	}, nil
}

// Free releases the memory associated with the expression filter
func (ef *ExprFilter) Free() {
	if ef.expr != nil {
		C.u4freeDefault(unsafe.Pointer(ef.expr))
		ef.expr = nil
	}
}

// Evaluate evaluates the expression for the current record and returns the result as a boolean
func (ef *ExprFilter) Evaluate() (bool, error) {
	if ef.expr == nil {
		return false, NewError("expression filter is not initialized")
	}

	// Evaluate the expression - this should return a logical result
	result := C.expr4true(ef.expr)
	return result != 0, nil
}

// EvaluateAsString evaluates the expression and returns the result as a string
func (ef *ExprFilter) EvaluateAsString() (string, error) {
	if ef.expr == nil {
		return "", NewError("expression filter is not initialized")
	}

	// Get the string result of the expression
	cResult := C.expr4str(ef.expr)
	if cResult == nil {
		return "", NewError("expression evaluation returned null")
	}

	return C.GoString(cResult), nil
}

// EvaluateAsDouble evaluates the expression and returns the result as a float64
func (ef *ExprFilter) EvaluateAsDouble() (float64, error) {
	if ef.expr == nil {
		return 0, NewError("expression filter is not initialized")
	}

	// Get the double result of the expression
	result := C.expr4double(ef.expr)
	return float64(result), nil
}

// ExprSearchOptions contains options for expression-based searching
type ExprSearchOptions struct {
	MaxResults int  // Maximum number of results to return (0 for unlimited)
	UseIndex   bool // Whether to try to use indexes for optimization
}

// ExprMatch represents a single expression match result
type ExprMatch struct {
	RecordNumber int                    // 1-indexed record number
	FieldReaders map[string]FieldReader // Field readers for accessing the record
}

// ExprSearchResult contains all matches from an expression search
type ExprSearchResult struct {
	Expression   string      // The expression used
	Matches      []ExprMatch // All matching records
	TotalScanned int         // Total records scanned
	TotalMatched int         // Total records that matched
}

// SearchByExpression searches for records matching a dBASE expression
func (v *Vulpo) SearchByExpression(expression string, options *ExprSearchOptions) (*ExprSearchResult, error) {
	if !v.Active() {
		return nil, NewError("database not open")
	}

	if options == nil {
		options = &ExprSearchOptions{}
	}

	// Create expression filter
	filter, err := v.NewExprFilter(expression)
	if err != nil {
		return nil, NewErrorf("failed to create expression filter: %v", err)
	}
	defer filter.Free()

	result := &ExprSearchResult{
		Expression: expression,
		Matches:    make([]ExprMatch, 0),
	}

	// Save original position
	originalPosition := v.Position()
	defer func() {
		if originalPosition > 0 {
			_ = v.Goto(originalPosition) // Ignore error in defer
		}
	}()

	// Go to the first record
	err = v.First()
	if err != nil {
		return nil, NewErrorf("failed to go to first record: %v", err)
	}

	// Iterate through all records
	for !v.EOF() {
		result.TotalScanned++

		// Evaluate the expression for the current record
		matches, err := filter.Evaluate()
		if err != nil {
			return nil, NewErrorf("failed to evaluate expression: %v", err)
		}

		if matches {
			// Create field readers for all fields
			fieldReaders := make(map[string]FieldReader)
			for i := 0; i < v.FieldCount(); i++ {
				fieldDef := v.Field(i)
				if fieldDef != nil {
					fieldReader, err := v.getFieldReader(fieldDef.Name())
					if err == nil {
						fieldReaders[fieldDef.Name()] = fieldReader
					}
				}
			}

			match := ExprMatch{
				RecordNumber: v.Position(),
				FieldReaders: fieldReaders,
			}

			result.Matches = append(result.Matches, match)
			result.TotalMatched++

			// Check if we've reached the maximum number of results
			if options.MaxResults > 0 && result.TotalMatched >= options.MaxResults {
				break
			}
		}

		// Move to the next record
		err = v.Next()
		if err != nil {
			break // End of file or error
		}
	}

	return result, nil
}

// CountByExpression counts the number of records matching a dBASE expression
func (v *Vulpo) CountByExpression(expression string) (int, error) {
	if !v.Active() {
		return 0, NewError("database not open")
	}

	// Create expression filter
	filter, err := v.NewExprFilter(expression)
	if err != nil {
		return 0, NewErrorf("failed to create expression filter: %v", err)
	}
	defer filter.Free()

	count := 0

	// Save original position
	originalPosition := v.Position()
	defer func() {
		if originalPosition > 0 {
			_ = v.Goto(originalPosition) // Ignore error in defer
		}
	}()

	// Go to the first record
	err = v.First()
	if err != nil {
		return 0, NewErrorf("failed to go to first record: %v", err)
	}

	// Iterate through all records
	for !v.EOF() {
		// Evaluate the expression for the current record
		matches, err := filter.Evaluate()
		if err != nil {
			return 0, NewErrorf("failed to evaluate expression: %v", err)
		}

		if matches {
			count++
		}

		// Move to the next record
		err = v.Next()
		if err != nil {
			break // End of file or error
		}
	}

	return count, nil
}

// ForEachExpressionMatch iterates through records matching a dBASE expression
func (v *Vulpo) ForEachExpressionMatch(expression string, callback func(map[string]FieldReader) error) error {
	if !v.Active() {
		return NewError("database not open")
	}

	// Create expression filter
	filter, err := v.NewExprFilter(expression)
	if err != nil {
		return NewErrorf("failed to create expression filter: %v", err)
	}
	defer filter.Free()

	// Save original position
	originalPosition := v.Position()
	defer func() {
		if originalPosition > 0 {
			_ = v.Goto(originalPosition) // Ignore error in defer
		}
	}()

	// Go to the first record
	err = v.First()
	if err != nil {
		return NewErrorf("failed to go to first record: %v", err)
	}

	// Iterate through all records
	for !v.EOF() {
		// Evaluate the expression for the current record
		matches, err := filter.Evaluate()
		if err != nil {
			return NewErrorf("failed to evaluate expression: %v", err)
		}

		if matches {
			// Create field readers for all fields
			fieldReaders := make(map[string]FieldReader)
			for i := 0; i < v.FieldCount(); i++ {
				fieldDef := v.Field(i)
				if fieldDef != nil {
					fieldReader, err := v.getFieldReader(fieldDef.Name())
					if err == nil {
						fieldReaders[fieldDef.Name()] = fieldReader
					}
				}
			}

			// Call the callback function
			if err := callback(fieldReaders); err != nil {
				return err
			}
		}

		// Move to the next record
		err = v.Next()
		if err != nil {
			break // End of file or error
		}
	}

	return nil
}

// GetExpressionText returns the original expression text
func (ef *ExprFilter) GetExpressionText() string {
	return ef.exprText
}

// IsValid checks if the expression filter is valid and usable
func (ef *ExprFilter) IsValid() bool {
	return ef.expr != nil && ef.vulpo != nil
}
