package dao

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/spf13/viper"
)

var MasterDB *gorm.DB

func Init() {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		viper.GetString("storage.user"),
		viper.GetString("storage.password"),
		viper.GetString("storage.host"),
		viper.GetString("storage.port"),
		viper.GetString("storage.dbname"),
		viper.GetString("storage.charset"))
	MasterDB, err = gorm.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
}
