package main

import (
	"errors"
	"fmt"
	"github.com/UCMAndesLab/gosMAP"
	"github.com/gin-gonic/gin"
	"html/template"
	//"io/ioutil"
	"net/http"
	"regexp"
	//"strconv"
	//"time"
)

var responseTemplate = template.Must(template.ParseFiles("view.html", "query.html"))
var apikey string = "rU3eqtaE4zBSzZKjoUS9Q7fVPbTmKmD2eOUr"

var validPath = regexp.MustCompile("^/(query|save|view)/([a-zA-Z0-9]+)$")
var uuid string

type Page struct {
	Title    string
	ReadData []gosMAP.ReadPair
}

/*type Sample struct{
	Time time.Time
}*/

func createPage(title string) (*Page, error) {

	//uuid := "51427e0d-ee71-5df2-90b5-ebc3cc720f87"
	conn, e := gosMAP.Connect("http://mercury:8079", apikey)
	if e != nil {
		return nil, e
	}

	d, err := conn.Get(uuid, 0, 0, 10)
	fmt.Println(uuid)
	if err != nil {
		return nil, err
	}

	return &Page{Title: title, ReadData: d[0].Readings}, nil
}

func viewHandler(c *gin.Context) {

	title := c.Request.URL.Path
	//query := c.Request.FormValue("query")

	fmt.Println(title)

	/*if err != nil {
		return
	}*/

	p, err := createPage(title)

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

	conn, err := gosMAP.Connect("http://mercury:8079", apikey)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Cannot connect to the server"))
	}

	query := c.Request.FormValue("query")

	//fmt.Println(query)

	//d, err := conn.QueryList("select distinct uuid where Metadata/Location/Building = 'Facilities A'")
	d, err := conn.QueryList(fmt.Sprintf("select distinct uuid where Metadata/Location/Building = '%s'", query))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	uuid = d[0]
	for i := range d {
		fmt.Println(d[i])
	}

	c.Redirect(http.StatusFound, "/view/"+query)
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
	router.GET("/view/:query", viewHandler)
	router.GET("/query/", queryHandler) //ok, for a new page with no title yet, don't include :adding in the url
	router.POST("/save/", saveHandler)
	router.Run(":8080")
	//createPage("TestData")

	//exampleQuery()

}
