package main

import (
	"errors"
	"fmt"
	"github.com/UCMAndesLab/goSMAP"
	"github.com/gin-gonic/gin"
	"html/template"
	//"io/ioutil"
	"net/http"
	"regexp"
	//"strconv"
	//"time"
)

var responseTemplate = template.Must(template.ParseFiles("view.html", "add.html"))

var validPath = regexp.MustCompile("^/(add|save|view)/([a-zA-Z0-9]+)$")

type Page struct {
	Title    string
	ReadData []gosMAP.ReadPair
}

/*type Sample struct{
	Time time.Time
}*/

func createPage(title string) (*Page, error) {
	key := "rU3eqtaE4zBSzZKjoUS9Q7fVPbTmKmD2eOUr"
	uuid := "51427e0d-ee71-5df2-90b5-ebc3cc720f87"
	conn, e := gosMAP.Connect("http://mercury:8079", key)

	d, err := conn.Get(uuid, 0, 0, 10)
	if e != nil {
		panic(err)
	}

	return &Page{Title: title, ReadData: d[0].Readings}, nil
}

func exampleQuery(c *gin.Context) []gosMAP.Data {
	key := "rU3eqtaE4zBSzZKjoUS9Q7fVPbTmKmD2eOUr"
	uuid := "51427e0d-ee71-5df2-90b5-ebc3cc720f87"

	conn, err := gosMAP.Connect("http://mercury:8079", key)

	if err != nil {
		panic(err)
	}

	q := conn.Query(fmt.Sprintf("select * where uuid='%s'", uuid))

	/*for _, r := range q[0].Readings {
		fmt.Printf("%s", r.Time.String())
	}*/

	return q
}

func viewHandler(c *gin.Context) {

	title, err := getTitle(c)

	if err != nil {
		return
	}

	p, err := createPage(title)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	//responseTemplate.Execute(c.Writer, p)
	renderTemplate(c, "view", p)
}
func addHandler(c *gin.Context) {
	//title := c.Request.FormValue("title")
	renderTemplate(c, "add", nil)
}
func saveHandler(c *gin.Context) {
	title := c.Request.FormValue("title")

	c.Redirect(http.StatusFound, "/view/"+title)
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
	router.GET("/view/:title", viewHandler)
	router.GET("/add/", addHandler) //ok, for a new page with no title yet, don't include :adding in the url
	router.POST("/save/", saveHandler)
	router.Run(":8080")
	//createPage("TestData")

	//exampleQuery()

}
