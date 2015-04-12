package dwarf

import (
	"net/http"
	"time"

	"html/template"
	"appengine"
	"appengine/datastore"
)

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/savedoc", saveDoc)
}


func root(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	// Get from datastore
	limit := 20
	q := datastore.NewQuery("Document").Ancestor(documentKey(c)).Order("-LastUpdated").Limit(limit)
	documents := make([]Document, 0, limit)
	if _, err := q.GetAll(c, &documents); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := readingTemplate.Execute(w, documents); err != nil {
		default_doc := Document {
			Name: "black",
			Content: "Catch me if you can",
			LastUpdated: time.Now(),
		}
		readingTemplate.Execute(w, []Document{default_doc})
	}
}


type  Document struct {
	Name string
	Content string
	LastUpdated time.Time
}


func documentKey(ctx appengine.Context) *datastore.Key {
	return datastore.NewKey(ctx, "DocBook", "default_docbook", 0, nil)
}


func saveDoc(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	doc := Document{
		Content: r.FormValue("doctext"),
		Name: r.FormValue("docname"),
		LastUpdated: time.Now(),
	}
	doc_key := datastore.NewIncompleteKey(ctx, "Document", documentKey(ctx))
	_, err := datastore.Put(ctx, doc_key, &doc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}


var readingTemplate = template.Must(template.New("ty").Parse(editingTemplateHTML))


const editingTemplateHTML = `
<html>
  <body>
    <div>
      <form action="/savedoc" method="post">
        <div><input name="docname" cols="80"></input></div>
        <div><textarea name="doctext" rows="10" cols="80"></textarea></div>
        <div><input type="submit" value="Save doc"></div>
      </form>
    </div>
    {{range .}}
      <div>
        <p>{{ .Name }}, {{ .LastUpdated.Format "Mon Jan 2" }}</p>
        <p>{{ .Content }}</p>
      </div>
    {{end}}
  </body>
</html>
`
