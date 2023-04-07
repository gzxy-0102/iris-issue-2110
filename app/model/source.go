package model

import "fmt"

type Source struct {
	Base
	SourceMark string `json:"source_mark" gorm:"column:source_mark;type:varchar(100);not null;default:'';comment:数据源标识"`
	ProjectId  uint64 `json:"project_id" gorm:"column:project_id;type:bigint(20);not null;default:0;comment:所属项目"`
	UserId     uint64 `gorm:"column:user_id;type:bigint(20);not null;default:0;comment:所属用户" json:"user_id"`
	SourceName string `json:"source_name" gorm:"column:source_name;type:varchar(60);not null;default:'';comment:数据源名称"`
	Device     string `gorm:"column:device;type:varchar(100);not null;default:'';comment:驱动类型" json:"device"`
	IP         string `gorm:"column:ip;type:varchar(100);not null;default:'';comment:IP地址" json:"ip"`
	Port       uint64 `gorm:"column:port;type:bigint(20);not null;default:0;comment:端口" json:"port"`
	Database   string `gorm:"column:database;type:varchar(100);not null;default:'';comment:数据库名称" json:"database"`
	Charset    string `gorm:"column:charset;type:varchar(100);not null;default:utf8mb4;comment:数据库编码" json:"charset"`
	User       string `gorm:"column:user;type:varchar(100);not null;default:'';comment:用户名" json:"user"`
	Password   string `gorm:"column:password;type:varchar(100);not null;default:'';comment:密码" json:"password"`
}

func (source *Source) Dsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", source.User, source.Password, source.IP, source.Port, source.Database, source.Charset)
}

func (source *Source) TableName() string {
	return "source"
}
