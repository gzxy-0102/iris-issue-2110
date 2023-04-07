package request

type AddApiGroup struct {
	ProjectId uint64 `json:"project_id" binding:"required"`
	GroupName string `json:"group_name" binding:"required"`
}

type EditApiGroup struct {
	ID        uint64 `json:"id" binding:"required"`
	GroupName string `json:"group_name" binding:"required"`
}
