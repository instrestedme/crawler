package logic

import (
	"crawler/dao"
	"crawler/model"
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"strconv"
	"strings"
	"time"
)

// 定义用于解析返回json结构体
type RspData struct {
	Retcode int
	Result  `json:"result"`
}
type Result struct {
	Retcode          int
	Grade            int
	SpeCourseList    `json:"spe_course_list"`
	SysCoursePkgList []SysCourse `json:"sys_course_pkg_list"`
}

type SpeCourseList struct {
	Page  int
	Size  int
	Total int
	Data  []Course
}
type Course struct {
	Cid       int
	Name      string
	Grade     int
	Subject   int
	AfAmount  int       `json:"af_amount"`
	PreAmount int       `json:"pre_amount"`
	TeList    []Teacher `json:"te_list"`
}

type Teacher struct {
	Name string
}
type SysCourse struct {
	SubjectPackageId string `json:"subject_package_id"`
}

// 专题课
type SpecCourse struct {
	Grade   int
	Subject int
}

func (spe *SpecCourse) Work() {
	var (
		err       error
		getCourse model.Course
	)
	isEnd := false
	page := 1
	for !isEnd {
		specCourseList, sysCouseList := crawSinglePage(page, spe)
		// 去抓取系统课
		for _, sysCoursePkg := range sysCouseList {
			go func() {
				WorkPool.Run(&ServeSysCourse{
					SubjectPackageId: sysCoursePkg.SubjectPackageId,
				})
			}()
		}
		// 将专题课落库
		for _, speCourse := range specCourseList.Data {
			// 查询是否有此课程
			year, month, date := time.Now().Date()
			teStr := ""
			for k, te := range speCourse.TeList {
				if len(speCourse.TeList)-1 == k {
					teStr += te.Name
				} else {
					teStr += te.Name + ","
				}
			}
			err = dao.MasterDB.Model(model.Course{}).Where("id=?", speCourse.Cid).First(&getCourse).Error
			// 如果没有就创建
			if err == gorm.ErrRecordNotFound {
				newCourse := model.Course{
					Model: gorm.Model{
						ID: uint(speCourse.Cid),
					},
					BeginTime: uint(year*10000 + int(month)*100 + date),
					StopTime:  uint(year*10000 + int(month)*100 + date),
					Title:     speCourse.Name,
					Teacher:   teStr,
					PrePrice:  strconv.Itoa(speCourse.PreAmount),
					AfPrice:   strconv.Itoa(speCourse.AfAmount),
					Category:  1,
					Grade:     uint(speCourse.Grade),
					Subject:   uint(speCourse.Subject),
				}
				err = dao.MasterDB.Model(model.Course{}).Create(&newCourse).Error
				if err != nil {
					fmt.Println("插入课程失败：", err)
				}
				continue
			}
			if err != nil {
				fmt.Println("查询课程失败")
				continue
			}
			// 如果有就更新
			err = dao.MasterDB.Model(model.Course{}).Where("id=?", speCourse.Cid).Updates(model.Course{
				Model: gorm.Model{
					ID: uint(speCourse.Cid),
				},
				BeginTime: 0,
				StopTime:  uint(year*10000 + int(month)*100 + date),
				Title:     speCourse.Name,
				Teacher:   teStr,
				PrePrice:  strconv.Itoa(speCourse.PreAmount),
				AfPrice:   strconv.Itoa(speCourse.AfAmount),
				Category:  1,
				Grade:     uint(speCourse.Grade),
				Subject:   uint(speCourse.Cid),
			}).Error
			if err != nil {
				fmt.Println("更新专题课失败: ", err)
			}
		}
		if specCourseList.Page*specCourseList.Size >= specCourseList.Total {
			isEnd = true
		} else {
			page++
		}
	}
}

func crawSinglePage(page int, spe *SpecCourse) (*SpeCourseList, []SysCourse) {
	var (
		urlBuilder strings.Builder
		err        error
		rspData    RspData
	)
	cc := colly.NewCollector()
	cc.Limit(&colly.LimitRule{
		Parallelism: 1,
		RandomDelay: 5 * time.Second,
	})
	cc.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3")
		r.Headers.Set("Accept-Encoding", "gzip, deflate, br")
		r.Headers.Set("Accept-Language", "en,zh-CN;q=0.9,zh-TW;q=0.8,zh;q=0.7,ja;q=0.6")
		r.Headers.Set("Cache-Control", "max-age=0")
		r.Headers.Set("Cookie", viper.GetString("requestHeader.cookie"))
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
		r.Headers.Set("Referer", fmt.Sprintf(viper.GetString("crawl.gradeSubject.referer"), spe.Subject))
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.70 Safari/537.36")
	})
	cc.OnError(func(response *colly.Response, temerr error) {
		fmt.Println("获取课程数据失败")
		err = temerr
	})
	cc.OnResponse(func(response *colly.Response) {
		fmt.Println("专题课程返回了")
		err = json.Unmarshal(response.Body, &rspData)
		if err != nil {
			fmt.Println("解析课程数据错误")
			return
		}
		fmt.Println("打印专题课程", rspData)
	})
	urlBuilder.WriteString(viper.GetString("crawl.schema"))
	urlBuilder.WriteString("://")
	urlBuilder.WriteString(viper.GetString("crawl.website"))
	urlBuilder.WriteString("/")
	urlBuilder.WriteString(viper.GetString("crawl.getCourses.url"))
	urlBuilder.WriteString("?client=")
	urlBuilder.WriteString(viper.GetString("crawl.getCourses.client"))
	urlBuilder.WriteString("&platform=")
	urlBuilder.WriteString(viper.GetString("crawl.getCourses.platform"))
	urlBuilder.WriteString("&version=")
	urlBuilder.WriteString(viper.GetString("crawl.getCourses.version"))
	urlBuilder.WriteString("&grade=")
	urlBuilder.WriteString(strconv.Itoa(spe.Grade))
	urlBuilder.WriteString("&subject=")
	urlBuilder.WriteString(strconv.Itoa(spe.Subject))
	urlBuilder.WriteString("&showid=")
	urlBuilder.WriteString(viper.GetString("crawl.getCourses.showid"))
	urlBuilder.WriteString("&page=")
	urlBuilder.WriteString(strconv.Itoa(page))
	urlBuilder.WriteString("&size=")
	urlBuilder.WriteString(viper.GetString("crawl.getCourses.size"))
	cc.Visit(urlBuilder.String())
	return &rspData.SpeCourseList, rspData.SysCoursePkgList
}
