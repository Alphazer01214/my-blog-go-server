package entity

type Notification struct {
	UserId uint `json:"user_id" gorm:"primaryKey"`
}
