package dynmgrm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// compatibility check
var (
	_ gorm.Valuer = (*List)(nil)
	_ sql.Scanner = (*List)(nil)
)

// List is a DynamoDB list type.
//
// See: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/HowItWorks.NamingRulesDataTypes.html
type List []interface{}

// GormDataType returns the data type for Gorm.
func (l *List) GormDataType() string {
	return "dglist"
}

// Scan implements the [sql.Scanner#Scan]
//
// [sql.Scanner#Scan]: https://golang.org/pkg/database/sql/#Scanner
func (l *List) Scan(value interface{}) error {
	if len(*l) != 0 {
		return ErrCollectionAlreadyContainsItem
	}
	sv, ok := value.([]interface{})
	if !ok {
		return errors.Join(ErrFailedToCast, fmt.Errorf("incompatible %T and %T", l, value))
	}
	*l = sv
	return resolveCollectionsNestedInList(l)
}

// GormValue implements the [gorm.Valuer] interface.
//
// [gorm.Valuer]: https://pkg.go.dev/gorm.io/gorm#Valuer
func (l List) GormValue(_ context.Context, db *gorm.DB) clause.Expr {
	if err := resolveCollectionsNestedInList(&l); err != nil {
		_ = db.AddError(err)
		return clause.Expr{}
	}
	av, err := toDocumentAttributeValue[*types.AttributeValueMemberL](l)
	if err != nil {
		_ = db.AddError(err)
		return clause.Expr{}
	}
	return clause.Expr{SQL: "?", Vars: []interface{}{*av}}
}

// resolveCollectionsNestedInList resolves nested collection type attribute.
func resolveCollectionsNestedInList(l *List) error {
	for i, v := range *l {
		if v, ok := v.(map[string]interface{}); ok {
			m := Map{}
			err := m.Scan(v)
			if err != nil {
				*l = nil
				return err
			}
			(*l)[i] = m
			continue
		}
		if isCompatibleWithSet[int](v) {
			s := newSet[int]()
			if err := s.Scan(v); err == nil {
				(*l)[i] = s
				continue
			}
		}
		if isCompatibleWithSet[float64](v) {
			s := newSet[float64]()
			if err := s.Scan(v); err == nil {
				(*l)[i] = s
				continue
			}
		}
		if isCompatibleWithSet[string](v) {
			s := newSet[string]()
			if err := s.Scan(v); err == nil {
				(*l)[i] = s
				continue
			}
		}
		if isCompatibleWithSet[[]byte](v) {
			s := newSet[[]byte]()
			if err := s.Scan(v); err == nil {
				(*l)[i] = s
				continue
			}
		}
		if v, ok := v.([]interface{}); ok {
			il := List{}
			err := il.Scan(v)
			if err != nil {
				*l = nil
				return err
			}
			(*l)[i] = il
		}
	}
	return nil
}
