package request

type base struct {
	IP       string `json:"ip" binding:"required,ip"`
	Port     uint64 `json:"port" binding:"required"`
	Database string `json:"database" binding:"required"`
	User     string `json:"user" binding:"required"`
	Password string `json:"password" binding:"required"`
	Device   string `json:"device" binding:"required,oneof=mysql pgsql"`
	Charset  string `json:"charset" binding:"required"`
}
type AddSource struct {
	ProjectId  uint64 `json:"project_id" binding:"required"`
	SourceName string `json:"source_name" binding:"required"`
	base
}

type TestSource struct {
	base
}

type EditSource struct {
	ID         uint64 `json:"id" binding:"required"`
	SourceName string `json:"source_name" binding:"required"`
	base
}
