package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/UCMAndesLab/gosMAP"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gin-gonic/gin"
	"html/template"
)

var apikey string = "rU3eqtaE4zBSzZKjoUS9Q7fVPbTmKmD2eOUr"

type Page struct {
	Title    string
	ReadData gosMAP.Data
}

func viewHandler(c *gin.Context) {
	fmt.Fprintf(c.Writer, "<h1>Something</h1>")
}

func queryHandler(c *gin.Context) {
	t := template.ParseFiles("query.html")
	t.Execute(c.Writer, nil)
}

func saveHandler(c *gin.Context) {
	building := c.Request.FormValue("query")

	d, err := Get(building)
}

func Get(building string) (memcache.Item, error) {
	conn, err := gosMAP.Connect("http://mercury:8079", apikey)
	if err != nil {
		return nil, errors.New("Cannot connect to mercury")
	}

	d, err := conn.Mc.Get(building)

	if err == nil {
		fmt.Println("The cached value exists; cache hit")
		return d, nil
	} else {
		fmt.Println("cache miss, doesn't exist")
		q := conn.QueryList(fmt.Sprintf("select distinct uuid where Metadata/Location/Building = '%s'", query))
		//d := conn.Get(q[0], 0, 0, 10)

		b := json.Marshal(q)
		err = conn.Mc.Set(&memcache.Item{
			Key:   building,
			Value: q,
		})

		if err != nil {
			return nil, errors.New("Something went wrong with setting the memcache")
		}
		return Get(building)
	}
}

func main() {
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})
	router.GET("/view/", viewHandler)
	router.GET("/query/", queryHandler)
	router.POST("/save/", saveHandler)
	router.Run(":8080")
}
