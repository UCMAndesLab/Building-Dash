package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/UCMAndesLab/gosMAP"
	"github.com/gin-gonic/gin"
	"html/template"
	"net/http"
	"regexp"
	"strings"
	//"strconv"
	//"time"
	"github.com/bradfitz/gomemcache/memcache"
)

var responseTemplate = template.Must(template.ParseFiles("view.html", "query.html"))
var apikey string = "rU3eqtaE4zBSzZKjoUS9Q7fVPbTmKmD2eOUr"

var validPath = regexp.MustCompile("^/(query|save|view)/([a-zA-Z0-9]+)$")

type Page struct {
	Title    string
	ReadData []gosMAP.Data
}

/*type Sample struct{
	Time time.Time
}*/

func createPage(title string, c *gin.Context) (*Page, error) {

	//uuid := "51427e0d-ee71-5df2-90b5-ebc3cc720f87"
	conn, err := gosMAP.Connect("http://mercury:8079", apikey)
	conn.ConnectMemcache("localhost:11211")

	if err != nil {
		return nil, err
	}

	building := c.Request.URL.Path[len("/view/"):]

	key := strings.Replace(building, " ", "\\", -1)

	item, err := conn.Mc.Get(key)

	if err != nil {
		return nil, err
	}

	var query []string

	err = json.Unmarshal(item.Value, &query)

	//d, err := conn.Get(query[0], 0, 0, 10)

	if err != nil {
		return nil, err
	}
	/*for i := range uuid {
		d, err := conn.Get(uuid[i], 0, 0, 10)
		//fmt.Println(uuid)
		if err != nil {
			return nil, err
		}
		if len(d) != 0 {
			if len(d[0].Readings) != 0 {
				data = append(data, d[0])
			}
		}
	}*/

	var data []gosMAP.Data
	for i := range query {
		d, err := conn.Get(query[i], 0, 0, 10)

		if err != nil {
			return nil, err
		}

		if len(d) != 0 {
			if len(d[0].Readings) != 0 {
				data = append(data, d[0])
			}
		}
	}
	return &Page{Title: title, ReadData: data}, nil
}

func viewHandler(c *gin.Context) {

	title := c.Request.URL.Path[len("/view/"):]
	//query := c.Request.FormValue("query")

	p, err := createPage(title, c)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	//responseTemplate.Execute(c.Writer, p)
	renderTemplate(c, "view", p)
	//fmt.Fprintf(c.Writer, "<h1>Something</h1>")
}
func queryHandler(c *gin.Context) {
	//title := c.Request.FormValue("title")
	renderTemplate(c, "query", nil)
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

func renderTemplate(c *gin.Context, tmpl string, p *Page) {
	err := responseTemplate.ExecuteTemplate(c.Writer, tmpl+".html", p)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
}

func getTitle(c *gin.Context) (string, error) {
	m := validPath.FindStringSubmatch(c.Request.URL.Path)

	if m == nil {
		c.AbortWithError(http.StatusNotFound, errors.New("Invalid title"))
		return "", errors.New("Invalid title")
	}

	return m[2], nil
}

func main() {

	router := gin.Default()
	router.LoadHTMLFiles("index.html")
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})
	router.GET("/view/:query", viewHandler)
	router.GET("/query/", queryHandler) //ok, for a new page with no title yet, don't include :adding in the url
	router.POST("/save/", saveHandler)
	router.Run(":8080")
	//createPage("TestData")

	//exampleQuery()

}
