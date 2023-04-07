package dynamicSource

import (
	"2110/app/model"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type RDBMysql struct {
	Source model.Source
}

func (d *RDBMysql) Register() error {
	db, err := gorm.Open(mysql.Open(d.Source.Dsn()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	var manager RDBManager
	manager.DB = db
	manager.id = d.Source.ID
	manager.DSN = d.Source.Dsn()
	if err != nil {
		manager.Errors = append(manager.Errors, err)
	}
	manager.Device = d.Source.Device
	updateRDB(d.Source.ID, &manager)
	return err
}

func (d *RDBMysql) UnRegister() error {
	value, ok := RDB.Load(d.Source.ID)
	if ok {
		manage := value.(*RDBManager)
		db, err := manage.DB.DB()
		if err != nil {
			log.Errorf("获取DB错误：%v", err)
		} else {
			_ = db.Close()
		}
	}
	RDB.Delete(d.Source.ID)
	return nil
}

func (d *RDBMysql) Test() bool {
	orm, err := gorm.Open(mysql.Open(d.Source.Dsn()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return false
	}
	db, _ := orm.DB()
	_ = db.Close()
	return true
}

func (d *RDBMysql) AnalysisDataBase() ([]DBInformation, error) {
	manager := getManagerBySource(d.Source)
	if manager != nil {
		var dbInformation []DBInformation
		var err error
		result := manager.DB.Raw("select table_name,engine,table_comment,table_collation,create_time from information_schema.tables where table_schema = (select database()) order by create_time desc ").Scan(&dbInformation)
		err = result.Error
		if err != nil {
			manager.Errors = append(manager.Errors, err)
		}
		updateRDB(d.Source.ID, manager)
		return dbInformation, err
	}
	return nil, errors.New("获取DB错误")
}

func (d *RDBMysql) AnalysisTable(table string) ([]TableInformation, error) {
	manager := getManagerBySource(d.Source)
	if manager != nil {
		var tableInformation []TableInformation
		var err error
		sql := fmt.Sprintf("select column_name,data_type,column_comment,column_key, extra ,is_nullable ,column_type from information_schema.columns where table_name = '%s' and table_schema = (select database())", table)
		result := manager.DB.Raw(sql).Scan(&tableInformation)
		err = result.Error
		if err != nil {
			manager.Errors = append(manager.Errors, err)
		}
		updateRDB(d.Source.ID, manager)
		return tableInformation, err
	}
	return nil, errors.New("获取DB错误")
}

func (d *RDBMysql) TableExists(table string) (bool, error) {
	manager := getManagerBySource(d.Source)
	if manager != nil {
		var exists uint
		result := manager.DB.Raw("select count(*) from information_schema.TABLES where TABLE_NAME = ?", table).Scan(&exists)
		if result.Error != nil {
			manager.Errors = append(manager.Errors, result.Error)
		}
		updateRDB(d.Source.ID, manager)
		return exists > 0, result.Error
	}
	return false, errors.New("数据源未注册")
}
