package cmd

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// ── Board ─────────────────────────────────────────────────────────────────────

const (
	bW      = 50 // board width
	bH      = 16 // board height
	canCol  = 8  // player column
	canW    = 6  // can width
	canH    = 3  // can height (rows)
	pipeW   = 5  // pipe width
	frameMs = 55 // ms per frame
	flapV   = -1.5
	gravV   = 0.15
	pipeSp  = 0.7
)

// ── Can art ───────────────────────────────────────────────────────────────────

var canArt = [3]string{
	".----.",
	"|soda|",
	"'----'",
}

var canDeadArt = [3]string{
	".----.",
	"|XXXX|",
	"'----'",
}

// ── Types ─────────────────────────────────────────────────────────────────────

type fpipe struct {
	x      float64
	gapTop int  // row where gap starts
	scored bool // already scored
}

type fgame struct {
	y, vy   float64 // can vertical pos and velocity
	pipes   []fpipe
	score   int
	best    int
	dead    bool
	started bool
	tick    int
	rng     *rand.Rand
}

func newFGame(best int) *fgame {
	return &fgame{
		y:    float64(bH/2 - 1),
		best: best,
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (g *fgame) gapH() int {
	h := 7 - g.score/8
	if h < 5 {
		h = 5
	}
	return h
}

func (g *fgame) flap() {
	if !g.dead {
		g.vy = flapV
		g.started = true
	}
}

func (g *fgame) step() {
	if g.dead || !g.started {
		return
	}
	g.tick++

	// physics
	g.vy += gravV
	g.y += g.vy

	// ceiling
	if g.y < 0 {
		g.y = 0
		g.vy = 0
	}
	// floor
	if int(g.y)+canH > bH {
		g.dead = true
		return
	}

	// move pipes
	for i := range g.pipes {
		g.pipes[i].x -= pipeSp
	}
	keep := g.pipes[:0]
	for _, p := range g.pipes {
		if p.x > -float64(pipeW)-1 {
			keep = append(keep, p)
		}
	}
	g.pipes = keep

	// spawn
	spacing := 28.0 - float64(g.score)/5.0
	if spacing < 18 {
		spacing = 18
	}
	spawn := len(g.pipes) == 0
	if !spawn {
		last := g.pipes[len(g.pipes)-1]
		if float64(bW)-last.x >= spacing {
			spawn = true
		}
	}
	if spawn {
		gap := g.gapH()
		minTop := 1
		maxTop := bH - gap - 1
		if maxTop < minTop {
			maxTop = minTop
		}
		top := minTop + g.rng.Intn(maxTop-minTop+1)
		g.pipes = append(g.pipes, fpipe{x: float64(bW), gapTop: top})
	}

	// score
	for i := range g.pipes {
		if !g.pipes[i].scored && int(g.pipes[i].x)+pipeW < canCol {
			g.pipes[i].scored = true
			g.score++
			if g.score > g.best {
				g.best = g.score
			}
		}
	}

	// collision
	g.collide()
}

func (g *fgame) collide() {
	cT := int(g.y)
	cB := cT + canH - 1
	cL := canCol
	cR := canCol + canW - 1

	gap := g.gapH()
	for _, p := range g.pipes {
		pL := int(p.x)
		pR := pL + pipeW - 1
		if cR < pL || cL > pR {
			continue
		}
		// safe if entirely inside gap
		if cT >= p.gapTop && cB < p.gapTop+gap {
			continue
		}
		g.dead = true
		return
	}
}

func (g *fgame) draw() []byte {
	grid := make([][]byte, bH)
	for r := range grid {
		grid[r] = make([]byte, bW)
		for c := range grid[r] {
			grid[r][c] = ' '
		}
	}

	gap := g.gapH()

	// draw pipes
	for _, p := range g.pipes {
		px := int(p.x)
		// top section
		for r := 0; r < p.gapTop; r++ {
			drawPipeRow(grid, r, px, false)
		}
		// top cap (wider lip at gap edge)
		if p.gapTop > 0 {
			drawPipeCap(grid, p.gapTop-1, px)
		}
		// bottom cap
		botStart := p.gapTop + gap
		if botStart < bH {
			drawPipeCap(grid, botStart, px)
		}
		// bottom section
		for r := botStart; r < bH; r++ {
			drawPipeRow(grid, r, px, false)
		}
	}

	// draw can
	art := canArt
	if g.dead {
		art = canDeadArt
	}
	cT := int(g.y)
	for r, line := range art {
		row := cT + r
		if row >= 0 && row < bH {
			for c := 0; c < len(line); c++ {
				col := canCol + c
				if col >= 0 && col < bW {
					grid[row][col] = line[c]
				}
			}
		}
	}

	// build frame
	var b strings.Builder
	b.WriteString("\033[H") // cursor home

	// header
	b.WriteString(fmt.Sprintf("\r\n  Score: %-6d Best: %d\033[K\r\n", g.score, g.best))

	// top border
	b.WriteString("  +")
	b.WriteString(strings.Repeat("-", bW))
	b.WriteString("+\033[K\r\n")

	// game rows
	for r := 0; r < bH; r++ {
		b.WriteString("  |")
		b.Write(grid[r])
		b.WriteString("|\033[K\r\n")
	}

	// bottom border
	b.WriteString("  +")
	b.WriteString(strings.Repeat("-", bW))
	b.WriteString("+\033[K\r\n")

	// status
	b.WriteString("\r\n")
	if g.dead {
		b.WriteString("  GAME OVER!  R = restart  Q = quit\033[K\r\n")
	} else if !g.started {
		b.WriteString("  SPACE = flap    Q = quit\033[K\r\n")
	} else {
		b.WriteString("  SPACE = flap    Q = quit\033[K\r\n")
	}
	b.WriteString("\033[K")

	return []byte(b.String())
}

func drawPipeRow(grid [][]byte, row, px int, _ bool) {
	if row < 0 || row >= bH {
		return
	}
	for c := 0; c < pipeW; c++ {
		col := px + c
		if col >= 0 && col < bW {
			if c == 0 || c == pipeW-1 {
				grid[row][col] = '|'
			} else {
				grid[row][col] = '#'
			}
		}
	}
}

func drawPipeCap(grid [][]byte, row, px int) {
	if row < 0 || row >= bH {
		return
	}
	for c := -1; c <= pipeW; c++ {
		col := px + c
		if col >= 0 && col < bW {
			if c == -1 || c == pipeW {
				grid[row][col] = '+'
			} else {
				grid[row][col] = '='
			}
		}
	}
}

// ── Command ───────────────────────────────────────────────────────────────────

var canCmd = &cobra.Command{
	Use:    "can",
	Short:  "...",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return playCan()
	},
}

func playCan() error {
	fd := int(os.Stdin.Fd())
	old, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("terminal error: %w", err)
	}
	defer term.Restore(fd, old)

	w := os.Stdout
	w.WriteString("\033[2J\033[H\033[?25l") // clear + hide cursor
	defer w.WriteString("\033[?25h\033[2J\033[H")

	g := newFGame(0)

	// draw initial frame before input loop
	w.Write(g.draw())

	// input
	keys := make(chan byte, 32)
	go func() {
		buf := make([]byte, 8)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				return
			}
			for i := 0; i < n; i++ {
				keys <- buf[i]
			}
		}
	}()

	ticker := time.NewTicker(frameMs * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case b := <-keys:
			switch b {
			case ' ':
				g.flap()
			case 'r', 'R':
				if g.dead {
					g = newFGame(g.best)
				}
			case 'q', 'Q', 3:
				return nil
			case 27: // esc prefix
				select {
				case b2 := <-keys:
					if b2 == '[' {
						select {
						case b3 := <-keys:
							if b3 == 'A' {
								g.flap()
							}
						default:
						}
					}
				default:
				}
			}

		case <-ticker.C:
			g.step()
			w.Write(g.draw())
		}
	}
}

func init() {
	rootCmd.AddCommand(canCmd)
}
