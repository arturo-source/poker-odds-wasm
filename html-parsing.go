package main

import (
	"bytes"
	"fmt"
	"html/template"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/arturo-source/poker-engine"
)

const (
	NoSuit        = "#FFD700"
	SpadesColor   = "#000000"
	ClubsColor    = "#008000"
	HeartsColor   = "#FF4500"
	DiamondsColor = "#4169E1"
)

func colorize(txt string, color string) string {
	return fmt.Sprintf(`<span style="color: %s;">%s</span>`, color, txt)
}

func colorizeCards(cards poker.Cards) template.HTML {
	colorizeSuit := func(suit poker.Cards, color string) string {
		var cardsStr string

		suitStr := poker.SUIT_VALUES[suit]
		cardsSuited := cards & suit

		for cardNum, cardNumStr := range poker.NUMBER_VALUES {
			card := cardsSuited & cardNum
			if card != poker.NO_CARD {
				cardsStr += colorize(cardNumStr+suitStr, color)
			}
		}

		return cardsStr
	}

	spadesStr := colorizeSuit(poker.SPADES, SpadesColor)
	clubsStr := colorizeSuit(poker.CLUBS, ClubsColor)
	heartsStr := colorizeSuit(poker.HEARTS, HeartsColor)
	diamondsStr := colorizeSuit(poker.DIAMONDS, DiamondsColor)

	return template.HTML(spadesStr + clubsStr + heartsStr + diamondsStr)
}

func parseUserInputs(handsStr, boardStr string) (hands []poker.Cards, board poker.Cards, err error) {
	handsStr = strings.TrimSpace(handsStr)
	boardStr = strings.TrimSpace(boardStr)
	// Read all Args input and transform them into cards
	var allCards []poker.Cards
	if len(handsStr) == 0 {
		err = fmt.Errorf("at least one hand is needed")
		return
	}

	if len(boardStr) > 10 {
		err = fmt.Errorf("maximum cards in board are 5")
		return
	}

	handsStrArray := strings.Split(handsStr, " ")
	for _, handStr := range handsStrArray {
		if len(handStr) != 4 {
			err = fmt.Errorf("%s hand is not valid, hands must have 2 cards with a valid suit", colorize(handStr, NoSuit))
			return
		}

		firstCardStr, secondCardStr := handStr[:2], handStr[2:]
		firstCard, secondCard := poker.NewCard(firstCardStr), poker.NewCard(secondCardStr)
		if firstCard == poker.NO_CARD {
			err = fmt.Errorf("%s card (%s hand) is not valid", colorize(firstCardStr, NoSuit), colorize(handStr, NoSuit))
			return
		}
		if secondCard == poker.NO_CARD {
			err = fmt.Errorf("%s card (%s hand) is not valid", colorize(secondCardStr, NoSuit), colorize(handStr, NoSuit))
			return
		}

		hand := poker.JoinCards(firstCard, secondCard)
		hands = append(hands, hand)

		allCards = append(allCards, firstCard, secondCard)
	}

	// Read --board input and transform them into cards
	for i := 0; i < len(boardStr); i += 2 {
		end := i + 2
		if end > len(boardStr) {
			end = len(boardStr)
		}

		cardStr := boardStr[i:end]
		card := poker.NewCard(cardStr)
		if card == poker.NO_CARD {
			err = fmt.Errorf("%s card (%s board) is not valid", colorize(cardStr, NoSuit), colorize(boardStr, NoSuit))
			return
		}

		board = board.AddCards(card)

		allCards = append(allCards, card)
	}

	// Check if any card is repeated
	allCardsJoined := poker.JoinCards(allCards...)
	for _, card := range allCards {
		if !allCardsJoined.CardsArePresent(card) {
			err = fmt.Errorf("card %s is duplicated", colorizeCards(card))
			return
		}

		allCardsJoined = allCardsJoined.QuitCards(card)
	}

	return hands, board, err
}

func getErrorInHTML(err error) string {
	const html = `<span>Error parsing input: {{.}}</span>`
	buf := new(bytes.Buffer)

	t, _ := template.New("error").Parse(html)
	errorHtml := template.HTML(err.Error())
	t.Execute(buf, errorHtml)

	return buf.String()
}

const resultsTemplate = `
{{if .board}}
<table>
	<tbody>
		<tr><td>board:</td><td>{{colorizeCards .board}}</td></tr>
	</tbody>
</table>
{{end}}
<table>
	<thead>
		<th>hand</th><th>win</th><th>tie</th>
	</thead>
	<tbody>
		{{range .orderedPlayers}}
		{{$eq := index $.equities .}}
		<tr>
			<td>{{colorizeCards .Hand}}</td> <td>{{printf "%.1f" (percentage $eq.Wins $.nCombinations)}}%</td> <td>{{printf "%.1f" (percentage $eq.Ties $.nCombinations)}}%</td>
		</tr>
		{{end}}
	</tbody>
</table>
<table>
	<thead>
		<th></th>{{range .orderedPlayers}}<th>{{colorizeCards .Hand}}</th> {{end}}
	</thead>
	<tbody>
		{{range $hk := .handKinds}}
		<tr>
			<td>{{$hk}}</td>
			{{range $.orderedPlayers}}
			{{$eq := index $.equities .}}
			{{$handEqPercentage := (percentage (index $eq.Hands $hk) (sum $eq.Wins $eq.Ties) )}}

			{{if or (isNaN $handEqPercentage) (eq $handEqPercentage 0.0)}}
			<td>.</td>
			{{else if lt $handEqPercentage 0.1}}
			<td>>0.1%</td>
			{{else}}
			<td>{{printf "%.1f" $handEqPercentage}}%</td>
			{{end}}

			{{end}}
		</tr>
		{{end}}
	</tbody>
</table>
<p>{{.nCombinations}} combinations calculated in {{.timeElapsed}}</p>
`

func getResultsInHTML(handsStr, boardStr string) string {
	hands, board, err := parseUserInputs(handsStr, boardStr)
	if err != nil {
		return getErrorInHTML(err)
	}

	var start = time.Now()
	equities, nCombinations := calculateEquities(hands, board)
	timeElapsed := time.Since(start)

	buf := new(bytes.Buffer)
	t, err := template.New("results").Funcs(
		template.FuncMap{
			"colorizeCards": colorizeCards,
			"isNaN":         math.IsNaN,
			"percentage": func(n, total uint) float64 {
				return float64(n) / float64(total) * 100
			},
			"sum": func(nums ...uint) uint {
				var sum uint
				for _, num := range nums {
					sum += num
				}

				return sum
			},
		}).Parse(resultsTemplate)

	var orderedPlayers = make([]*poker.Player, 0, len(equities))
	for player := range equities {
		orderedPlayers = append(orderedPlayers, player)
	}
	sort.Slice(orderedPlayers, func(i, j int) bool {
		return equities[orderedPlayers[i]].Wins > equities[orderedPlayers[j]].Wins
	})
	handKinds := []poker.HandKind{poker.HIGHCARD, poker.PAIR, poker.TWOPAIR, poker.THREEOFAKIND, poker.STRAIGHT, poker.FLUSH, poker.FULLHOUSE, poker.FOUROFAKIND, poker.STRAIGHTFLUSH, poker.ROYALFLUSH}

	t.Execute(buf, map[string]any{
		"board":          board,
		"equities":       equities,
		"orderedPlayers": orderedPlayers,
		"handKinds":      handKinds,
		"nCombinations":  nCombinations,
		"timeElapsed":    timeElapsed,
	})
	return buf.String()
}
