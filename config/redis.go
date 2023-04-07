package config

import "time"

type RedisConfiguration struct {
	Enable             bool          `yaml:"Enable"`             // 是否启用Redis
	Addr               string        `yaml:"Addr"`               // redis地址，格式 host:port
	Username           string        `yaml:"Username"`           // redis用户, redis server没有设置可以忽略
	Password           string        `yaml:"Password"`           // redis密码，redis server没有设置可以忽略
	Database           int           `yaml:"Database"`           // redis数据库，序号从0开始，默认是0，可以忽略
	MaxRetries         int           `yaml:"MaxRetries"`         // redis操作失败最大重试次数，默认不重试。
	MinRetryBackoff    time.Duration `yaml:"MinRetryBackoff"`    // 最小重试时间间隔. 默认是 8ms ; -1 表示关闭.
	MaxRetryBackoff    time.Duration `yaml:"MaxRetryBackoff"`    // 最大重试时间间隔 默认是 512ms; -1 表示关闭.
	DialTimeout        time.Duration `yaml:"DialTimeout"`        // redis连接超时时间. 默认5秒
	ReadTimeout        time.Duration `yaml:"ReadTimeout"`        // 读取超时时间 默认3秒
	WriteTimeout       time.Duration `yaml:"WriteTimeout"`       // 写超时时间
	PoolSize           int           `yaml:"PoolSize"`           // redis连接池的最大连接数. 默认连接池大小等于 cpu个数 * 10
	MinIdleConns       int           `yaml:"MinIdleConns"`       // redis连接池最小空闲连接数.
	MaxConnAge         time.Duration `yaml:"MaxConnAge"`         // redis连接最大的存活时间，默认不会关闭过时的连接.
	PoolTimeout        time.Duration `yaml:"PoolTimeout"`        // 当你从redis连接池获取一个连接之后，连接池最多等待这个拿出去的连接多长时间。 默认是等待 ReadTimeout + 1 秒.
	IdleTimeout        time.Duration `yaml:"IdleTimeout"`        // redis连接池多久会关闭一个空闲连接. 默认是 5 分钟. -1 则表示关闭这个配置项
	IdleCheckFrequency time.Duration `yaml:"IdleCheckFrequency"` // 多长时间检测一下，空闲连接 默认是 1 分钟. -1 表示关闭空闲连接检测
}
