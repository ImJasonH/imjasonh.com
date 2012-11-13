package imjasonh

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"time"
)

type Shortcut struct {
	Key     string
	URL     string
	User    string
	Created time.Time
}

const (
	form = `<html><body>
<form action="/go" method="POST">
  <label for="key">Key</label>
  <input type="text" name="key" id="key"></input><br />

  <label for="url">URL</label>
  <input type="text" name="url" id="url"></input><br />

  <input type="submit" value="Submit"></input> 
</form>
<a href="{{.LogoutURL}}">Log out</a>
</body></html>
`

	login = `<html><body>
  <a href="{{.LoginURL}}">Log in to create shortcuts</a>
</body></html>`

	path = "/go"
)

func init() {
	http.HandleFunc("/go", go_)
	http.HandleFunc("/go/", go_)
}

// go_ registers new shortcut links, and redirects to those shortcut links.
func go_(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.URL.Path == path || r.URL.Path == path+"/" {
		usr := user.Current(c)
		if usr == nil {
			loginURL, err := user.LoginURL(c, path)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			t := template.Must(template.New("login").Parse(login))
			t.Execute(w, map[string]string{
				"LoginURL": loginURL,
			})
			return
		}

		if r.Method == "GET" {
			// Display new shortcut form
			logout, err := user.LogoutURL(c, r.URL.String())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			t := template.Must(template.New("form").Parse(form))
			t.Execute(w, map[string]string{
				"LogoutURL": logout,
			})
		} else if r.Method == "POST" {
			// Save shortcut
			k := r.FormValue("key")
			u := r.FormValue("url")

			if k == "" || u == "" {
				http.Error(w, "Must provide key and URL", http.StatusBadRequest)
				return
			}
			parsed, err := url.Parse(u)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if !parsed.IsAbs() {
				http.Error(w, "URL must be absolute", http.StatusBadRequest)
				return
			}
			if r.URL.IsAbs() && parsed.Host == r.URL.Host {
				http.Error(w, "URL cannot link to this app", http.StatusBadRequest)
				return
			}

			dsKey := datastore.NewKey(c, "Shortcut", k, 0, nil)

			// Prevent overwriting existing shortcuts owned by other users
			var s Shortcut
			err = datastore.Get(c, dsKey, &s)

			if err == nil && s.User != usr.Email {
				http.Error(w, "Shortcut is owned by another user", http.StatusBadRequest)
				return
			}
			if err != nil && err != datastore.ErrNoSuchEntity {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			s = Shortcut{
				URL:     u,
				User:    usr.Email,
				Created: time.Now(),
			}
			if _, err := datastore.Put(c, dsKey, &s); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Fprintf(w, "Success!")
		} else {
			// TODO: Support DELETE?
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		}
	} else {
		key := r.URL.Path[len(path)+1:]

		var s Shortcut
		dsKey := datastore.NewKey(c, "Shortcut", key, 0, nil)
		if err := datastore.Get(c, dsKey, &s); err != nil {
			if err == datastore.ErrNoSuchEntity {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		http.Redirect(w, r, s.URL, http.StatusMovedPermanently)
	}
}
