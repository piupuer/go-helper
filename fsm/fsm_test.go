package fsm

import (
	"fmt"
	"github.com/piupuer/go-helper/fsm/request"
	"github.com/piupuer/go-helper/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"testing"
)

var db *gorm.DB

func init() {
	dsn := "root:root@tcp(127.0.0.1:4306)/gin_web_stage?charset=utf8mb4&parseTime=True&loc=Local&timeout=10000ms"
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		// 禁用外键(指定外键时不会在mysql创建真实的外键约束)
		DisableForeignKeyConstraintWhenMigrating: true,
		// 指定表前缀
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: "tb_",
		},
	})
	if err != nil {
		panic(fmt.Sprintf("[单元测试]初始化mysql异常: %v", err))
	}
	db = db.Debug()

	// db.AutoMigrate(
	// 	new(Test),
	// )
}

func TestMigrate(t *testing.T) {
	Migrate(db, WithPrefix("tb_fsm"))
}

func TestFsm_CreateMachine(t *testing.T) {
	tx := db.Begin()
	f := New(tx)
	f.CreateMachine(request.CreateMachineReq{
		Name:                       "请假审批",
		SubmitterName:              "请假人",
		SubmitterEditFields:        "name,time,type",
		SubmitterConfirm:           models.ReqUint(1),
		SubmitterConfirmEditFields: "status",
		Levels: []request.CreateEventReq{
			{
				Name:       "领导1(只有用户)",
				Edit:       1,
				Refuse:     1,
				EditFields: "status",
				Users: []uint{
					4, 5, 6,
				},
			},
			{
				Name:   "领导2(用户和角色)",
				Edit:   0,
				Refuse: 1,
				Roles: []uint{
					4,
				},
				Users: []uint{
					8,
				},
			},
			{
				Name:   "领导3(只有角色)",
				Edit:   0,
				Refuse: 1,
				Roles: []uint{
					5,
				},
			},
		},
	})

	tx.Commit()
}

func TestFsm_SubmitLog(t *testing.T) {
	uid := "log1"
	tx := db.Begin()
	f := New(tx)
	_, err := f.SubmitLog(request.CreateLogReq{
		MId:             1,   // CreateMachine创建后生成的数据库id
		Category:        1,   // 自定义分类
		Uuid:            uid, // 唯一编号
		SubmitterUserId: 123, // 提交人id
	})
	if err != nil {
		fmt.Println(err)
	}

	tx.Commit()
}

func TestFsm_ApproveLog(t *testing.T) {
	uid := "log1"
	tx := db.Begin()
	f := New(tx)
	var err error
	// 通过
	_, err = f.ApproveLog(request.ApproveLogReq{
		MId:            1,   // CreateMachine创建后生成的数据库id
		Category:       1,   // 自定义分类
		Uuid:           uid, // 唯一编号
		ApprovalUserId: 4,
		Approved:       1,
	})
	if err != nil {
		fmt.Println(err)
	}
	// 拒绝
	_, err = f.ApproveLog(request.ApproveLogReq{
		MId:             1,   // CreateMachine创建后生成的数据库id
		Category:        1,   // 自定义分类
		Uuid:            uid, // 唯一编号
		ApprovalRoleId:  4,
		Approved:        2,
		ApprovalOpinion: "信息填错1",
	})
	if err != nil {
		fmt.Println(err)
	}
	// 拒绝
	_, err = f.ApproveLog(request.ApproveLogReq{
		MId:             1,   // CreateMachine创建后生成的数据库id
		Category:        1,   // 自定义分类
		Uuid:            uid, // 唯一编号
		ApprovalUserId:  4,
		Approved:        2,
		ApprovalOpinion: "信息填错2",
	})
	if err != nil {
		fmt.Println(err)
	}
	// 重新提交
	_, err = f.ApproveLog(request.ApproveLogReq{
		MId:            1,   // CreateMachine创建后生成的数据库id
		Category:       1,   // 自定义分类
		Uuid:           uid, // 唯一编号
		ApprovalUserId: 123,
		Approved:       1,
	})
	if err != nil {
		fmt.Println(err)
	}
	// 通过
	_, err = f.ApproveLog(request.ApproveLogReq{
		MId:             1,   // CreateMachine创建后生成的数据库id
		Category:        1,   // 自定义分类
		Uuid:            uid, // 唯一编号
		ApprovalUserId:  5,
		Approved:        1,
		ApprovalOpinion: "ok1",
	})
	if err != nil {
		fmt.Println(err)
	}
	// 通过
	_, err = f.ApproveLog(request.ApproveLogReq{
		MId:             1,   // CreateMachine创建后生成的数据库id
		Category:        1,   // 自定义分类
		Uuid:            uid, // 唯一编号
		ApprovalRoleId:  4,
		Approved:        1,
		ApprovalOpinion: "ok2",
	})
	if err != nil {
		fmt.Println(err)
	}
	// 通过
	_, err = f.ApproveLog(request.ApproveLogReq{
		MId:             1,   // CreateMachine创建后生成的数据库id
		Category:        1,   // 自定义分类
		Uuid:            uid, // 唯一编号
		ApprovalRoleId:  5,
		Approved:        1,
		ApprovalOpinion: "ok3",
	})
	if err != nil {
		fmt.Println(err)
	}
	// 提交人确认
	_, err = f.ApproveLog(request.ApproveLogReq{
		MId:            1,   // CreateMachine创建后生成的数据库id
		Category:       1,   // 自定义分类
		Uuid:           uid, // 唯一编号
		ApprovalUserId: 123,
		Approved:       1,
	})
	if err != nil {
		fmt.Println(err)
	}

	tx.Commit()
}

func TestFsm_ApproveLog1(t *testing.T) {
	uid := "log2"
	tx := db.Begin()
	f := New(tx)
	var err error
	_, err = f.SubmitLog(request.CreateLogReq{
		MId:             1,   // CreateMachine创建后生成的数据库id
		Category:        1,   // 自定义分类
		Uuid:            uid, // 唯一编号
		SubmitterUserId: 234, // 提交人id
	})
	if err != nil {
		fmt.Println(err)
	}

	// 其它人取消
	_, err = f.ApproveLog(request.ApproveLogReq{
		MId:            1,   // CreateMachine创建后生成的数据库id
		Category:       1,   // 自定义分类
		Uuid:           uid, // 唯一编号
		ApprovalUserId: 5,
		Approved:       3,
	})
	if err != nil {
		fmt.Println(err)
	}
	// 提交人取消
	_, err = f.ApproveLog(request.ApproveLogReq{
		MId:            1,   // CreateMachine创建后生成的数据库id
		Category:       1,   // 自定义分类
		Uuid:           uid, // 唯一编号
		ApprovalUserId: 234,
		Approved:       3,
	})
	if err != nil {
		fmt.Println(err)
	}

	_, err = f.SubmitLog(request.CreateLogReq{
		MId:             1,   // CreateMachine创建后生成的数据库id
		Category:        1,   // 自定义分类
		Uuid:            uid, // 唯一编号
		SubmitterRoleId: 567, // 提交角色id
		SubmitterUserId: 456, // 提交人id
	})
	if err != nil {
		fmt.Println(err)
	}

	// 提交角色取消
	_, err = f.ApproveLog(request.ApproveLogReq{
		MId:            1,   // CreateMachine创建后生成的数据库id
		Category:       1,   // 自定义分类
		Uuid:           uid, // 唯一编号
		ApprovalRoleId: 567,
		Approved:       3,
	})
	if err != nil {
		fmt.Println(err)
	}

	tx.Commit()
}

func TestFsm_CancelLogs(t *testing.T) {
	tx := db.Begin()
	f := New(tx)
	var err error
	_, err = f.SubmitLog(request.CreateLogReq{
		MId:             1,      // CreateMachine创建后生成的数据库id
		Category:        1,      // 自定义分类
		Uuid:            "log3", // 唯一编号
		SubmitterUserId: 123,    // 提交人id
	})
	if err != nil {
		fmt.Println(err)
	}
	_, err = f.SubmitLog(request.CreateLogReq{
		MId:             1,      // CreateMachine创建后生成的数据库id
		Category:        1,      // 自定义分类
		Uuid:            "log4", // 唯一编号
		SubmitterUserId: 234,    // 提交人id
	})
	if err != nil {
		fmt.Println(err)
	}
	_, err = f.SubmitLog(request.CreateLogReq{
		MId:             1,      // CreateMachine创建后生成的数据库id
		Category:        1,      // 自定义分类
		Uuid:            "log5", // 唯一编号
		SubmitterUserId: 345,    // 提交人id
	})
	if err != nil {
		fmt.Println(err)
	}

	f.CancelLogs(1)
	tx.Commit()
}

func TestFsm_FindPendingLogsByApprover(t *testing.T) {
	tx := db.Begin()
	f := New(tx)
	fmt.Println(f.FindPendingLogsByApprover(request.PendingLogReq{
		ApprovalRoleId: 1,
		ApprovalUserId: 2,
		Category:       1,
	}))
	tx.Commit()
}

func TestFsm_FindLogs(t *testing.T) {
	tx := db.Begin()
	f := New(tx)
	fmt.Println(f.FindLog(request.LogReq{
		Category: 1,
		Uuid:     "log1",
	}))
	tx.Commit()
}

func TestFsm_GetLogTrack(t *testing.T) {
	tx := db.Begin()
	f := New(tx)
	logs, _ := f.FindLog(request.LogReq{
		Category: 1,
		Uuid:     "log2",
	})
	fmt.Println(f.GetLogTrack(logs))
	tx.Commit()
}
