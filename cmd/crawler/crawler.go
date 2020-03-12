package main

import (
	"crawler/dao"
	"crawler/logic"
	"crawler/model"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/robfig/cron"
	"github.com/spf13/viper"
	"strconv"
	"strings"
	"time"
)

var (
	GradeSubCache []GradeSubject //返回的年级-科目缓存
)

// 解析返回json的结构体
type GradeSubject struct {
	Grade   int
	Subject []int
}

type Result struct {
	GradeSubjects []GradeSubject `json:"grade_subjects"`
}

type GradeSubRspData struct {
	RetCode int
	Result  `json:"result"`
}

func ServeCrawle() {
	var (
		startSubject []int //抓取首页subjectId用作请求grade_subjects的启动referer
		err          error
	)

	//// 定时任务
	c := cron.New()
	//防止没有配置报错，设置默认值
	viper.SetDefault("crawl.spec", "* * * */1 * *")
	c.AddFunc(viper.GetString("crawl.spec"), func() {
		startSubject = CrawlSubjects()
		fmt.Println("打印科目id", startSubject)
		for _, subjectId := range startSubject {
			err = CrawlGradeSubjects(subjectId)
			if err == nil {
				break
			} else {
				fmt.Println("获取年级科目信息出错：", err)
				continue
			}
		}
		// 遍历年级科目组合获取课程，这里可以并发抓取，考虑到一天只有一次，且会被封ip，所以设置的工作线程一个
		// 检查是否有缓存，没有从数据库查
		if len(GradeSubCache) == 0 {
			err = dao.MasterDB.Model(model.GradeSubjects{}).Find(&GradeSubject{}).Error
			if err != nil {
				panic(errors.New("查询年级-科目出错"))
			}
		}
		for _, gradeSub := range GradeSubCache {
			for _, subId := range gradeSub.Subject {
				logic.WorkPool.Run(&logic.SpecCourse{
					Grade:   gradeSub.Grade,
					Subject: subId,
				})
			}
		}
	})
	c.Start()
}

// 抓取grade_subjects数据
func CrawlGradeSubjects(subjectId int) error {
	var (
		err     error
		rspData GradeSubRspData
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
		r.Headers.Set("Referer", fmt.Sprintf(viper.GetString("crawl.gradeSubject.referer"), subjectId))
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.70 Safari/537.36")
	})
	cc.OnResponse(func(response *colly.Response) {
		err = json.Unmarshal(response.Body, &rspData)
		if err != nil {
			fmt.Println("解析失败：", err)
			return
		}
		// 缓存
		GradeSubCache = rspData.Result.GradeSubjects
		// 批量插入年级——科目数据
		for _, gradeSubject := range rspData.Result.GradeSubjects {
			var (
				sliceStr strings.Builder
				getCount int
			)
			for k, v := range gradeSubject.Subject {
				if len(gradeSubject.Subject)-1 == k {
					sliceStr.WriteString(strconv.Itoa(v))
				} else {
					sliceStr.WriteString(strconv.Itoa(v) + ",")
				}
			}
			err = dao.MasterDB.Model(model.GradeSubjects{}).Where("grade=?", gradeSubject.Grade).Count(&getCount).Error
			if err != nil {
				fmt.Println("查询年级-科目失败")
				return
			}
			// 如果有就更新
			if getCount > 0 {
				err = dao.MasterDB.Model(model.GradeSubjects{}).Where("grade=?", gradeSubject.Grade).Updates(map[string]interface{}{"subject": sliceStr.String()}).Error
				if err != nil {
					fmt.Println("更新年级-科目失败")
				}
				return
			}
			// 插入
			err = dao.MasterDB.Model(model.GradeSubjects{}).Create(&model.GradeSubjects{Grade: uint(gradeSubject.Grade), Subject: sliceStr.String()}).Error
			if err != nil {
				fmt.Println("插入失败")
				return
			}
		}
		if err != nil {
			fmt.Println("批量插入年级-科目数据失败")
			return
		}
	})
	cc.OnError(func(response *colly.Response, temerr error) {
		fmt.Println("请求年级科目信息出错：", temerr)
		err = temerr
		return
	})
	cc.Visit(viper.GetString("crawl.schema") + "://" + viper.GetString("crawl.website") + "/" + viper.GetString("crawl.gradeSubject.url"))
	return err
}

// 抓取科目,用作请求grade_subject的referer的启动点
func CrawlSubjects() []int {
	var (
		subjectId []int
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
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.70 Safari/537.36")
	})
	cc.OnHTML("div.filter-inner>ul.subject-list", func(e *colly.HTMLElement) {
		e.ForEach("a", func(i int, element *colly.HTMLElement) {
			href := element.Attr("href")
			fmt.Println("打印href: ", href)
			if !strings.HasPrefix(href, "http") {
				href = viper.GetString("crawl.schema") + ":" + href
			}
			id := element.Attr("data-value")
			if id != "" {
				intId, err := strconv.Atoi(id)
				if err == nil {
					subjectId = append(subjectId, int(intId))
				}
			}
		})
	})
	cc.Visit(viper.GetString("crawl.schema") + "://" + viper.GetString("crawl.website"))
	return subjectId
}
