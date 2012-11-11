package imjasonh

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"fmt"
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
	form = `<html><body><form action="/go" method="POST">
  <label for="key">Key</label>
  <input type="text" name="key" id="key"></input><br />

  <label for="url">URL</label>
  <input type="text" name="url" id="url"></input><br />

  <input type="submit" value="Submit"></input> 
</form>
<a href="%s">Log out</a>
</body></html>`

	login = `<html><body>
  <a href="%s">Log in to create shortcuts</a>
</body></html>`

	success = `<html><body>
  <h1>Success!</h1>
</body></html>`

	path = "/go"
)

// go_ registers new shortcut links, and redirects to those shortcut links.
func go_(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Form.Get("ok") == "1" {
		fmt.Fprintf(w, success)
		return
	}

	q := r.Form.Get("q")
	// TODO: Set up handler so that it matches regexp so the URL can be /go/asdf instead of /go?q=asdf
	if q == "" {
		usr := user.Current(c)
		if usr == nil {
			url, _ := user.LoginURL(c, path)
			fmt.Fprintf(w, login, url)
			return
		}

		if r.Method == "GET" {
			// Display new shortcut form
			logout, _ := user.LogoutURL(c, r.URL.String())
			fmt.Fprintf(w, fmt.Sprintf(form, logout))
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
			http.Redirect(w, r, "/go?ok=1", http.StatusSeeOther)
		} else {
			// TODO: Support DELETE?
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		}
	} else {
		var s Shortcut
		dsKey := datastore.NewKey(c, "Shortcut", q, 0, nil)
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
