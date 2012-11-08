package imjasonh

import (
	"appengine"
	"appengine/urlfetch"
	"code.google.com/p/goauth2/oauth"
	storage "code.google.com/p/google-api-go-client/storage/v1beta1"
	"fmt"
	"net/http"
	"time"
)

const (
	bucket = "imjasonh-slurp"
)

var (
	gcs    storage.Service
	expiry = time.Unix(0, 0)
)

// slurp copies the contents of a given URL to a given GCS key.
func slurp(w http.ResponseWriter, req *http.Request) {
	key := req.FormValue("key")
	url := req.FormValue("url")

	if key == "" || url == "" {
		http.Error(w, "Must specify key and url", http.StatusBadRequest)
		return
	}

	c := appengine.NewContext(req)
	if err := updateStorageClient(c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	client := urlfetch.Client(c)
	resp, err := client.Get(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	store := gcs.Objects.Insert(bucket, &storage.Object{Name: key}).Media(resp.Body)
	if _, err := store.Do(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gsUrl := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucket, key)
	http.Redirect(w, req, gsUrl, http.StatusMovedPermanently)
}

func updateStorageClient(c appengine.Context) error {
	if gcs != nil && time.Now().Before(expiry) {
		return nil
	}

	token, exp, err := appengine.AccessToken(c, storage.DevstorageRead_writeScope)
	if err != nil {
		return err
	}
	transport := &oauth.Transport{
		Token:     token,
		Transport: http.DefaultTransport,
	}
	expiry = exp
	if gcs, err = storage.New(transport.Client()); err != nil {
		return err
	}
	return nil
}
