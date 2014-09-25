/*
Package dictionary provides an interface to a dictionary with fancy features
such as complex letter frequency.

Author: Justin Cook <jhcook@gmail.com>
*/

package dictionary

import (
	"bufio"
	"log"
	"os"
)

type Dictionary struct {
	Word  string
	Ci    uint32
	Words []string
}

type MyString string

type MyRune []rune

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalln("dictionary.readLines:", err)
		return nil, err
	}
	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func NewDictionary(path string) *Dictionary {
	dict := new(Dictionary)
	dict.Ci = 1000
	if len(path) == 0 {
		path = "/usr/share/dict/words"
	}
	log.Println("dictionary.NewDictionary:", path)
	dict.Words, _ = readLines(path)
	return dict
}

func (d *Dictionary) NextWord() string {
	d.Ci++
	d.Word = d.Words[d.Ci]
	return d.Word
}