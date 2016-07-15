<h1>Buidling Dash</h1>
<h4>This is a project that can execute for a local host.</h4>
<h4>To begin the program, type go run webpage.go</h4>

<h4>Then, go to your browser and search: locahlhost:8080/<br>
You will then be directed to the home page and you select search,<br>
which will redirect you to the query page. 
</h4>


<h3>Usage<h3>

<h4><b>func createPage</h4>
```go
func createPage(title string) (*Page, error){}
```
<p>
The title is the based on the query that an individual has made. 
Depending on the query, if the uuid's are valid the gosMAP.Data slices
are retrieved via the Get function from the gosMAP documentation. 
When the values are retrieved, they'll be placed within a page struct. 
</p>
<h4><b>type Page</h4>
```go 
type Page struct{
  Title string
  ReadData []gosMAP.Data
}
```
