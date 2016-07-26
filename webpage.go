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

var responseTemplate = template.Must(template.ParseFiles("view.html", "query.html", "display.html"))
var apikey string = "rU3eqtaE4zBSzZKjoUS9Q7fVPbTmKmD2eOUr"

var validPath = regexp.MustCompile("^/(query|save|view|dispaly)/([a-zA-Z0-9]+)$")

type Page struct {
	Path     string                 `json: "Path"`
	UUid     string                 `json: "UUid"`
	Info     map[string]interface{} `json: "Info"`
	ReadData gosMAP.Data
}

type Tags struct {
	Path string `json: "Path"`
	Uuid string `json: "Uuid"`
}
type Information struct {
	TagSlices []Tags `json: "Tag"`
}

func createPage(title string, c *gin.Context) (*Information, error) {

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

	var uuid []string

	err = json.Unmarshal(item.Value, &uuid)

	if err != nil {
		return nil, err
	}

	var tags []Tags

	for i := range uuid {
		t, err := conn.Tag(uuid[i])

		if err != nil {
			return nil, err
		}

		tag := Tags{
			Path: t.Path,
			Uuid: t.Uuid,
		}

		tags = append(tags, tag)
	}

	return &Information{TagSlices: tags}, nil
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

	renderTemplate(c, "display", p)
}

func viewHandler(c *gin.Context) {

	title := c.Request.URL.Path[len("/view/"):]

	tags, err := createPage(title, c)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	renderTemplate(c, "view", tags)
}
func queryHandler(c *gin.Context) {

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

		q, err := conn.QueryList(fmt.Sprintf("select distinct uuid where Metadata/Location/Building = '%s'", building))

		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		b, err := json.Marshal(q)

		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, errors.New("Cannot marshal"))
			return
		}

		item := memcache.Item{
			Key:   key,
			Value: b,
		}

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

func renderTemplate(c *gin.Context, tmpl string, p interface{}) {
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
	router.GET("/display/:path", displayHandler)
	router.Run(":8000")

}
