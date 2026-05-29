package response

type TomoriStats struct {
	UserCount    int64 `json:"user_count"`
	PostCount    int64 `json:"post_count"`
	CommentCount int64 `json:"comment_count"`
	FileCount    int64 `json:"file_count"`
	VideoCount   int64 `json:"video_count"`
	AgentCount   int64 `json:"agent_count"`
	ChatCount    int64 `json:"chat_count"`
}

type TomoriUserList struct {
	Items    []UserInfo `json:"items"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
	Total    int64      `json:"total"`
}

type TomoriBlacklistItem struct {
	ID        string `json:"id"`
	Token     string `json:"token"`
	CreatedAt string `json:"created_at"`
}

type TomoriBlacklist struct {
	Items    []TomoriBlacklistItem `json:"items"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
	Total    int64                 `json:"total"`
}

type TomoriAgentInfo struct {
	AgentId     uint   `json:"agent_id"`
	UserId      uint   `json:"user_id"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	ModelName   string `json:"model_name"`
	Provider    string `json:"provider"`
	Activate    bool   `json:"activate"`
	CreatedAt   string `json:"created_at"`
}

type TomoriAgentList struct {
	Items    []TomoriAgentInfo `json:"items"`
	Total    int64             `json:"total"`
}
