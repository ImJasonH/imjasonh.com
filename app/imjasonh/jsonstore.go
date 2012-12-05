package imjasonh

import (
	"appengine"
	"appengine/datastore"
	"io"
	"net/http"
	"encoding/json"
	"strconv"
)

const (
	kind = "JsonObject"
)

func init() {
	http.HandleFunc("/jsonstore", insert)
	http.HandleFunc("/jsonstore/", get)
}

func get(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Support DELETE
	// TODO: Save creation timestamp on objects

	sid := r.URL.Path[len("/jsonstore/"):]
	id, err := strconv.ParseInt(sid, 10, 64)
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
		m[p.Name] = p.Value
	}
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

	plist := make(datastore.PropertyList, 0)
	for k, v := range m {
		p := datastore.Property {
			Name: k,
			Value: v,
		}
		plist = append(plist, p)
	}

	c := appengine.NewContext(r)

	k := datastore.NewIncompleteKey(c, kind, nil)
	k, err = datastore.Put(c, k, &plist)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	m["__key__"] = k.IntID()
	json.NewEncoder(w).Encode(m)
}

func parse(r io.Reader) (map[string]interface{}, error) {
	var m map[string]interface{}
	if err := json.NewDecoder(r).Decode(&m); err != nil {
		return nil, err
	}
	return m, nil
}
