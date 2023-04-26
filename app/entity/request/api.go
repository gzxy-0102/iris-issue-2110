package request

import (
	"2110/app/entity/common"
)

type AddApi struct {
	ProjectId uint64 `json:"project_id" binding:"required"`
	GroupId   uint64 `json:"group_id" binding:"required"`
	SourceId  uint64 `json:"source_id" binding:"required"`
	ApiName   string `json:"api_name" binging:"required"`
}

type EditApi struct {
	ID       uint64           `json:"id" binding:"required"`
	GroupId  uint64           `json:"group_id" binding:"required"`
	SourceId uint64           `json:"source_id" binding:"required"`
	ApiName  string           `json:"api_name" binding:"required"`
	Payload  string           `json:"payload" binding:"required"`
	State    common.StateEnum `json:"state" binding:"required,oneof=0 1 2"`
}

type CheckApiPayload struct {
	Payload  string         `json:"payload" binding:"required"`
	SourceId uint64         `json:"source_id" binding:"required"`
	Params   map[string]any `json:"params"`
}
