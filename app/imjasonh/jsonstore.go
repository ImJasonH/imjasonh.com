package imjasonh

import (
	"appengine"
	"appengine/datastore"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

const (
	kind       = "JsonObject"
	idKey      = "_id"
	createdKey = "_created"
)

func init() {
	http.HandleFunc("/jsonstore", jsonstore)
	http.HandleFunc("/jsonstore/", jsonstore)
}

func jsonstore(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	c := appengine.NewContext(r)
	if r.URL.Path == "/jsonstore" {
		switch r.Method {
		case "POST":
			insert(w, r)
			return
		case "GET":
			list(w, c)
			return
		}
	} else {
		sid := r.URL.Path[len("/jsonstore/"):]
		if path == "" {
			http.Error(w, "Must specify ID", http.StatusBadRequest)
			return
		}
		id, err := strconv.ParseInt(sid, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch r.Method {
		case "GET":
			get(w, id, c)
			return
		case "DELETE":
			delete(w, id, c)
			return
		case "PUT":
			update(w, id, c)
			return
		}
	}
	http.Error(w, "Unsupported Method", http.StatusMethodNotAllowed)
}

func delete(w http.ResponseWriter, id int64, c appengine.Context) {
	k := datastore.NewKey(c, kind, "", id, nil)
	if err := datastore.Delete(c, k); err != nil {
		if err == datastore.ErrNoSuchEntity {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func get(w http.ResponseWriter, id int64, c appengine.Context) {
	k := datastore.NewKey(c, kind, "", id, nil)
	var plist datastore.PropertyList
	if err := datastore.Get(c, k, &plist); err != nil {
		if err == datastore.ErrNoSuchEntity {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	m := make(map[string]interface{})
	for _, p := range plist {
		if _, exists := m[p.Name]; exists {
			if _, isArr := m[p.Name].([]interface{}); isArr {
				m[p.Name] = append(m[p.Name].([]interface{}), p.Value)
			} else {
				m[p.Name] = []interface{}{m[p.Name], p.Value}
			}
		} else {
			m[p.Name] = p.Value
		}
	}
	m[idKey] = id
	json.NewEncoder(w).Encode(m)
}

func insert(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Unsupported Method", http.StatusMethodNotAllowed)
		return
	}
	var m map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	m[createdKey] = time.Now()

	plist := make(datastore.PropertyList, 0, len(m))
	for k, v := range m {
		if _, mult := v.([]interface{}); mult {
			for _, mv := range v.([]interface{}) {
				plist = append(plist, datastore.Property{
					Name:     k,
					Value:    mv,
					Multiple: true,
				})
			}
		} else {
			plist = append(plist, datastore.Property{
				Name:  k,
				Value: v,
			})
		}
	}

	c := appengine.NewContext(r)

	k := datastore.NewIncompleteKey(c, kind, nil)
	k, err := datastore.Put(c, k, &plist)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	m[idKey] = k.IntID()
	json.NewEncoder(w).Encode(m)
}

func list(w http.ResponseWriter, c appengine.Context) {
	// TODO: Implement this, with rudimentary queries.
}

func update(w http.ResponseWriter, id int64, c appengine.Context) {
	// TODO: Implement this.
}
