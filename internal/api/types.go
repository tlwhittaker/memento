package api

import "time"

const (
	RowStatusNormal   = "NORMAL"
	RowStatusArchived = "ARCHIVED"
)

const (
	VisibilityPrivate   = "PRIVATE"
	VisibilityProtected = "PROTECTED"
	VisibilityPublic    = "PUBLIC"
)

type Memo struct {
	Name        string     `json:"name"`
	UID         string     `json:"uid"`
	Content     string     `json:"content"`
	CreateTime  time.Time  `json:"createTime"`
	UpdateTime  time.Time  `json:"updateTime"`
	DisplayTime time.Time  `json:"displayTime"`
	RowStatus   string     `json:"rowStatus"`
	Visibility  string     `json:"visibility"`
	Pinned      bool       `json:"pinned"`
	Creator     string     `json:"creator"`
	Tags        []string   `json:"tags,omitempty"`
	Resources   []Resource `json:"resources,omitempty"`
	Relations   []Relation `json:"relations,omitempty"`
}

type Resource struct {
	Name         string    `json:"name"`
	UID          string    `json:"uid"`
	Filename     string    `json:"filename"`
	Type         string    `json:"type"`
	Size         int64     `json:"size"`
	CreateTime   time.Time `json:"createTime"`
	ExternalLink string    `json:"externalLink,omitempty"`
}

type Relation struct {
	Memo        string `json:"memo"`
	RelatedMemo string `json:"relatedMemo"`
	Type        string `json:"type"`
}

type ListMemosResponse struct {
	Memos         []Memo `json:"memos"`
	NextPageToken string `json:"nextPageToken,omitempty"`
}

type CreateMemoRequest struct {
	Content    string `json:"content"`
	Visibility string `json:"visibility,omitempty"`
}

type UpdateMemoRequest struct {
	Content    string `json:"content,omitempty"`
	RowStatus  string `json:"rowStatus,omitempty"`
	Visibility string `json:"visibility,omitempty"`
	Pinned     *bool  `json:"pinned,omitempty"`
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return e.Message
}
