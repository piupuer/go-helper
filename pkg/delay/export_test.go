package delay

import (
	"fmt"
	"github.com/piupuer/go-helper/pkg/req"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"testing"
	"time"
)

func TestNewExport(t *testing.T) {
	db, _ := gorm.Open(mysql.Open("root:root@tcp(127.0.0.1:4306)/gin_web?charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True&loc=Local&timeout=10000ms"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "tb_",
			SingularTable: true,
		},
	})
	MigrateExport(
		WithExportDbNoTx(db),
	)
	ex := NewExport(
		WithExportDbNoTx(db),
	)
	ex.Start("uuid1", "export 1", "category 1", "start")
	i := 0
	for {
		if i >= 100 {
			break
		}
		ex.Pending("uuid1", fmt.Sprintf("%d%%", i))
		i++
		time.Sleep(10 * time.Millisecond)
	}
	ex.End("uuid1", fmt.Sprintf("%d%%", i), "/tmp/1.xlsx")

	fmt.Println(ex.FindHistory(&req.DelayExportHistory{}))
}
