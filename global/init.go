package global

import (
	"crawler/dao"
	"crawler/logic"
	"crawler/model"
	"crawler/util"
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var once = new(sync.Once)

var (
	WhichSite = flag.String("site", "", "要抓取的网址")
	Config    = flag.String("config", "config", "环境变量档案名称，默认 config")
)

func Init() {
	once.Do(func() {
		if !flag.Parsed() {
			flag.Parse()
		}
		rand.Seed(time.Now().UnixNano())

		viper.SetConfigName(*Config)
		viper.AddConfigPath(inferRootDir() + "/config")
		err := viper.ReadInConfig()
		if err != nil {
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
		}
		dao.Init()
		// 初始化工作池
		logic.InitPool()
		// 自动建表
		dao.MasterDB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&model.GradeSubjects{}, &model.Course{})
	})
}

// 推断配置文件的目录
func inferRootDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	var infer func(d string) string
	infer = func(d string) string {
		if util.Exist(d + "/config") {
			return d
		}

		return infer(filepath.Dir(d))
	}

	return infer(cwd)
}
