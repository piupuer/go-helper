package query

import (
	"fmt"
	"github.com/piupuer/go-helper/pkg/log"
	localUtils "github.com/piupuer/go-helper/pkg/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	"gorm.io/gorm/utils"
	"reflect"
	"sort"
	"strings"
)

// process preload column
func (rd *Redis) processPreload() {
	if rd.Error == nil && len(rd.Statement.preloads) > 0 {
		preloadMap := map[string][]string{}
		for _, item := range rd.Statement.preloads {
			preloadFields := strings.Split(item.schema, ".")
			for idx := range preloadFields {
				preloadMap[strings.Join(preloadFields[:idx+1], ".")] = preloadFields[:idx+1]
			}
		}

		preloadNames := make([]string, len(preloadMap))
		idx := 0
		for key := range preloadMap {
			preloadNames[idx] = key
			idx++
		}
		sort.Strings(preloadNames)

		for _, name := range preloadNames {
			var (
				curSchema     = rd.Statement.Schema
				preloadFields = preloadMap[name]
				rels          = make([]*schema.Relationship, 0)
			)

			for _, preloadField := range preloadFields {
				if rel := curSchema.Relationships.Relations[preloadField]; rel != nil {
					rels = append(rels, rel)
					curSchema = rel.FieldSchema
				} else {
					rd.AddError(fmt.Errorf("%v: %w", name, gorm.ErrUnsupportedRelation))
				}
			}

			if len(rels) > 0 {
				preload(rd, rels)
			}
		}
	}
}

func preload(db *Redis, rels []*schema.Relationship) {
	var (
		reflectValue     = db.Statement.ReflectValue
		rel              = rels[len(rels)-1]
		relForeignKeys   []string
		relForeignFields []*schema.Field
		foreignFields    []*schema.Field
		foreignValues    [][]interface{}
		identityMap      = map[string][]reflect.Value{}
	)

	if len(rels) > 1 {
		reflectValue = schema.GetRelationsValues(reflectValue, rels[:len(rels)-1])
	}

	ins := db.Session()

	if rel.JoinTable != nil {
		// many2many join
		var joinForeignFields, joinRelForeignFields []*schema.Field
		var joinForeignKeys []string
		for _, ref := range rel.References {
			if ref.OwnPrimaryKey {
				joinForeignKeys = append(joinForeignKeys, ref.ForeignKey.DBName)
				joinForeignFields = append(joinForeignFields, ref.ForeignKey)
				foreignFields = append(foreignFields, ref.PrimaryKey)
			} else if ref.PrimaryValue != "" {
				ins = ins.Where(localUtils.CamelCaseLowerFirst(ref.ForeignKey.DBName), "=", ref.PrimaryValue)
			} else {
				joinRelForeignFields = append(joinRelForeignFields, ref.ForeignKey)
				relForeignKeys = append(relForeignKeys, ref.PrimaryKey.DBName)
				relForeignFields = append(relForeignFields, ref.PrimaryKey)
			}
		}

		joinIdentityMap, joinForeignValues := schema.GetIdentityFieldValuesMap(reflectValue, foreignFields)
		if len(joinForeignValues) == 0 {
			return
		}

		// rel.JoinTable.MakeSlice().Elem() convert to map
		// joinResults is pointer array like []*users
		joinResults := rel.JoinTable.MakeSlice().Elem()
		joinResultsType := joinResults.Type()
		// get real elem
		joinResultType := joinResultsType.Elem().Elem()
		relatedRows := make([]map[string]interface{}, 0)
		column, values := schema.ToQueryValues(rel.JoinTable.Table, joinForeignKeys, joinForeignValues)
		// column type: 1.clause.Column 2.[]clause.Column
		cols := make([]clause.Column, 0)
		if col1, ok := column.(clause.Column); ok {
			cols = append(cols, col1)
		} else if col2, ok := column.([]clause.Column); ok {
			cols = append(cols, col2...)
		}
		for _, col := range cols {
			ins.AddError(
				ins.
					Table(localUtils.CamelCaseLowerFirst(rel.JoinTable.Name)).
					Where(localUtils.CamelCaseLowerFirst(col.Name), "in", toQueryValues(values)).
					Find(&relatedRows).
					Error,
			)
		}

		// convert join identity map to relation identity map
		fieldValues := make([]interface{}, len(joinForeignFields))
		joinFieldValues := make([]interface{}, len(joinRelForeignFields))

		for _, result := range relatedRows {
			for idx, field := range joinForeignFields {
				fieldValues[idx] = result[localUtils.CamelCaseLowerFirst(field.Name)]
			}

			for idx, field := range joinRelForeignFields {
				joinFieldValues[idx] = result[localUtils.CamelCaseLowerFirst(field.Name)]
			}

			if results, ok := joinIdentityMap[utils.ToStringKey(fieldValues...)]; ok {
				joinKey := utils.ToStringKey(joinFieldValues...)
				identityMap[joinKey] = append(identityMap[joinKey], results...)
			}
			// reflect.New is pointer
			itemPtr := reflect.New(joinResultType)
			item := itemPtr.Elem()
			for _, field := range rel.JoinTable.PrimaryFields {
				// redis json number is float64, need convert
				item.
					FieldByName(field.Name).
					Set(
						reflect.ValueOf(
							result[localUtils.CamelCaseLowerFirst(field.DBName)],
						).Convert(item.FieldByName(field.Name).Type()),
					)
				joinResults = reflect.Append(joinResults, itemPtr)
			}
		}

		_, foreignValues = schema.GetIdentityFieldValuesMap(joinResults, joinRelForeignFields)
	} else {
		// one2many or many2one
		for _, ref := range rel.References {
			if ref.OwnPrimaryKey {
				relForeignKeys = append(relForeignKeys, ref.ForeignKey.DBName)
				relForeignFields = append(relForeignFields, ref.ForeignKey)
				foreignFields = append(foreignFields, ref.PrimaryKey)
			} else if ref.PrimaryValue != "" {
				ins = ins.Where(localUtils.CamelCaseLowerFirst(ref.ForeignKey.DBName), "=", ref.PrimaryValue)
			} else {
				relForeignKeys = append(relForeignKeys, ref.PrimaryKey.DBName)
				relForeignFields = append(relForeignFields, ref.PrimaryKey)
				foreignFields = append(foreignFields, ref.ForeignKey)
			}
		}

		identityMap, foreignValues = schema.GetIdentityFieldValuesMap(reflectValue, foreignFields)
		if len(foreignValues) == 0 {
			return
		}
	}

	reflectResults := rel.FieldSchema.MakeSlice().Elem()
	column, values := schema.ToQueryValues(clause.CurrentTable, relForeignKeys, foreignValues)

	// column: 1.clause.Column 2.[]clause.Column
	cols := make([]clause.Column, 0)
	if col1, ok1 := column.(clause.Column); ok1 {
		cols = append(cols, col1)
	} else if col2, ok2 := column.([]clause.Column); ok2 {
		cols = append(cols, col2...)
	}
	for _, col := range cols {
		ins.AddError(
			ins.
				Table(localUtils.CamelCaseLowerFirst(rel.FieldSchema.Name)).
				Where(localUtils.CamelCaseLowerFirst(col.Name), "in", toQueryValues(values)).
				Find(reflectResults.Addr().Interface()).
				Error,
		)
	}
	fieldValues := make([]interface{}, len(relForeignFields))

	// clean up old values before preloading
	switch reflectValue.Kind() {
	case reflect.Struct:
		switch rel.Type {
		case schema.HasMany, schema.Many2Many:
			rel.Field.Set(reflectValue, reflect.MakeSlice(rel.Field.IndirectFieldType, 0, 0).Interface())
		default:
			rel.Field.Set(reflectValue, reflect.New(rel.Field.FieldType).Interface())
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < reflectValue.Len(); i++ {
			switch rel.Type {
			case schema.HasMany, schema.Many2Many:
				rel.Field.Set(reflectValue.Index(i), reflect.MakeSlice(rel.Field.IndirectFieldType, 0, 0).Interface())
			default:
				rel.Field.Set(reflectValue.Index(i), reflect.New(rel.Field.FieldType).Interface())
			}
		}
	}

	for i := 0; i < reflectResults.Len(); i++ {
		elem := reflectResults.Index(i)
		for idx, field := range relForeignFields {
			fieldValues[idx], _ = field.ValueOf(elem)
		}

		for _, data := range identityMap[utils.ToStringKey(fieldValues...)] {
			reflectFieldValue := rel.Field.ReflectValueOf(data)
			if reflectFieldValue.Kind() == reflect.Ptr && reflectFieldValue.IsNil() {
				reflectFieldValue.Set(reflect.New(rel.Field.FieldType.Elem()))
			}

			reflectFieldValue = reflect.Indirect(reflectFieldValue)
			switch reflectFieldValue.Kind() {
			case reflect.Struct:
				rel.Field.Set(data, reflectResults.Index(i).Interface())
			case reflect.Slice, reflect.Array:
				if reflectFieldValue.Type().Elem().Kind() == reflect.Ptr {
					rel.Field.Set(data, reflect.Append(reflectFieldValue, elem).Interface())
				} else {
					rel.Field.Set(data, reflect.Append(reflectFieldValue, elem.Elem()).Interface())
				}
			}
		}
	}
}

func toQueryValues(values []interface{}) (results []int) {
	for _, v := range values {
		// uint primary key convert to redis json int
		if item1, ok1 := v.(uint); ok1 {
			results = append(results, int(item1))
		} else if item2, ok2 := v.(int); ok2 {
			results = append(results, item2)
		} else {
			log.Warn("the primary key type of the current association table is incompatible")
		}
	}
	return
}
