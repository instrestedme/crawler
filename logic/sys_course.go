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

// 解析json数据用
type SysRspData struct {
	Retcode int
	Result  ParseSysResult `json:"result"`
}

type ParseSysResult struct {
	Retcode int
	Courses []ParseSysCourse
}
type ParseSysCourse struct {
	Cid       uint
	Name      string
	Grade     uint
	Subject   uint      `json:"subject"`
	PreAmount uint      `json:"pre_amount"`
	AfAmount  uint      `json:"af_amount"`
	TeList    []Teacher `json:"te_list"`
}

// Request URL: https://fudao.qq.com/cgi-proxy/course/get_course_package_info?client=4&platform=3&version=30&subject_package_id=str_sys_course_pkg_info_1_6002_8_0&t=0.8308425060287739
// 系统课
type ServeSysCourse struct {
	SubjectPackageId string `json:"subject_package_id"`
}

func (sys *ServeSysCourse) Work() {
	var (
		urlBuilder strings.Builder
		err        error
		rspData    SysRspData
		getCourse  model.Course
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
		r.Headers.Set("Referer", fmt.Sprintf(viper.GetString("crawl.getSysCourses.referer"), sys.SubjectPackageId))
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.70 Safari/537.36")
	})
	cc.OnError(func(response *colly.Response, temerr error) {
		fmt.Println("获取系统课程数据失败")
		err = temerr
	})
	cc.OnResponse(func(response *colly.Response) {
		fmt.Println("系统课程返回了")
		err = json.Unmarshal(response.Body, &rspData)
		if err != nil {
			fmt.Println("解析系统课程数据错误")
			return
		}
		fmt.Println("打印系统课程：", rspData)
	})
	urlBuilder.WriteString(viper.GetString("crawl.schema"))
	urlBuilder.WriteString("://")
	urlBuilder.WriteString(viper.GetString("crawl.website"))
	urlBuilder.WriteString("/")
	urlBuilder.WriteString(viper.GetString("crawl.getSysCourses.url"))
	urlBuilder.WriteString("?client=")
	urlBuilder.WriteString(viper.GetString("crawl.getSysCourses.client"))
	urlBuilder.WriteString("&platform=")
	urlBuilder.WriteString(viper.GetString("crawl.getSysCourses.platform"))
	urlBuilder.WriteString("&version=")
	urlBuilder.WriteString(viper.GetString("crawl.getSysCourses.version"))
	urlBuilder.WriteString("&subject_package_id=")
	urlBuilder.WriteString(sys.SubjectPackageId)
	cc.Visit(urlBuilder.String())
	// 将系统课落库
	for _, sysCorse := range rspData.Result.Courses {
		// 查询是否有此课程
		year, month, date := time.Now().Date()
		teStr := ""
		for k, te := range sysCorse.TeList {
			if len(sysCorse.TeList)-1 == k {
				teStr += te.Name
			} else {
				teStr += te.Name + ","
			}
		}
		err = dao.MasterDB.Model(model.Course{}).Where("id=?", sysCorse.Cid).First(&getCourse).Error
		// 如果没有就创建
		if err == gorm.ErrRecordNotFound {
			newCourse := model.Course{
				Model: gorm.Model{
					ID: sysCorse.Cid,
				},
				BeginTime: uint(year*10000 + int(month)*100 + date),
				StopTime:  uint(year*10000 + int(month)*100 + date),
				Title:     sysCorse.Name,
				Teacher:   teStr,
				PrePrice:  strconv.Itoa(int(sysCorse.PreAmount)),
				AfPrice:   strconv.Itoa(int(sysCorse.AfAmount)),
				Category:  2,
				Grade:     sysCorse.Grade,
				Subject:   sysCorse.Subject,
			}
			err = dao.MasterDB.Model(model.Course{}).Create(&newCourse).Error
			if err != nil {
				fmt.Println("插入系统课程课程失败：", err)
			}
			continue
		}
		if err != nil {
			fmt.Println("查询系统课程失败")
			continue
		}
		// 如果有就更新
		err = dao.MasterDB.Model(model.Course{}).Where("id=?", sysCorse.Cid).Updates(model.Course{
			Model: gorm.Model{
				ID: sysCorse.Cid,
			},
			BeginTime: 0,
			StopTime:  uint(year*10000 + int(month)*100 + date),
			Title:     "",
			Teacher:   teStr,
			PrePrice:  strconv.Itoa(int(sysCorse.PreAmount)),
			AfPrice:   strconv.Itoa(int(sysCorse.AfAmount)),
			Category:  2,
			Grade:     sysCorse.Grade,
			Subject:   sysCorse.Subject,
		}).Error
		if err != nil {
			fmt.Println("更新专题课失败: ", err)
		}
	}
}

func crawSingleSysPage(page int, spe *SpecCourse) (*SpeCourseList, []SysCourse) {
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
		fmt.Println("返回了")
		err = json.Unmarshal(response.Body, &rspData)
		if err != nil {
			fmt.Println("解析课程数据错误")
			return
		}
		fmt.Println(rspData)
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
