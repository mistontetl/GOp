package models

type AppCounter struct {
	ID              uint `gorm:"column:id;primaryKey"`
	GlobalCounter   int  `gorm:"column:global_counter"`
	PersonalCounter int  `gorm:"column:personal_counter"`
}

func (AppCounter) TableName() string {
	return "app_counters"
}
