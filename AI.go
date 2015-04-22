package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"runtime"
	"sync"
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

var scoreWeights = [][]uint{
	[]uint{48, 47, 46, 45},
	[]uint{22, 23, 24, 25},
	[]uint{11, 10, 9, 8},
	[]uint{1, 2, 3, 4},
	//[]uint{16, 15, 14, 13},
	//[]uint{12, 11, 10, 9},
	//[]uint{8, 7, 6, 5},
	//[]uint{4, 3, 2, 1},
	//[]uint{16, 12, 8, 4},
	//[]uint{12, 9, 6, 3},
	//[]uint{8, 6, 4, 2},
	//[]uint{4, 3, 2, 1},
	//[]int{4, 5, 12, 13}, []int{3, 6, 11, 14}, []int{2, 7, 10, 15}, []int{1, 8, 9, 16},
}

func (b Board) Score() int {
	var scores [4]int
	for k := 0; k < 4; k++ {
		sum := 0
		min := 1 << 20
		for i, row := range b.Cells {
			minrow := min
			maxrow := 0
			for _, c := range row {
				if maxrow < c.Value {
					maxrow = c.Value
				}

				if minrow > c.Value {
					minrow = c.Value
				}
			}

			if i > 0 && maxrow <= min {
				mini := 0
				for m, c := range b.Cells[i-1] {
					if c.Value == min {
						mini = m
						break
					}
				}

				rowscores := make([]uint, len(row))
				copy(rowscores, scoreWeights[i])

				tmp := rowscores[0]
				rowscores[0] = rowscores[mini]
				rowscores[mini] = tmp

				tmp1 := tmp - 1
				for j := mini - 1; j >= 0; j-- {
					rowscores[j] = tmp1
					tmp1 -= 1
				}

				tmp1 = tmp + 1
				for j := mini + 1; j < 4; j++ {
					rowscores[j] = tmp1
					tmp1 -= 1
				}

				for j, c := range row {
					if c.Value > 0 {
						sum += c.Value << rowscores[j]
					}
				}
			} else if i > 0 && maxrow > min {
				mini := 0
				for m, c := range b.Cells[i-1] {
					if c.Value == min {
						mini = m
						break
					}
				}

				rowscores := make([]uint, len(row))
				copy(rowscores, scoreWeights[i])

				tmp := rowscores[3]
				rowscores[3] = rowscores[mini]
				rowscores[mini] = tmp

				tmp1 := tmp - 1
				for j := mini - 1; j >= 0; j-- {
					rowscores[j] = tmp1
					tmp1 += 1
				}

				tmp1 = tmp + 1
				for j := mini + 1; j < 4; j++ {
					rowscores[j] = tmp1
					tmp1 += 1
				}

				for j, c := range row {
					if c.Value > 0 {
						sum += c.Value << rowscores[j]
					}
				}
			} else {
				sumrow := 0
				sumrow2 := 0

				for j, c := range row {
					if c.Value > 0 {
						//if c.Value < min {
						sumrow += c.Value << scoreWeights[i][j]
						sumrow2 += c.Value << scoreWeights[i][len(row)-j-1]
						//} else {
						//	sumrow -= c.Value<<(uint(i)*4) - (min << scoreWeights[i][j])
						//	sumrow2 -= c.Value<<(uint(i)*4) - (min << scoreWeights[i][len(row)-j-1])
						//}

						//if i > 0 {
						//	side := b.Cells[i-1][j].Value
						//	if side > c.Value {
						//		sumrow -= (side - c.Value) << scoreWeights[i][j]
						//		sumrow2 -= (side - c.Value) << scoreWeights[i][len(row)-j-1]
						//	} else {
						//		sumrow -= (2 * (c.Value - side)) << scoreWeights[i][j]
						//		sumrow2 -= (2 * (c.Value - side)) << scoreWeights[i][len(row)-j-1]
						//	}
						//}
					}
				}

				if sumrow > sumrow2 {
					sum += sumrow
				} else {
					sum += sumrow2
				}
			}

			if min > minrow {
				min = minrow
			}
		}

		scores[k] = sum

		b.rotateRight()
	}

	highest := 0
	for _, s := range scores {
		if highest < s {
			highest = s
		}
	}

	return highest
}

func (b Board) DeepCopy() Board {
	cp := Board{}
	cp.Cells = make([][]Tile, len(b.Cells))
	cp.Size = b.Size
	for i, row := range b.Cells {
		cp.Cells[i] = make([]Tile, len(row))
		for j, c := range row {
			cp.Cells[i][j] = c
		}
	}

	return cp
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

		//log.Printf("%+v\n", merged2)

		for j, _ := range row {
			if j < len(merged2) {
				row[j].Value = merged2[j]
			} else {
				row[j].Value = 0
			}
		}
	}

	//log.Printf("%+v\n", b)
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

func (b Board) AvaliableCount() int {
	avaliable := 0
	for _, row := range b.Cells {
		for _, c := range row {
			if c.Value == 0 {
				avaliable++
			}
		}
	}

	return avaliable
}

func (b Board) EachAddRandomTile(fn func(Board, float64)) {
	avaliable := float64(b.AvaliableCount())

	for i, row := range b.Cells {
		for j, c := range row {
			if c.Value == 0 {
				b1 := b.DeepCopy()
				b1.Cells[i][j].Value = 2
				fn(b1, (1.0/avaliable)*0.9)

				b1 = b.DeepCopy()
				b1.Cells[i][j].Value = 4
				fn(b1, (1.0/avaliable)*0.1)
			}
		}
	}
}

func NextMove(b Board, depth int) (int, float64) {
	var allscores [4]float64
	var canmove [4]bool
	for i := 0; i < 4; i++ {
		cp := b.DeepCopy()
		if cp.Move(i) {
			canmove[i] = true
			if depth > 1 {
				score2 := float64(0)
				w := sync.WaitGroup{}
				w.Add(2 * cp.AvaliableCount())
				cp.EachAddRandomTile(func(b1 Board, p float64) {
					go func() {
						_, s := NextMove(b1, depth-1)
						score2 += s * p
						w.Done()
					}()
				})
				w.Wait()

				allscores[i] = float64(cp.Score()) + score2
			} else {
				allscores[i] = float64(cp.Score())
			}
		} else {
			allscores[i] = -1
		}
	}

	if !canmove[0] && !canmove[1] && !canmove[2] && !canmove[3] {
		return 0, -(1 << 60)
	}

	var move int
	highest := -math.MaxFloat64
	for i, score := range allscores {
		if highest < score && canmove[i] {
			highest = score
			move = i
		}
	}

	//log.Printf("score: %+v\n", allscores)

	return move, highest
}

func main() {
	runtime.GOMAXPROCS(4)
	http.HandleFunc("/", AI)
	log.Fatal(http.ListenAndServe("localhost:8877", nil))
}

var loglist = []int{2, 4, 8, 16, 32, 64, 128, 256, 512,
	1024, 2048, 4096, 8192, 16384, 32768, 65536, 65536 * 2}

func log2(n int) {

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

	//log.Printf("%+v\n", state)
	cnt := state.AvaliableCount() //log.Println(jsonstate)
	if cnt == 1 {
		nextmove, _ := NextMove(state, 5)
		io.WriteString(w, fmt.Sprintf("move(%d);", nextmove))
	} else if cnt < 3 {
		nextmove, _ := NextMove(state, 4)
		io.WriteString(w, fmt.Sprintf("move(%d);", nextmove))
		//	} else if cnt < 7 {
		//		nextmove, _ := NextMove(state, 3)
		//		io.WriteString(w, fmt.Sprintf("move(%d);", nextmove))
	} else {
		nextmove, _ := NextMove(state, 3)
		io.WriteString(w, fmt.Sprintf("move(%d);", nextmove))
	}
	//log.Printf("next move:%d\n", nextmove)
	//time.Sleep(100 * time.Millisecond)
}
