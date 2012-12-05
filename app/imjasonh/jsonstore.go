package imjasonh

import (
	"appengine"
	"appengine/datastore"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	kind       = "JsonObject"
	idKey      = "_id"
	createdKey = "_created"
)

// TODO: Support PUT to update entities
func init() {
	http.HandleFunc("/jsonstore", insert)
	http.HandleFunc("/jsonstore/", getOrDelete)
}

func getOrDelete(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		get(w, r)
	case "DELETE":
		delete(w, r)
	default:
		http.Error(w, "Unsupported Method", http.StatusMethodNotAllowed)
	}
}

func getID(path string) (int64, error) {
	sid := path[len("/jsonstore/"):]
	if path == "" {
		return 0, errors.New("Must specify ID")
	}
	id, err := strconv.ParseInt(sid, 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func delete(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	c := appengine.NewContext(r)
	k := datastore.NewKey(c, kind, "", id, nil)
	if err = datastore.Delete(c, k); err != nil {
		if err == datastore.ErrNoSuchEntity {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func get(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c := appengine.NewContext(r)
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
	m, err := parse(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	m[createdKey] = time.Now()

	plist := make(datastore.PropertyList, 0)
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
	k, err = datastore.Put(c, k, &plist)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	m[idKey] = k.IntID()
	json.NewEncoder(w).Encode(m)
}

func parse(r io.Reader) (map[string]interface{}, error) {
	var m map[string]interface{}
	if err := json.NewDecoder(r).Decode(&m); err != nil {
		return nil, err
	}
	return m, nil
}
