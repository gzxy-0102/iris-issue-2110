package request

type AddProject struct {
	Name string `json:"name" binding:"required"`
}

type EditProject struct {
	ID   uint64 `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}
