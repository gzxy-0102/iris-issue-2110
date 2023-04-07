package dynamicSource

import (
	"2110/app/model"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/xwb1989/sqlparser"
	"gorm.io/gorm"
	"strings"
	"sync"
	"time"
)

var RDB sync.Map

type Manager interface {
	Register() error
	UnRegister() error
	Test() bool
	AnalysisDataBase() ([]DBInformation, error)
	AnalysisTable(table string) ([]TableInformation, error)
	TableExists(table string) (bool, error)
}

type RDBManager struct {
	DB       *gorm.DB  //	数据库实例
	id       uint64    //	数据源ID
	DSN      string    //	数据源地址
	Errors   []error   //	链接以来执行错误记录
	Device   string    //	驱动类型
	LastTime time.Time //	上次使用时间
}

// RegisterSource 注册数据源
func RegisterSource(source model.Source) error {
	log.Infof("正在注册数据源：%s", source.SourceName)
	var err error
	switch source.Device {
	case "mysql":
		rdb := RDBMysql{
			Source: source,
		}
		err = rdb.Register()
		break
	}
	return err
}

// UNRegisterSource 取消数据源注册
func UNRegisterSource(source model.Source) error {
	log.Infof("正在取消注册数据源：%s", source.SourceName)
	var err error
	switch source.Device {
	case "mysql":
		rdb := RDBMysql{
			Source: source,
		}
		err = rdb.UnRegister()
		break
	}
	return err
}

func TestSource(source model.Source) bool {
	var success bool
	switch source.Device {
	case "mysql":
		rdb := RDBMysql{Source: source}
		success = rdb.Test()
	}
	return success
}

func Ping(source model.Source) (bool, []error) {
	value, ok := RDB.Load(source.ID)
	if ok {
		manager := value.(*RDBManager)
		return true, manager.Errors
	}
	return false, []error{
		errors.New("数据源未注册"),
	}
}

func getManagerBySource(source model.Source) *RDBManager {
	value, ok := RDB.Load(source.ID)
	if ok {
		manager := value.(*RDBManager)
		return manager
	}
	//	当数据源池中不存在时 尝试进行注册一次避免因不活跃被销毁
	err := RegisterSource(source)
	if err != nil {
		log.Errorf("注册数据源失败：%v", source)
		return nil
	}
	value, ok = RDB.Load(source.ID)
	if ok {
		manager := value.(*RDBManager)
		return manager
	}
	return nil
}

func QueryBySource(source model.Source, sql string, param any) (error, any) {
	manager := getManagerBySource(source)
	log.Infof("Manager: %+v  %v SQL %s", manager, manager == nil, sql)
	if manager != nil {
		stmt, err := sqlparser.Parse(sql)
		if err != nil {
			updateRDB(source.ID, manager)
			return err, nil
		}
		switch stmt.(type) {
		case *sqlparser.Select:
			var queryResult []map[string]any
			var result *gorm.DB
			if param == nil {
				result = manager.DB.Raw(sql).Scan(&queryResult)
			} else {
				switch param.(type) {
				case []any:
					result = manager.DB.Raw(sql, param.([]any)...).Scan(&queryResult)
					break
				case map[string]any:
					result = manager.DB.Raw(sql, param).Scan(&queryResult)
					break
				default:
					updateRDB(source.ID, manager)
					return errors.New("错误的SQL参数"), nil
				}
			}
			if result.Error != nil {
				manager.Errors = append(manager.Errors, result.Error)
				updateRDB(source.ID, manager)
				return result.Error, nil
			}
			updateRDB(source.ID, manager)
			return nil, queryResult
		case *sqlparser.Insert, *sqlparser.Update, *sqlparser.Delete:
			var result *gorm.DB
			if param == nil {
				result = manager.DB.Exec(sql)
			} else {
				switch param.(type) {
				case []any:
					result = manager.DB.Exec(sql, param.([]any)...)
					break
				case map[string]any:
					result = manager.DB.Exec(sql, param)
					break
				default:
					updateRDB(source.ID, manager)
					return errors.New("错误的SQL参数"), nil
				}
			}
			if result.Error != nil {
				manager.Errors = append(manager.Errors, result.Error)
				updateRDB(source.ID, manager)
				return result.Error, false
			}
			updateRDB(source.ID, manager)
			return nil, true
		default:
			var result *gorm.DB
			if strings.HasPrefix(sql, "EXPLAIN") || strings.HasPrefix(sql, "explain") {
				var queryResult []map[string]any
				if param == nil {
					result = manager.DB.Raw(sql).Scan(&queryResult)
				} else {
					switch param.(type) {
					case []any:
						result = manager.DB.Raw(sql, param.([]any)...).Scan(&queryResult)
						break
					case map[string]any:
						result = manager.DB.Raw(sql, param).Scan(&queryResult)
						break
					default:
						updateRDB(source.ID, manager)
						return errors.New("错误的SQL参数"), nil
					}
				}
				if result.Error != nil {
					manager.Errors = append(manager.Errors, result.Error)
					updateRDB(source.ID, manager)
					return result.Error, nil
				}
				updateRDB(source.ID, manager)
				return nil, queryResult
			}
		}
	}
	return errors.New("数据源连接失败"), nil
}

func TableExists(source model.Source, tableName string) (bool, error) {
	var exists bool
	var err error
	switch source.Device {
	case "mysql":
		rdb := RDBMysql{
			Source: source,
		}
		exists, err = rdb.TableExists(tableName)
	}
	return exists, err
}

type DBInformation struct {
	TableName      string          `json:"table_name"`    //表名称
	Engine         string          `json:"engine"`        //表引擎
	TableComment   string          `json:"table_comment"` //表注释
	TableCollation string          `json:"table_collation"`
	CreateTime     model.LocalTime `json:"create_time"`
}

func AnalysisBySource(source model.Source) ([]DBInformation, error) {
	var dbInformation []DBInformation
	var err error
	switch source.Device {
	case "mysql":
		rdb := RDBMysql{Source: source}
		dbInformation, err = rdb.AnalysisDataBase()
	}

	return dbInformation, err
}

type TableInformation struct {
	ColumnName    string
	DataType      string
	ColumnComment string
	ColumnKey     string
	Extra         string
	IsNullable    string
	ColumnType    string
}

func AnalysisByTable(source model.Source, table string) ([]TableInformation, error) {
	var tableInformation []TableInformation
	var err error
	switch source.Device {
	case "mysql":
		rdb := RDBMysql{Source: source}
		tableInformation, err = rdb.AnalysisTable(table)
	}
	return tableInformation, err
}

func updateRDB(sourceId uint64, manager *RDBManager) {
	manager.LastTime = time.Now()
	RDB.Store(sourceId, manager)
}
