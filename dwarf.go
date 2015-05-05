package dwarf

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"time"
	
	"appengine"
	"appengine/datastore"
)

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/savedoc", saveDoc)
	http.HandleFunc("/onedoc", oneDoc)
}

func oneDoc(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	dockey_encoded := r.FormValue("dockey")
	doc := new(Document)
	if dockey_encoded != "" {
		dockey, _ := datastore.DecodeKey(dockey_encoded)
		if err := datastore.Get(ctx, dockey, doc); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}	
		doc.EncodedKey = dockey_encoded
	}

	doc.Html = Markdowner(doc.Content)

	if err := editingTemplate.Execute(w, doc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}


type KeyDoc struct {
	Name string
	Key string
}


func root(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	limit := 20
	q := datastore.NewQuery("Document").Ancestor(documentKey(c)).Order("-LastUpdated").Limit(limit)
	documents := make([]Document, 0, limit)
	keys, err := q.GetAll(c, &documents)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	num_docs := len(documents)
	key_docs := make([]KeyDoc, num_docs);
	for i := 0; i < num_docs; i++ {
		key_doc := KeyDoc {
			Name: documents[i].Name,
			Key: keys[i].Encode(),
		}
		key_docs[i] = key_doc
	}

	if err := folderTemplate.Execute(w, key_docs); err != nil {
		fmt.Fprint(w, "Error making key doc in template")
	}
}


type Document struct {
	Name string
	Content string
	Html template.HTML
	LastUpdated time.Time
	EncodedKey string
}


func documentKey(ctx appengine.Context) *datastore.Key {
	return datastore.NewKey(ctx, "DocumentParent", "documentkey", 0, nil)
}


func saveDoc(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	doc := Document{
		Content: r.FormValue("doctext"),
		Name: r.FormValue("docname"),
		LastUpdated: time.Now(),
		EncodedKey: r.FormValue("dockey"),
	}
	var doc_key *datastore.Key
	if doc.EncodedKey == "" {
		doc_key = datastore.NewIncompleteKey(ctx, "Document", documentKey(ctx))
	} else {
		doc_key, _ = datastore.DecodeKey(doc.EncodedKey)
	}
	_, err := datastore.Put(ctx, doc_key, &doc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}


var folderTemplate = template.Must(template.New("Folder").Parse(folderTemplateHTML))
var editingTemplate = template.Must(template.New("EditOneDoc").Parse(editingTemplateHTML))


const editingTemplateHTML = `
<html>
  <body>
    <div>
      <form action="/savedoc" method="post">
        <div><input name="docname" value="{{ .Name }}" cols="80"></input></div>
        <div>
            <textarea name="doctext" rows="10" cols="80">{{ .Content }}</textarea>
        </div>
        <div><input name="dockey" type="hidden" value="{{ .EncodedKey }}"></input></div>
        <div><input type="submit" value="Save doc"></div>
      </form>
    </div>
   <div>
      <p>{{ .Name }}, {{ .LastUpdated.Format "Mon Jan 2" }}</p>
      <div>{{ .Html }}</div>
   </div>
  </body>
</html>   
`

const folderTemplateHTML = `
<html>
  <body>
    <div>
      <h1>Files</h1>
    </div>
    </div>
    {{range .}}
      <div>
        <a href="/onedoc?dockey={{.Key}}">
          {{ .Name }}
        </a>
      </div>
    {{end}}
  </body>
</html>
`


// UTILS

func Markdowner(md_text string) template.HTML {
	html_text := make([]byte, 0, len(md_text) << 1)
	buffer := bytes.NewBuffer(html_text)

	newLine := true
	inUl := false

	for i := range md_text {
		c := md_text[i]

		// Ignore carriage returns.
		if c == '\r' {
			continue
		}

		if newLine {
 
			if c == '\n' && inUl {
				buffer.WriteString("</ul>")
				inUl = false
			} else if c == ' ' || c == '\t' {
				buffer.WriteByte(c)
			} else if c == '+' || c == '-' {
				if inUl {
					buffer.WriteString("<li>")
				} else {
					inUl = true
					buffer.WriteString("<ul><li>")
				}
				newLine = false
			} else if c != '\n' {
				buffer.WriteByte(c)
				newLine = false
			}
		} else {
			if c == '\n' {
				newLine = true
				buffer.WriteString("<br/>")
			} else {
				buffer.WriteByte(c)
			}
		}
	}

	return template.HTML(buffer.Bytes())
}
