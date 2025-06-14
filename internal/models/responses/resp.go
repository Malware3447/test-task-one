package responses

import "time"

type Create struct {
	Id          int32     `json:"id"`
	ProjectId   int32     `json:"projectId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Priority    int32     `json:"priority"`
	Removed     bool      `json:"removed"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Update struct {
	Id          int32     `json:"id"`
	ProjectId   int32     `json:"projectId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Priority    int32     `json:"priority"`
	Removed     bool      `json:"removed"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Remove struct {
	Id         int32 `json:"id"`
	CampaignId int32 `json:"campaignId"`
	Removed    bool  `json:"removed"`
}

type Meta struct {
	Total   int32 `json:"total"`
	Removed int32 `json:"removed"`
	Limit   int32 `json:"limit"`
	Offset  int32 `json:"offset"`
}

type Goods struct {
	Id          int32     `json:"id"`
	ProjectId   int32     `json:"projectId"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Priority    int32     `json:"priority"`
	Romoved     bool      `json:"removed"`
	CreatedAt   time.Time `json:"createdAt"`
}

type List struct {
	Meta  Meta    `json:"meta"`
	Goods []Goods `json:"goods"`
}

type Priorities struct {
	Id       int32 `json:"id"`
	Priority int32 `json:"priority"`
}

type Reprioritize struct {
	Priorities []Priorities `json:"priorities"`
}
