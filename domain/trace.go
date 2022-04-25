package domain

type Trace struct {
	ID       uint `gorm:"primaryKey"`
	Strategy string
}
