/*
Package hangman is an implementation of the game Hangman many of us played
as a child.

Author: Justin Cook <jhcook@gmail.com>
*/

package hangman

import (
	"encoding/json"
	"github.com/jhcook/game_engine/dictionary"
	"github.com/jhcook/game_engine/util"
	"log"
	"math/rand"
	"net/http"
	"time"
	"unicode/utf8"
)

var theBoys *men = nil
var dict *dictionary.Dictionary = nil

// This is used to represent each game
type hangman struct {
	Word   string
	WrdUni []rune
	Right  []rune
	Wrong  []rune
	Game   uint64
	TwoP   bool
	P1cred string
	P2cred string
	Timer  *time.Timer
}

// This is used as the structure for JSON decoding
type Message struct {
	Cmd  string
	Gid  uint64
	Play string
	Auth string
}

// This is a structure that holds references to each game instance
type men struct {
	Games   uint64
	Episode map[uint64]*hangman
}

// Instantiates 'theBoys' if not already instantiated as this is intended to
// be singleton.
func NewMen() {
	if theBoys != nil {
		log.Println("NewMen: already instantiated")
		return
	}
	dict = dictionary.NewDictionary("")
	theBoys = &men{Episode: make(map[uint64]*hangman)}
}

// Create a new game of hangman, add it to the singleton 'theBoys' and return
// to the caller.
func NewHangman(np int) *hangman {
	var p2 string
	dict.Ci = uint32(rand.Intn(len(dict.Words)))
	wrd := dict.NextWord()
	theBoys.Games++
	if np == 2 {
		p2 = util.Rand_str(64)
	} else {
		p2 = ""
	}
	game := &hangman{Word: wrd,
		WrdUni: util.StringToRuneArray(wrd),
		Right:  make([]rune, len(wrd)),
		Wrong:  make([]rune, 0, 256),
		Game:   theBoys.Games,
		P1cred: util.Rand_str(64),
		P2cred: p2,
		Timer:  time.NewTimer(time.Second * 60)}
	go func() {
		<-game.Timer.C
		log.Println("expired:", game.Game)
		if _, ok := theBoys.Episode[game.Game]; ok {
			delete(theBoys.Episode, game.Game)
		}
	}()
	theBoys.Episode[theBoys.Games] = game
	return game
}

func (g *hangman) evalChar(chr string) bool {
	var correct bool = false
	// Convert chr to rune
	letter, _ := utf8.DecodeRuneInString(chr)
	// See if rune is in the word
	for i, v := range g.WrdUni {
		if letter == v {
			g.Right[i] = letter
			correct = true
		}
	}
	if !correct {
		g.Wrong = append(g.Wrong, letter)
	}
	return correct
}

func (g *hangman) checkAuth(gid uint64, cred string) *hangman {
	if game, ok := theBoys.Episode[gid]; ok {
		if game.P1cred == cred {
			// On each play, reset the timer
			game.Timer.Reset(time.Second * 300)
			return game
		}
	}
	return nil
}

func Play(w http.ResponseWriter, r *http.Request) {
	var msg Message
	var game *hangman
	var ok bool = false
	var answer interface{} = nil
	// Decode the JSON
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&msg); err != nil {
		bytes, _ := json.Marshal(err)
		log.Println("hangman.Play.Decode:", err)
		w.Write(bytes)
		return
	}
	if len(msg.Cmd) > 0 && msg.Gid > 0 && len(msg.Play) > 0 && len(msg.Auth) > 0 {
		if game = game.checkAuth(msg.Gid, msg.Auth); game != nil {
			if msg.Cmd == "P1T" {
				game.evalChar(msg.Play)
			}
		} else {
			answer = struct {
				Error string
			}{"unauthorized"}
		}
	} else if len(msg.Cmd) > 0 && msg.Gid > 0 && len(msg.Auth) > 0 {
		if _, ok = theBoys.Episode[msg.Gid]; ok {
			if game = game.checkAuth(msg.Gid, msg.Auth); game != nil {
				if msg.Cmd == "STATUS" {
					log.Println(game)
				} else {
					answer = struct {
						Error string
					}{"game does not exist"}
				}
			} else {
				answer = struct {
					Error string
				}{"unauthorized"}
			}
		} else {
			answer = struct {
				Error string
			}{"unknown command"}
		}
	} else if len(msg.Cmd) > 0 {
		if msg.Cmd == "NEW" {
			game = NewHangman(1)
			answer = struct {
				Curr   []rune
				Missed []rune
				Game   uint64
				Cred   string
			}{game.Right, game.Wrong, game.Game, game.P1cred}
		} else {
			answer = struct {
				Error string
			}{"unknown command"}
		}
	} else {
		return
	}
	if answer == nil {
		// Use an anonymous structure to sanitize the data sent back.
		answer = struct {
			Word   string
			Curr   []rune
			Missed []rune
			Game   uint64
		}{game.Word, game.Right, game.Wrong, game.Game}
	}
	// Send the result down the wire.
	if bytes, err := json.Marshal(answer); err == nil {
		log.Println(string(bytes))
		w.Write(bytes)
	} else {
		log.Println(err)
	}
}
