package model

import "github.com/jinzhu/gorm"

type GradeSubjects struct {
	gorm.Model
	Grade   uint   `gorm:"type:int;not null;unique_index;default:0;comment:'年级'"`
	Subject string `gorm:"type:varchar(1024);not null; default:'';comment:'科目'"`
}
