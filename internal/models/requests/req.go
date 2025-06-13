package requests

type Create struct {
	Name string `json:"name"`
}

type Update struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type Remove struct {
	Id         int32 `json:"id"`
	CompaignId int32 `json:"campaignId"`
	Removed    bool  `json:"removed"`
}

type Reprioritize struct {
	NewPriority int32 `json:"newPriority"`
}
