package main

import (
	"crawler/handler/course"
	"crawler/handler/subjects"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

func ServeWeb() {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://github.com"
		},
		MaxAge: 12 * time.Hour,
	}))
	r.GET("/course-list", course.List)
	r.GET("/subjects-list", subjects.List)
	r.Run(":9090")
}
