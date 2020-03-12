package course

import (
	"crawler/dao"
	"crawler/model"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type ListService struct {
	Page    int  `validate:"gt=0" form:"page"`
	Size    int  `validate:"gt=0" form:"size"`
	Date    int  `form:"date"`
	Subject uint `form:"subject"`
}

type RspData struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data RspExecData `json:"data"`
}

type RspExecData struct {
	Page  int            `json:"page"`
	Size  int            `json:"size"`
	Total int            `json:"total"`
	Data  []model.Course `json:"data"`
}

func List(c *gin.Context) {
	var (
		err        error
		sqlStr     []string
		courseList []model.Course
		count      int
	)
	service := ListService{}
	err = c.ShouldBindQuery(&service)
	if err != nil {
		fmt.Println("参数错误：", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	argSlice := []interface{}{}
	sqlStr = []string{}
	if service.Date > 0 {
		sqlStr = append(sqlStr, `stop_time=?`)
		argSlice = append(argSlice, service.Date)
	}
	if service.Subject > 0 {
		sqlStr = append(sqlStr, `subject=?`)
		argSlice = append(argSlice, int(service.Subject))
	}
	if service.Page > 0 && service.Size > 0 {
		if len(argSlice) > 0 {
			err = dao.MasterDB.Model(model.Course{}).Offset((service.Page-1)*service.Size).Where(strings.Join(sqlStr, ` AND `), argSlice...).Limit(service.Size).Find(&courseList).Count(&count).Error
		} else {
			err = dao.MasterDB.Model(model.Course{}).Offset((service.Page - 1) * service.Size).Limit(service.Size).Find(&courseList).Count(&count).Error
		}
		if err != nil {
			if err == sql.ErrNoRows {
				c.AbortWithStatusJSON(400, gin.H{"code": 1, "msg": "没有更多数据"})
			}
			fmt.Println("查询列表失败: ", err)
			c.Err()
			return
		}

		c.JSON(200, RspData{
			Code: 0,
			Msg:  "成功",
			Data: RspExecData{
				Page:  service.Page,
				Size:  service.Size,
				Total: count,
				Data:  courseList,
			},
		})
		//c.JSON(200,gin.H{"msg":"获取成功","data":courseList})
	}

}
