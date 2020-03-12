package subjects

import (
	"crawler/dao"
	"crawler/model"
	"fmt"
	"github.com/gin-gonic/gin"
)

type ListService struct {
}

type RspData struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data []int  `json:"data"`
}

func List(c *gin.Context) {
	var (
		err         error
		subjectList []int
	)
	rows, err := dao.MasterDB.Model(&model.Course{}).Select(`distinct subject`).Rows()
	if err != nil {
		fmt.Println("查询科目失败")
		c.Err()
	}
	if err != nil {
		fmt.Println("查询科目失败")
		c.Err()
	}
	for rows.Next() {
		var temData int
		rows.Scan(&temData)
		subjectList = append(subjectList, temData)
	}
	c.JSON(200, RspData{
		Code: 0,
		Msg:  "成功",
		Data: subjectList,
	})
}
