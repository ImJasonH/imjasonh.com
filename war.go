package imjasonh

import (
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
)

var games, warCards, values, suits, numCards int

const warHTML = `
<html><head>
<title>{{.NumGames}} Games of War</title>

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

type gameResult int

const (
	// Results of games
	p1win gameResult = iota
	p2win
	tie
)

func init() {
	http.HandleFunc("/war", war)
}

func defaultAndInt(val string, def int) int {
	if val != "" {
		i, _ := strconv.ParseInt(val, 0, 16)
		return int(i)
	} else {
		return def
	}
	panic("Unreachable")
}

func pop(in []int) (int, []int) {
	if len(in) == 1 {
		return in[0], []int{}
	} else {
		return in[0], in[1:]
	}
	panic("Unreachable")
}

func playGame() gameResult {
	values := 13
	suits := 3
	numCards := values * suits

	allCards := rand.Perm(numCards)
	p1deck := allCards[numCards/2:]
	p2deck := allCards[:numCards/2]

	for len(p1deck) > 0 && len(p2deck) > 0 {
		var p1card, p2card int
		p1card, p1deck = pop(p1deck)
		p2card, p2deck = pop(p2deck)

		p1val := p1card % values
		p2val := p2card % values

		switch {
		case p1val == p2val:
			// TODO: Actually do war
		case p1val > p2val:
			p1deck = append(p1deck, p1card, p2card)
		case p2val > p1val:
			p2deck = append(p2deck, p1card, p2card)
		}
	}

	switch {
	case len(p1deck) > 0 && len(p2deck) == 0:
		return p1win
	case len(p2deck) > 0 && len(p1deck) == 0:
		return p2win
	default:
		return tie
	}
	panic("Unreachable")
}

// war simulates playing a number of games of the card game War.
func war(w http.ResponseWriter, r *http.Request) {
	games = defaultAndInt(r.FormValue("games"), 100)
	//warCards = defaultAndInt(r.FormValue("warCards"), 3)
	values = defaultAndInt(r.FormValue("values"), 13)
	suits = defaultAndInt(r.FormValue("suits"), 4)
	numCards = values * suits
	var p1wins, p2wins, ties int

	// TODO: Actually implement total score counter in datastore.
	allP1wins := -1
	allP2wins := -1

	for game := 0; game < games; game++ {
		// TODO: Do this in a goroutine.
		gameResult := playGame()
		switch {
		case gameResult == p1win:
			p1wins++
		case gameResult == p2win:
			p2wins++
		case gameResult == tie:
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
