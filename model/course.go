package model

import "github.com/jinzhu/gorm"

type Course struct {
	gorm.Model
	BeginTime uint   `gorm:"index:time;not null;default:0;comment:'开始记录日期'" json:"begin_time"`
	StopTime  uint   `gorm:"index:time;not null;default:0;comment:'结束记录日期'" json:"stop_time"`
	Title     string `gorm:"type:varchar(150);not null;default:'';comment:'课程标题'" json:"title"`
	Teacher   string `gorm:"type:varchar(100);not null;default:'';comment:'老师信息'" json:"teacher"`
	PrePrice  string `gorm:"type:varchar(20);not null;default:'';comment:'原价'" json:"pre_price"`
	AfPrice   string `gorm:"type:varchar(20);not null;default:'';comment:'现价'" json:"af_price"`
	Category  int    `gorm:"index:category;not null;default:0;comment:'分类：1-专题课程 2-系统课程'" json:"category"`
	Grade     uint   `gorm:"index:grade;not null; default:0;comment:'年级'" json:"grade"`
	Subject   uint   `gorm:"index:subject;not null;default:0;comment:'学科'" json:"subject"`
}
