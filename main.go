package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

var displayLetters = [...]string{
	"qwertyuiop",
	"asdfghjkl",
	"zxcvbnm",
}

type letterState uint8

const (
	notGuessed letterState = iota
	found
	foundPosition
	notInWord
)

type guessLetter struct {
	letter string
	state  letterState
}

type game struct {
	word         string
	guesses      [][]guessLetter
	letterStates [26]letterState
	done         bool
	won					 bool
	err          error
}

func NewGame() game {
	return game{
		getWord(),
		make([][]guessLetter, 0),
		[26]letterState{},
		false,
		false,
		nil,
	}
}

func (g *game) guess(word string) bool {
	if len(g.guesses) == 6 {
		g.err = fmt.Errorf("out of guesses")
		g.done = true
		return false
	}

	if !valid(word) {
		g.err = fmt.Errorf("'%s' is not a valid word", word)
		return false
	}

	newGuess := make([]guessLetter, len(word))
	guessBytes := []byte(word)

	letterCounts := make(map[byte]int)
	for _, r := range g.word {
		b := byte(r)
		if count, present := letterCounts[b]; present {
			letterCounts[b] = count + 1
		} else {
			letterCounts[b] = 1
		}
	}

	// fill newGuess and find exact matches
	for index, b := range guessBytes {
		var ls letterState = notGuessed
		if g.word[index] == b {
			newGuess[index] = guessLetter{string(b), foundPosition}
			letterCounts[b]--
			ls = foundPosition
		}

		li := b - 'a'
		if ls > g.letterStates[li] {
			g.letterStates[li] = ls
		}
	}

	// find yellow matches, skipping greens
	for index, b := range guessBytes {
		var ls letterState = notGuessed

		if len(newGuess[index].letter) != 0 {
			// already found
			continue
		} else if letterCounts[b] == 0 {
			// none of this letter left
			newGuess[index] = guessLetter{string(b), notInWord}
			ls = notInWord
		} else if strings.Contains(g.word, string(b)) {
			// yellow
			newGuess[index] = guessLetter{string(b), found}
			letterCounts[b]--
			ls = found
		} else {
			// not in word
			newGuess[index] = guessLetter{string(b), notInWord}
			ls = notInWord
		}

		li := b - 'a'
		if ls > g.letterStates[li] {
			g.letterStates[li] = ls
		}
	}
	g.guesses = append(g.guesses, newGuess)

	if word == g.word {
		g.won = true
		g.done = true
	}

	return true
}

func green(s string) string {
	return fmt.Sprintf("\x1b[1m\x1b[42m\x1b[30m%s\x1b[0m", s)
}
func yellow(s string) string {
	return fmt.Sprintf("\x1b[1m\x1b[43m\x1b[30m%s\x1b[0m", s)
}
func grey(s string) string {
	return fmt.Sprintf("\x1b[1m\x1b[90m\x1b[30m%s\x1b[0m", s)
}

func (g *game) display() {
	// game board

	fmt.Println("┌───┬───┬───┬───┬───┐")

	for row := 0; row < 6; row++ {
		if len(g.guesses) > row {
			fmt.Print("│")
			for _, c := range g.guesses[row] {
				switch c.state {
				case foundPosition:
					fmt.Printf("%s|", green(" " + c.letter + " "))
				case found:
					fmt.Printf("%s|", yellow(" " + c.letter + " "))
				default:
					fmt.Printf(" %s |", c.letter)
				}
			}
			fmt.Print("\n")
		} else {
			fmt.Println("│   │   │   │   │   │")
		}
		if row < 5 {
			fmt.Println("├───┼───┼───┼───┼───┤")
		}
	}

	fmt.Println("└───┴───┴───┴───┴───┘")

	// letters
	for index, row := range displayLetters {
		for _, r := range row {
			li := letterState(r - 'a')
			switch g.letterStates[li] {
			case notGuessed:
				fmt.Printf("%c", r)
			case found:
				fmt.Print(yellow(string(r)))
			case foundPosition:
				fmt.Print(green(string(r)))
			case notInWord:
				fmt.Print(grey(string(r)))
			}
			fmt.Print(" ")
		}
		fmt.Print("\n")

		for i := 0; i < index+1; i++ {
			fmt.Print(" ")
		}
	}
	fmt.Print("\n")
}

func (g *game) prompt() {
	fmt.Printf("guess %d/6: ", len(g.guesses)+1)
}

func getWord() string {
	index := rand.Intn(len(wordsToGuess))
	return wordsToGuess[index]
}

func valid(word string) bool {
	for _, candidate := range wordsToGuess {
		if word == candidate {
			return true
		}
	}

	for _, candidate := range allWords {
		if word == candidate {
			return true
		}
	}
	return false
}

func main() {
	rand.Seed(time.Now().UnixNano())

	g := NewGame()
	g.display()

	scanner := bufio.NewScanner(os.Stdin)
	g.prompt()
	for scanner.Scan() {
		if !g.guess(scanner.Text()) {
			fmt.Println(g.err)
		}
		g.display()
		if g.done {
			break
		}
		g.prompt()
	}

	if g.won {
		fmt.Println("you win!")
		for _, g := range g.guesses {
			for _, gl := range g {
				switch gl.state {
				case foundPosition:
					fmt.Print("🟩")
				case found:
					fmt.Print("🟨")
				case notInWord:
					fmt.Print("⬛")
				}
			}
			fmt.Print("\n")
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
