package logic

import (
	"github.com/spf13/viper"
)

var (
	WorkPool *Pool // 工作池
)

func InitPool() {
	WorkPool = NewPool(viper.GetInt("worker.parallel"))
}
