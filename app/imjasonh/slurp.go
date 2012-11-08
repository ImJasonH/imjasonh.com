package imjasonh

import (
	"appengine"
	"appengine/urlfetch"
	storage "code.google.com/p/google-api-go-client/storage/v1beta1"
	"net/http"
	"time"
)

const (
	bucket = "BUCKET" // TODO
)

var (
	gcs    storage.Service
	expiry time.Time
)

// slurp copies the contents of a given URL to a given GCS key.
func slurp(w http.ResponseWriter, req *http.Request) {
	key := req.FormValue("key")
	url := req.FormValue("url")

	c := appengine.NewContext(req)
	if err := updateStorageClient(c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	client := urlfetch.Client(c)
	if resp, err := client.Get(url); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	store := gcs.Objects.Insert(bucket, &storage.Object{Name: key}).Media(resp.Body)
	if _, err := store.Do(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func updateStorageClient(c appengine.Context) error {
	if gcs != nil && expiry != nil && time.Now().Before(expiry) {
		return nil
	}

	// TODO: Cache the transport for as long as the access token is valid.
	if token, exp, err := appengine.AccessToken(c, storage.DevstorageRead_writeScope); err != nil {
		return err
	}
	transport := &oauth.Transport{
		Token:     token,
		Transport: http.DefaultTransport,
	}
	expiry = exp
	if gcs, err = storage.New(t.Client()); err != nil {
		return err
	}
	return nil
}
