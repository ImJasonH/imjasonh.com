package imjasonh

import (
	"appengine"
	"appengine/datastore"
	"appengine/taskqueue"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	delay = 2 * time.Hour
	limit = 100

	// TODO: Add a button to immediately delete a message.
	mailHTML = `<html><body>
<h3>Mails to: {{.To}}</h3>
<table border="1">
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

	explainHTML = `<html><body>
  <h3>What is this?</h3>
  <p>Send an email to <b><i>anything</i>@imjasonh-hrd.appspotmail.com</b>, then visit <a href="/mail/anything">/mail/anything</a> to see the emails it has received.</p>
  <p>This is useful for debugging sending email, and also for signing up for spammy services that require email account authentication.</p>
  <p>This service is *public* and *not at all secure or reliable*. Please don't use this for anything serious, ever. I mean it.</p>
</body></html>`
)

func init() {
	http.HandleFunc("/_ah/mail/", inbound)
	http.HandleFunc("/_ah/queue/reapMail", reapMail)
	http.HandleFunc("/mail", explain)
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
	to := r.URL.Path[len("/mail/"):] + "@imjasonh-hrd.appspotmail.com"

	q := datastore.NewQuery("Mail").
		Filter("To =", to).
		Order("-Received").
		Limit(limit)
	cnt, err := q.Count(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mails := make([]map[string]string, cnt)
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

// explain explains this feature.
func explain(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, explainHTML)
}
