package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/views"
	"github.com/mattn/go-runewidth"
)

type root struct {
	parent *views.Application
	main   *views.CellView
	model  *model
	status *views.SimpleStyledTextBar
	title  *views.TextBar
	views.Panel
}

func (root *root) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyEscape:
			root.parent.Quit()
			return true
		case tcell.KeyEnter:
			root.model.y++
			root.updateKeys()
			return true
		case tcell.KeyRune:
			switch ev.Rune() {
			case 'Q', 'q':
				root.parent.Quit()
				return true
			case 'S', 's':
				root.model.y = root.model.y + 10000
				root.updateKeys()
				return true
			case 'V', 'v':
				root.model.y = root.model.y + 10
				root.updateKeys()
				return true
			case 'P', 'p':
				root.model.y = root.model.y - 1
				root.updateKeys()
				return true
			case 'H', 'h':
				root.SetTitle(root.title)
				root.updateKeys()
				return true
			case 'E', 'e':
				root.SetStatus(root.status)
				root.model.enab = true
				return true
			case 'D', 'd':
				root.RemoveWidget(root.title)
				root.RemoveWidget(root.status)
				root.model.enab = false
				return true
			}
		}
	}
	return true
}

type model struct {
	x        int
	y        int
	endx     int
	endy     int
	hide     bool
	enab     bool
	line     int
	lineRune []rune
	loc      string
	block    [][]byte
}

func (m *model) GetBounds() (int, int) {
	return m.endx, m.endy
}

func (m *model) MoveCursor(offx, offy int) {
	m.x += offx
	m.y += offy
	m.limitCursor()
}

func (m *model) limitCursor() {
	if m.x < 0 {
		m.x = 0
	}
	if m.x > m.endx-1 {
		m.x = m.endx - 1
	}
	if m.y < 0 {
		m.y = 0
	}
	if m.y > m.endy-1 {
		m.y = m.endy - 1
	}
	m.loc = fmt.Sprintf("Cursor is %d,%d", m.x, m.y)
}

func (m *model) GetCursor() (int, int, bool, bool) {
	return m.x, m.y, m.enab, !m.hide
}

func (m *model) SetCursor(x int, y int) {
	m.x = x
	m.y = y

	m.limitCursor()
}

func (m *root) updateKeys() {
	mm := m.model
	_, by := mm.GetBounds()
	if mm.y >= len(mm.block) {
		mm.y = len(mm.block) - by
	}
	m.parent.Update()
}

func setLineRune(str string) []rune {
	var lineRune []rune
	for _, runeValue := range str {
		switch runewidth.RuneWidth(runeValue) {
		case 0:
			lineRune = append(lineRune, rune(' '))
		case 1:
			lineRune = append(lineRune, runeValue)
		case 2:
			lineRune = append(lineRune, runeValue)
			lineRune = append(lineRune, rune(' '))
		}
	}
	return lineRune
}

func (m *model) GetCell(x, vy int) (rune, tcell.Style, []rune, int) {
	y := vy
	if m.y > 0 {
		y = m.y + vy
	}
	if x < 0 || y < 0 || y >= len(m.block) || x >= len(m.block[y]) {
		return 0, tcell.StyleDefault, nil, 1
	}
	if y != m.line {
		m.lineRune = setLineRune(string(m.block[y]))
		m.line = y
	}
	if x < len(m.lineRune) {
		return m.lineRune[x], tcell.StyleDefault, nil, 1
	}
	return 0, tcell.StyleDefault, nil, 1
}

func (m *model) SetCell(r io.Reader) {
	m.block = make([][]byte, 0)
	m.line = -1
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		b0 := scanner.Bytes()
		b1 := make([]byte, len(b0))
		copy(b1, b0)
		m.block = append(m.block, b1)
	}
}

func main() {
	root := &root{}
	app := &views.Application{}
	app.SetStyle(tcell.StyleDefault)

	root.parent = app
	root.main = views.NewCellView()
	m := &model{}
	root.main.SetModel(m)
	root.model = m
	root.main.SetStyle(tcell.StyleDefault)
	root.Panel.SetContent(root.main)

	status := views.NewSimpleStyledTextBar()
	status.SetStyle(tcell.StyleDefault.
		Background(tcell.ColorBlue).
		Foreground(tcell.ColorYellow))
	status.SetCenter("Status bar")
	root.status = status

	title := views.NewTextBar()
	title.SetStyle(tcell.StyleDefault.
		Background(tcell.ColorTeal).
		Foreground(tcell.ColorWhite))
	title.SetCenter("Title", tcell.StyleDefault)
	root.title = title

	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	go m.SetCell(file)

	app.SetRootWidget(root)
	if e := app.Run(); e != nil {
		fmt.Fprintln(os.Stderr, e.Error())
		os.Exit(1)
	}
}
