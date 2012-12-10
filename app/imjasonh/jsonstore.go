package imjasonh

// TODO: Support other request/response formats besides JSON (e.g., xml, gob)
// TODO: Figure out if PropertyList can support nested objects, or fail if they are detected.
// TODO: Add rudimentary single-property queries, pagination, sorting, etc.
// TODO: Allow clients to specify the Kind? Namespace datastore by user identity (and maintain user identity)?

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
	kind         = "JsonObject"
	idKey        = "_id"
	createdKey   = "_created"
	defaultLimit = 10
)

func init() {
	http.HandleFunc("/jsonstore", jsonstore)
	http.HandleFunc("/jsonstore/", jsonstore)
}

type UserQuery struct {
	Limit, Offset                      int
	FilterKey, FilterType, FilterValue string
	Cursor                             string
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
			// TODO: Parse user request into UserQuery and pass to list method
			list(w, UserQuery{}, c)
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
	var m map[string]interface{}
	if err := json.NewDecoder(r).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	m[createdKey] = time.Now()

	plist := mapToPlist(m)

	k := datastore.NewIncompleteKey(c, kind, nil)
	k, err := datastore.Put(c, k, &plist)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	m[idKey] = k.IntID()
	json.NewEncoder(w).Encode(m)
}

// plistToMap transforms a PropertyList such as you would get from the datastore into a map[string]interface{} suitable for JSON-encoding.
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

// mapToPlist transforms a map[string]interface{} such as you would get from decoding JSON into a PropertyList to store in the datastore.
func mapToPlist(m map[string]interface{}) datastore.PropertyList {
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
	return plist
}

func list(w http.ResponseWriter, uq UserQuery, c appengine.Context) {
	limit := 3
	q := datastore.NewQuery(kind).Limit(limit)

	items := make([]map[string]interface{}, 0, limit)

	var crs datastore.Cursor
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
		items = append(items, m)
		if crs, err = t.Cursor(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	r := map[string]interface{}{
		"items":          items,
		"nextStartToken": crs.String(),
	}
	json.NewEncoder(w).Encode(r)
}

func update(w http.ResponseWriter, id int64, r io.Reader, c appengine.Context) {
	var m map[string]interface{}
	if err := json.NewDecoder(r).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	plist := mapToPlist(m)

	k := datastore.NewKey(c, kind, "", id, nil)
	if _, err := datastore.Put(c, k, &plist); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	m[idKey] = id
	json.NewEncoder(w).Encode(m)
}
