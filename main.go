package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/arturo-source/poker-engine"
)

type handEquity map[poker.HandKind]uint

type equity struct {
	wins  uint
	ties  uint
	hands handEquity
}

func getCombinations(hands []poker.Cards, board poker.Cards) <-chan poker.Cards {
	combc := make(chan poker.Cards)
	handsJoined := poker.JoinCards(hands...)

	var allCombinations func(currComb poker.Cards, start int, n int)
	allCombinations = func(currComb poker.Cards, start int, n int) {
		if n == 0 {
			combc <- currComb.AddCards(board)
			return
		}

		for i := start; i < poker.MAX_CARDS; i++ {
			if board.HasBit(i) {
				continue
			}
			if handsJoined.HasBit(i) {
				continue
			}

			currComb = currComb.BitToggle(i)
			allCombinations(currComb, i+1, n-1)
			currComb = currComb.BitToggle(i)
		}
	}

	cardsInBoard := board.Count()
	go func() {
		allCombinations(poker.NO_CARD, 0, poker.MAX_CARDS_IN_BOARD-cardsInBoard)
		close(combc)
	}()

	return combc
}

func calculateEquities(hands []poker.Cards, board poker.Cards) (equities map[*poker.Player]equity, nCombinations uint) {
	equities = make(map[*poker.Player]equity)
	players := make([]*poker.Player, 0, len(hands))
	for _, hand := range hands {
		player := &poker.Player{Hand: hand}
		players = append(players, player)
		equities[player] = equity{hands: make(handEquity)}
	}

	for comb := range getCombinations(hands, board) {
		winners := poker.GetWinners(comb, players)
		justOneWinner := len(winners) == 1

		for _, winner := range winners {
			playerEquity := equities[winner.Player]

			playerEquity.hands[winner.HandKind]++
			if justOneWinner {
				playerEquity.wins++
			} else {
				playerEquity.ties++
			}

			equities[winner.Player] = playerEquity
		}

		nCombinations++
	}

	return
}

func parseUserInputs(handsStr, boardStr string) (hands []poker.Cards, board poker.Cards, err error) {
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
	var html string

	html += `<span>`
	html += `Error parsing input: `
	html += err.Error()
	html += `</span>`

	return html
}

//export getResultsInHTML
func getResultsInHTML(handsStr, boardStr string) string {
	hands, board, err := parseCommandLine()
	if err != nil {
		return getErrorInHTML(err)
	}

	var start = time.Now()
	equities, nCombinations := calculateEquities(hands, board)
	timeElapsed := time.Since(start)

	// printResults(board, equities, nCombinations, timeElapsed)
	// TODO: replicate what printResults func does, but with an HTML template
	return ""
}

func main() {}
