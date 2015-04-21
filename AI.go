package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

type Tile struct {
	//Position struct {
	//	X int `json:"x"`
	//	Y int `json:"y"`
	//} `json:"position"`

	Value int `json:"value"`
}

type Board struct {
	Size  int      `json:"size"`
	Cells [][]Tile `json:"cells"`
}

var scoreShifts = [][]int{
	[]int{1, 2, 3, 4},
	[]int{8, 7, 6, 5},
	[]int{9, 10, 11, 12},
	[]int{16, 15, 14, 13},
	//[]int{4, 5, 12, 13}, []int{3, 6, 11, 14}, []int{2, 7, 10, 15}, []int{1, 8, 9, 16},
}

func (b Board) Score() int {
	sum := 0
	for i, row := range b.Cells {
		sumrow := 0
		sumrow2 := 0
		for j, c := range row {
			sumrow += c.Value << uint(scoreShifts[i][j])
			sumrow2 += c.Value << uint(scoreShifts[i][len(row)-j-1])
		}

		if sumrow > sumrow2 {
			sum += sumrow
		} else {
			sum += sumrow2
		}
	}

	return sum
}

func (b Board) DeepCopy() Board {
	bbytes, err := json.Marshal(&b)
	if err != nil {
		log.Fatal(err)
	}

	cloned := Board{}
	if err := json.Unmarshal(bbytes, &cloned); err != nil {
		log.Fatal(err)
	}
	return cloned
}

func (b Board) MoveLeft() bool {
	before := b.DeepCopy()
	for _, row := range b.Cells {
		var merged [4]int
		if row[0].Value == row[1].Value {
			merged[0] = 2 * row[0].Value
			if row[2].Value == row[3].Value {
				merged[1] = 2 * row[2].Value
			} else {
				merged[1] = row[2].Value
				merged[2] = row[3].Value
			}
		} else if row[1].Value == row[2].Value {
			merged[0] = row[0].Value
			merged[1] = 2 * row[1].Value
			merged[2] = row[3].Value
		} else if row[2].Value == row[3].Value {
			merged[0] = row[0].Value
			merged[1] = row[1].Value
			merged[2] = row[2].Value * 2
		} else {
			merged[0] = row[0].Value
			merged[1] = row[1].Value
			merged[2] = row[2].Value
			merged[3] = row[3].Value
		}

		var merged2 []int
		for _, v := range merged {
			if v != 0 {
				merged2 = append(merged2, v)
			}
		}

		log.Printf("%+v\n", merged2)

		for j, _ := range row {
			if j < len(merged2) {
				row[j].Value = merged2[j]
			} else {
				row[j].Value = 0
			}
		}
	}

	log.Printf("%+v\n", b)
	return !b.Equal(before)
}

func (b Board) MoveDown() bool {
	b.rotateRight()
	canmove := b.MoveLeft()
	b.rotateLeft()
	return canmove
}

func (b Board) MoveRight() bool {
	b.rotateRight()
	b.rotateRight()
	canmove := b.MoveLeft()
	b.rotateRight()
	b.rotateRight()
	return canmove
}

func (b Board) MoveUp() bool {
	b.rotateLeft()
	canmove := b.MoveLeft()
	b.rotateRight()
	return canmove
}

func (b Board) Equal(a Board) bool {
	abytes, err := json.Marshal(&a)
	if err != nil {
		log.Fatal(err)
	}
	bbytes, err := json.Marshal(&b)
	if err != nil {
		log.Fatal(err)
	}
	return bytes.Equal(abytes, bbytes)
}

func (b Board) rotateRight() {
	var values [4][4]int
	for i, row := range b.Cells {
		for j, c := range row {
			values[i][j] = c.Value
		}
	}

	for i, row := range values {
		for j, _ := range row {
			b.Cells[i][j].Value = values[len(row)-j-1][i]
		}
	}
}

func (b Board) rotateLeft() {
	b.rotateRight()
	b.rotateRight()
	b.rotateRight()
}

func (b Board) Move(dir int) bool {
	switch dir {
	case 0:
		return b.MoveLeft()
	case 1:
		return b.MoveUp()
	case 2:
		return b.MoveRight()
	case 3:
		return b.MoveDown()
	default:
		log.Fatal("unknown direction")
		return false
	}
}

func NextMove(b Board, depth int) (int, int) {
	var allscores [4]int

	for i := 0; i < 4; i++ {
		copy := b.DeepCopy()
		if copy.Move(i) {
			if depth > 1 {
				_, score2 := NextMove(copy.DeepCopy(), depth-1)
				allscores[i] = copy.Score() + score2
			} else {
				allscores[i] = copy.Score()
			}
		} else {
			allscores[i] = -1
		}
	}

	var move int
	highest := -1
	for i, score := range allscores {
		if highest < score {
			highest = score
			move = i
		}
	}

	log.Printf("score: %+v\n", allscores)

	return move, highest
}

func main() {
	http.HandleFunc("/", AI)
	log.Fatal(http.ListenAndServe("localhost:8877", nil))
}

func AI(w http.ResponseWriter, req *http.Request) {
	jsonstate, err := url.QueryUnescape(req.URL.RawQuery)
	if err != nil {
		log.Print(err)
		return
	}

	state := Board{}
	if err := json.Unmarshal([]byte(jsonstate), &state); err != nil {
		log.Print(err)
		return
	}
	//flip
	cells := make([][]Tile, 4)
	copy(cells, state.Cells)
	for i, row := range cells {
		state.Cells[len(cells)-i-1] = row
	}

	log.Printf("%+v\n", state)
	log.Println(jsonstate)
	nextmove, _ := NextMove(state, 2)
	log.Printf("next move:%d\n", nextmove)
	//time.Sleep(time.Millisecond)
	io.WriteString(w, fmt.Sprintf("move(%d);", nextmove))
}
