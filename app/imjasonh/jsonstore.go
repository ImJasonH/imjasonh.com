package imjasonh

// TODO: Support other request/response formats besides JSON (e.g., xml, gob)
// TODO: Figure out if PropertyList can support nested objects, or fail if they are detected.

import (
	"appengine"
	"appengine/datastore"
	"encoding/json"
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

func init() {
	http.HandleFunc("/jsonstore", jsonstore)
	http.HandleFunc("/jsonstore/", jsonstore)
}

// jsonstore dispatches requests to the relevant API method and arranges certain common state
func jsonstore(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	c := appengine.NewContext(r)
	if r.URL.Path == "/jsonstore" {
		switch r.Method {
		case "POST":
			insert(w, r.Body, c)
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
		case "POST":
			// This is strictly "replace all properties/values", not "add new properties, update existing"
			update(w, id, r.Body, c)
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
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
}

func get(w http.ResponseWriter, id int64, c appengine.Context) {
	k := datastore.NewKey(c, kind, "", id, nil)
	var plist datastore.PropertyList
	if err := datastore.Get(c, k, &plist); err != nil {
		if err == datastore.ErrNoSuchEntity {
			http.Error(w, "Not Found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	m := plistToMap(plist, k)
	json.NewEncoder(w).Encode(m)
}

func insert(w http.ResponseWriter, r io.Reader, c appengine.Context) {
	plist, m, err := jsonToPlist(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	k := datastore.NewIncompleteKey(c, kind, nil)
	k, err = datastore.Put(c, k, &plist)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	m[idKey] = k.IntID()
	json.NewEncoder(w).Encode(m)
}

func plistToMap(plist datastore.PropertyList, k *datastore.Key) map[string]interface{} {
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
	m[idKey] = k.IntID()
	return m
}

// jsonToPlist decodes a JSON stream into a PropertyList for storing in the datastore, and a JSON-encodable representation of the data.
func jsonToPlist(r io.Reader) (datastore.PropertyList, map[string]interface{}, error) {
	var m map[string]interface{}
	if err := json.NewDecoder(r).Decode(&m); err != nil {
		return nil, nil, err
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
	return plist, m, nil
}

// TODO: Add rudimentary single-property queries, pagination, sorting, etc.
func list(w http.ResponseWriter, c appengine.Context) {
	q := datastore.NewQuery(kind).Limit(10)

	r := []map[string]interface{}{}

	for t := q.Run(c); ; {
		var plist datastore.PropertyList
		k, err := t.Next(&plist)
		if err == datastore.Done {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		m := plistToMap(plist, k)
		r = append(r, m)
	}
	json.NewEncoder(w).Encode(r)
}

func update(w http.ResponseWriter, id int64, r io.Reader, c appengine.Context) {
	plist, m, err := jsonToPlist(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	k := datastore.NewKey(c, kind, "", id, nil)
	if _, err := datastore.Put(c, k, &plist); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	m[idKey] = id
	json.NewEncoder(w).Encode(m)
}
