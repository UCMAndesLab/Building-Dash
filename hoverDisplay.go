package main

import (
	//"bytes"
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

	Info map[string]interface{} `json: "Info"`

	ReadData gosMAP.Data
}

type Tags struct {
	Path string `json: "Path"`
	Uuid string `json: "Uuid"`
}

type Information struct {
	Tag []Tags `json: "Tag"`
}

var apikey string = "rU3eqtaE4zBSzZKjoUS9Q7fVPbTmKmD2eOUr"

func viewHandler(c *gin.Context) {

	conn, err := gosMAP.Connect("http://mercury:8079", apikey)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

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

	var tags []Tags

	for i := range uuid {
		t, err := conn.Tag(uuid[i])

		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		tag := Tags{
			Path: t.Path,
			Uuid: t.Uuid,
		}

		tags = append(tags, tag)
	}

	//data, err := conn.Get(uuid[0], 0, 0, 10)

	/*if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	p := &Page{
		Path:     tag.Path,
		UUid:     uuid[0],
		Info:     tag.Metadata,
		ReadData: data[0],
	}*/

	info := &Information{
		Tag: tags,
	}
	t, err := template.ParseFiles("testView.html")

	if err != nil {

		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if err = t.Execute(c.Writer, info); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
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

func displayHandler(c *gin.Context) {
	uuid := c.Request.URL.Path[len("/display/"):]

	conn, err := gosMAP.Connect("http://mercury:8079", apikey)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	data, err := conn.Get(uuid, 0, 0, 10)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	tag, err := conn.Tag(uuid)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	p := &Page{
		Path:     tag.Path,
		UUid:     uuid,
		Info:     tag.Metadata,
		ReadData: data[0],
	}

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
	router.GET("/display/:path", displayHandler)
	router.Run(":8000")
}
