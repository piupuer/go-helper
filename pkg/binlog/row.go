package binlog

import (
	"fmt"
	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/schema"
	"github.com/golang-module/carbon"
	"github.com/piupuer/go-helper/pkg/utils"
)

const (
	idName        = "id"
	deletedAtName = "deletedAt"
)

// mysql row change, set to redis
func RowChange(ops Options, e *canal.RowsEvent) {
	database := e.Table.Schema
	table := e.Table.Name
	idIndex := -1
	deletedAtIndex := -1
	for i, column := range e.Table.Columns {
		name := utils.CamelCaseLowerFirst(column.Name)
		if name == idName {
			idIndex = i
		}
		if name == deletedAtName {
			deletedAtIndex = i
		}
		if idIndex >= 0 && deletedAtIndex >= 0 {
			break
		}
	}
	// gorm v2 e.Rows some fields type is []uint8(alias for []byte)
	// so convert uint8 to string
	rows := make([][]interface{}, len(e.Rows))
	for i, eRow := range e.Rows {
		row := make([]interface{}, len(eRow))
		for j, eItem := range eRow {
			if eV, ok := eItem.([]uint8); ok {
				row[j] = string(eV)
			} else {
				columnType := e.Table.Columns[j].RawType
				// convert to carbon.DateTimeString
				// grom v2 time type is datetime(3)
				if t, ok := eItem.(string); ok && columnType == "datetime(3)" {
					eItem = carbon.ToDateTimeString{
						Carbon: carbon.Parse(t),
					}
				}
				row[j] = eItem
			}
		}
		rows[i] = row
	}

	cacheKey := fmt.Sprintf("%s_%s", database, table)
	// get old rows
	oldRowsStr, err := ops.redis.Get(ops.ctx, cacheKey).Result()
	newRows := make([]map[string]interface{}, 0)
	changeRows := make([][]interface{}, 0)
	if err == nil {
		// decompress
		oldRows := utils.DeCompressStrByZlib(oldRowsStr)
		utils.Json2Struct(oldRows, &newRows)
	}
	rowCount := len(newRows)
	// convert rows to json to keep same type with oldRows
	utils.Struct2StructByJson(rows, &changeRows)

	switch e.Action {
	case canal.InsertAction:
		// insert change
		for _, changeRow := range changeRows {
			row := getRow(ops, changeRow, e.Table)
			if row[deletedAtName] == nil {
				// when deleteAt is null to set cache because gorm soft deleted
				newRows = append(newRows, row)
			}
		}
		break
	case canal.UpdateAction:
		// update change
		// two item is one group
		for i, l := 0, len(changeRows); i < l; i += 2 {
			oldRow := changeRows[i]
			newRow := changeRows[i+1]
			// get old row index
			index := getOldRowIndex(ops, newRows, oldRow, e.Table)
			if len(newRows) > 0 && index >= 0 {
				if deletedAtIndex >= 0 && oldRow[deletedAtIndex] == nil && newRow[deletedAtIndex] != nil {
					if index < rowCount-1 {
						newRows = append(newRows[:index], newRows[index+1:]...)
					} else {
						newRows = append(newRows[:index])
					}
				} else {
					newRows[index] = getRow(ops, newRow, e.Table)
				}
			} else {
				newRows = append(newRows, getRow(ops, newRow, e.Table))
			}
		}

		break
	case canal.DeleteAction:
		indexes := make([]int, 0)
		for _, changeRow := range changeRows {
			index := getOldRowIndex(ops, newRows, changeRow, e.Table)
			if index > -1 {
				indexes = append(indexes, index)
			}
		}
		deletedCount := 0
		for _, index := range indexes {
			i := index - deletedCount
			if index < rowCount-1 {
				newRows = append(newRows[:i], newRows[i+1:]...)
				deletedCount++
			} else {
				newRows = append(newRows[:i])
			}
		}
		break
	}
	compress, err := utils.CompressStrByZlib(utils.Struct2Json(newRows))
	if err != nil {
		ops.logger.Error(ops.ctx, "[binlog row change]compress err: %v, %v", err, e)
		return
	}
	err = ops.redis.Set(ops.ctx, cacheKey, compress, 0).Err()
	if err != nil {
		ops.logger.Error(ops.ctx, "[binlog row change]set to redis err: %v, %v", err, e)
	}
}

// get old row index
func getOldRowIndex(ops Options, oldRows []map[string]interface{}, data []interface{}, table *schema.Table) int {
	newRow := getRow(ops, data, table)
	for i, row := range oldRows {
		m := make(map[string]interface{}, 0)
		utils.CompareDiff(row, newRow, &m)
		// no change
		if len(m) == 0 {
			return i
		}
	}
	return -1
}

// get row fields map from data
func getRow(ops Options, data []interface{}, table *schema.Table) map[string]interface{} {
	row := make(map[string]interface{}, 0)
	count := len(data)
	for i, column := range table.Columns {
		var item interface{}
		if i < count {
			// canal not convert tinyint(1), custom conversion to uint
			if column.RawType == "tinyint(1)" {
				switch data[i].(type) {
				// canal's tinyint(1) type is float64
				case float64:
					item = uint(data[i].(float64))
					break
				}
			} else {
				item = data[i]
			}
			row[utils.CamelCaseLowerFirst(column.Name)] = item
		}
	}
	if count != len(table.Columns) {
		ops.logger.Warn(ops.ctx, "[binlog get row], columns: %v, data: %v", table.Columns, data)
	}
	return row
}

func PosChange(ops Options, pos mysql.Position) {
	err := ops.redis.Set(ops.ctx, fmt.Sprintf("%s_%s", ops.binlogPos, ops.dsn.DBName), utils.Struct2Json(pos), 0).Err()
	if err != nil {
		ops.logger.Error(ops.ctx, "[binlog pos change]err: %v, %v", err, pos)
	}
}
