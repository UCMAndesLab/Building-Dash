package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/UCMAndesLab/gosMAP"
	"github.com/gin-gonic/gin"
	"html/template"
	"net/http"
	//"regexp"
	"strings"
	//"strconv"
	"github.com/bradfitz/gomemcache/memcache"
	//"time"
)

type Page struct {
	Path string `json: "Path"`

	UUid string `json: "UUid"`

	Info map[string]interface{}

	ReadData gosMAP.Data
}

var apikey string = "rU3eqtaE4zBSzZKjoUS9Q7fVPbTmKmD2eOUr"

func viewHandler(c *gin.Context) {

	conn, err := gosMAP.Connect("http://mercury:8079", apikey)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	t, _ := template.ParseFiles("testView.html")

	conn.ConnectMemcache("localhost:11211")

	building := c.Request.URL.Path[len("/testView/"):]

	key := strings.Replace(building, " ", "\\", -1)

	item, err := conn.Mc.Get(key)

	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	var uuid []string

	err = json.Unmarshal(item.Value, &uuid)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	tag, err := conn.Tag(uuid[0])

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	data, err := conn.Get(uuid[0], 0, 0, 10)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	p := &Page{
		Path:     tag.Path,
		UUid:     uuid[0],
		Info:     tag.Metadata,
		ReadData: data[0],
	}
	t.Execute(c.Writer, p)
}

func queryHandler(c *gin.Context) {
	t, _ := template.ParseFiles("query.html")
	t.Execute(c.Writer, nil)
}

func saveHandler(c *gin.Context) {

	conn, err := gosMAP.Connect("http://mercury:8079", apikey)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	conn.ConnectMemcache("localhost:11211")

	building := c.Request.FormValue("query")

	key := strings.Replace(building, " ", "\\", -1)

	//fmt.Println(key)

	_, err = conn.Mc.Get(key)

	if err != nil {

		query, err := conn.QueryList(fmt.Sprintf("select distinct uuid where Metadata/Location/Building = '%s'", building))

		if err != nil {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}

		b, err := json.Marshal(query)

		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		item := memcache.Item{
			Key:   key,
			Value: b,
		}

		err = conn.Mc.Add(&item)

		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, errors.New(err.Error()+":line 154"))
			return
		}

		c.Redirect(http.StatusFound, "/testView/"+building)
	} else {
		c.Redirect(http.StatusFound, "/testView/"+building)
	}

}

func hoverHandler(c *gin.Context) {

	path := c.Request.URL.Path[len("/display/"):]

	conn, err := gosMAP.Connect("http://mercury:8079", apikey)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	conn.ConnectMemcache("localhost:11211")

	item, err := conn.Mc.Get(path)

	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	var p Page

	err = json.Unmarshal(item.Value, &p)

	t, _ := template.ParseFiles("display.html")

	t.Execute(c.Writer, p)

}

func main() {

	router := gin.Default()

	router.LoadHTMLFiles("testIndex.html")

	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "testIndex.html", nil)
	})

	router.GET("/testView/:building", viewHandler)

	router.GET("/query/", queryHandler)
	router.POST("/save/", saveHandler)
	router.GET("/display/:path", hoverHandler)
	router.Run(":8080")
}
