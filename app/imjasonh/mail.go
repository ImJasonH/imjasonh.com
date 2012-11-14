package imjasonh

import (
	"appengine"
	"appengine/datastore"
	"appengine/taskqueue"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	delay = 2 * time.Hour
	limit = 100

// TODO: Add a button to immediately delete a message.
	mailHTML = `<html><body>
<h3>{{.To}}</h3>
<table>
  {{range .Mails}}
    <tr>
      <td>{{.Received}}</td>
      <td><pre>{{.Text}}</pre></td>
    </tr>
  {{else}}
    No mails have been sent to this address.
  {{end}}
</table>
</body></html>`
)

func init() {
	http.HandleFunc("/_ah/mail/", inbound)
	http.HandleFunc("/_ah/queue/reapMail", reapMail)
	http.HandleFunc("/mail/", view)
}

type Mail struct {
	To       string
	Text     []byte
	Received time.Time
}

// inbound handles incoming email requests by persisting a new Mail and enqueing a task to delete it later.
func inbound(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		c.Errorf("%v", err.Error())
		return
	}

	m := Mail{
		To:       r.URL.Path[len("/_ah/mail/"):],
		Text:     b,
		Received: time.Now(),
	}
	dsKey := datastore.NewIncompleteKey(c, "Mail", nil)
	dsKey, err = datastore.Put(c, dsKey, &m)
	if err != nil {
		c.Errorf("%v", err.Error())
		return
	}

	task := taskqueue.Task{
		Delay:   delay,
		Payload: []byte(dsKey.String()),
	}
	if _, err = taskqueue.Add(c, &task, "reapMail"); err != nil {
		c.Errorf("%v", err)
	}
}

// reapMail handles a TaskQueue request to delete an old Mail.
func reapMail(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	key, err := ioutil.ReadAll(r.Body)
	if err != nil {
		c.Errorf("%v", err)
		return
	}
	dsKey := datastore.NewKey(c, "Mail", string(key), 0, nil)
	if err := datastore.Delete(c, dsKey); err != nil {
		c.Errorf("%v", err)
	}
}

// view lists the Mails sent to a particular address.
func view(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	if r.URL.Path == "/mail/" {
		http.Error(w, "Specify a mailbox", http.StatusBadRequest)
		return
	}
	to := strings.Split(r.URL.Path[len("/mail/"):], "@")[0]

	q := datastore.NewQuery("Mail").
		Filter("To =", to).
		Order("-Received").
		Limit(100)
	mails := make([]map[string]string, 100)
	i := 0
	for t := q.Run(c); ; {
		var m Mail
		_, err := t.Next(&m)
		if err == datastore.Done {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		mails[i] = map[string]string{
			"Text":     string(m.Text),
			"Received": m.Received.String(),
		}
		i++
	}
	t := template.Must(template.New("mails").Parse(mailHTML))
	t.Execute(w, map[string]interface{}{
		"To":    to,
		"Mails": mails,
	})
}
