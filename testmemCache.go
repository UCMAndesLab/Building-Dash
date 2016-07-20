package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/UCMAndesLab/gosMAP"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gin-gonic/gin"
	"html/template"
	//"log"
	"net/http"
	//"time"
	"strings"
)

var apikey string = "rU3eqtaE4zBSzZKjoUS9Q7fVPbTmKmD2eOUr"

type Page struct {
	Title    string
	ReadData gosMAP.Data
}

func viewHandler(c *gin.Context) {

	conn, err := gosMAP.Connect("http://mercury:8079", apikey)
	conn.ConnectMemcache("localhost:11211")

	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	buidling := c.Request.URL.Path[len("/view/"):]

	key := strings.Replace(buidling, " ", "\\", -1)

	item, err := conn.Mc.Get(key)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var query []string

	err = json.Unmarshal(item.Value, &query)

	fmt.Println(query[0])

	d, err := conn.Get(query[0], 0, 0, 10)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	fmt.Fprintf(c.Writer, "<h1>%s</h1>", d)
}

func queryHandler(c *gin.Context) {

	t, _ := template.ParseFiles("query.html")
	t.Execute(c.Writer, nil)
}

func saveHandler(c *gin.Context) {
	building := c.Request.FormValue("query")

	key := strings.Replace(building, " ", "\\", -1)

	conn, err := gosMAP.Connect("http://mercury:8079", apikey)

	conn.ConnectMemcache("localhost:11211")

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Cannot connect to mercury"))
	}

	_, err = conn.Mc.Get(key)

	if err != nil {
		/*c.AbortWithError(http.StatusInternalServerError, errors.New("Cache miss"))
		return*/

		q, _ := conn.QueryList(fmt.Sprintf("select distinct uuid where Metadata/Location/Building = '%s'", building))

		b, err := json.Marshal(q)

		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, errors.New("Cannot marshal"))
			return
		}

		item := memcache.Item{Key: key, Value: b, Expiration: 0}

		err = conn.Mc.Add(&item)

		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Redirect(http.StatusFound, "/view/"+building)

	} else {
		c.Redirect(http.StatusFound, "/view/"+building)
	}
}

func main() {
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})
	router.GET("/view/:key", viewHandler)
	router.GET("/query/", queryHandler)
	router.POST("/save/", saveHandler)
	router.Run(":8080")
}
