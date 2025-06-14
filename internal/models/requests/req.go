package requests

type Create struct {
	Name string `json:"name"`
}

type Update struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type Reprioritize struct {
	NewPriority int32 `json:"newPriority"`
}
