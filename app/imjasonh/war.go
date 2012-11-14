package imjasonh

import (
	"appengine"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
)

const (
	warHTML = `
<html><head>
<meta http-equiv="Content-Type" content="text/html; charset=ISO-8859-1" />
<title>100 Games of War</title>

<style>
.title { font-size: 75px; }
.score { font-size: 150px; font-weight: bold; }
.red { color: #ff0000; }
.blue { color: #0000ff; }
table {
  margin-left: auto;
  margin-right: auto;
  margin-top: 100px; font-family : arial;
  text-align: center;
  font-family: arial;
}
</style></head><body>
<table>
  <tr><td class="title" colspan="3">{{.NumGames}} Games of War</td></tr>
  <tr>
    <td class="score red">{{.P1Wins}}</td>
    <td>
      <img src="http://chart.apis.google.com/chart?chs=200x200&chd=t:{{.P1Wins}},{{.P2Wins}},{{.Ties}}&cht=p&chco=ff0000,0000ff,ffffff" />
    </td>
    <td class="score blue">{{.P2Wins}}</td>
  </tr>
  <tr>
    <td>{{.AllP1Wins}}</td>
    <td>&laquo; All time wins &raquo;</td>
    <td>{{.AllP2Wins}}</td>
  </tr>
  <tr>
    <td></td>
    <td><small><a href="/">refresh</a> | <a href="/about">what?</a> | <a href="https://imjasonh.googlecode.com/git/app/imjasonh/war.go">src</a></small></td>
    <td></td>
  </tr>
</table></body></html>`
)

func init() {
	http.HandleFunc("/war", war)
}

func defaultAndInt(val string, def int) int {
	if val == "" {
		i := int64(0)
		i, _ = strconv.ParseInt(val, 0, 16)
		return int(i)
	} else {
		return def
	}
	return -1
}

// war simulates playing a number of games of the card game War.
func war(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	games := defaultAndInt(r.FormValue("games"), 100)
	//warCards := defaultAndInt(r.FormValue("warCards"), 3)
	values := defaultAndInt(r.FormValue("values"), 13)
	suits := defaultAndInt(r.FormValue("suits"), 4)
	numCards := values * suits
	var p1wins, p2wins, ties int

	// TODO: Actually implement card-playing logic.
	allP1wins := 100
	allP2wins := 200

	for game := 0; game < games; game++ {
		c.Infof("Game", game)
		allCards := rand.Perm(numCards)
		half := numCards / 2
		p1cards := allCards[:half]
		p2cards := allCards[half:]

		// TODO: Actually implement War logic.
		if p1cards[0] > p2cards[0] {
			p1wins++
		} else if p2cards[0] > p1cards[0] {
			p2wins++
		} else {
			ties++
		}
	}

	t := template.Must(template.New("war").Parse(warHTML))
	t.Execute(w, map[string]int{
		"P1Wins":    p1wins,
		"P2Wins":    p2wins,
		"Ties":      ties,
		"NumGames":  games,
		"AllP1Wins": allP1wins,
		"AllP2Wins": allP2wins,
	})
}
