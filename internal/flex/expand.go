package flex

import (
	"context"
	"fmt"
	"reflect"
)

// Expand "expands" a resource's "business logic" data structure,
// implemented using Terraform Plugin Framework data types, into
// an AWS SDK for Go v2 API data structure.
// The resource's data structure is walked and exported fields that
// have a corresponding field in the API data structure (and a suitable
// target data type) are copied.
func Expand(ctx context.Context, tfObject, apiObject any) error {
	if err := walkStructFields(ctx, tfObject, apiObject, expandVisitor{}); err != nil {
		return fmt.Errorf("Expand[%T, %T]: %w", tfObject, apiObject, err)
	}

	return nil
}

// walkStructFields traverses `from` calling `visitor` for each exported field.
func walkStructFields(ctx context.Context, from any, to any, visitor fieldVisitor) error {
	valFrom, valTo := reflect.ValueOf(from), reflect.ValueOf(to)

	if kind := valFrom.Kind(); kind == reflect.Ptr {
		valFrom = valFrom.Elem()
	}
	if kind := valTo.Kind(); kind != reflect.Ptr {
		return fmt.Errorf("target (%T): %s, want pointer", to, kind)
	}
	valTo = valTo.Elem()

	typFrom, typTo := valFrom.Type(), valTo.Type()

	if typFrom.Kind() != reflect.Struct {
		return fmt.Errorf("source: %s, want struct", typFrom)
	}
	if typTo.Kind() != reflect.Struct {
		return fmt.Errorf("target: %s, want struct", typTo)
	}

	for i := 0; i < typFrom.NumField(); i++ {
		field := typFrom.Field(i)
		if field.PkgPath != "" {
			continue // Skip unexported fields.
		}
		fieldName := field.Name
		toFieldVal := valTo.FieldByName(fieldName)
		if !toFieldVal.IsValid() {
			continue // Corresponding field not found in to.
		}
		if err := visitor.visit(ctx, fieldName, valFrom.Field(i), toFieldVal); err != nil {
			return fmt.Errorf("visit (%s): %w", fieldName, err)
		}
	}

	return nil
}

type fieldVisitor interface {
	visit(context.Context, string, reflect.Value, reflect.Value) error
}

type expandVisitor struct{}

func (v expandVisitor) visit(ctx context.Context, fieldName string, valFrom, valTo reflect.Value) error {
	return nil
}
