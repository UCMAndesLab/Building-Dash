package main

import (
	//"encoding/json"
	//"errors"
	//"fmt"
	//"github.com/UCMAndesLab/gosMAP"
	"github.com/gin-gonic/gin"
	"html/template"
	//"net/http"
	//"regexp"
	//"strings"
	//"strconv"
	//"time"
	//"github.com/bradfitz/gomemcache/memcache"
)

func viewHandler(c *gin.Context) {

	t, _ := template.ParseFiles("testView.html")

	t.Execute(c.Writer, nil)
}

func main() {

	router := gin.Default()

	router.LoadHTMLFiles("testIndex.html")

	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "testIndex.html", nil)
	})

	router.GET("/testView/", viewHandler)
	router.Run(":8080")
}
