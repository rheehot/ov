package oviewer

import (
	"context"
	"log"
)

// moveLine moves to the specified line.
func (m *Document) moveLine(lN int) int {
	lN = min(lN, m.BufEndNum())
	m.topLN = lN
	m.topLX = 0
	return lN
}

// moveTop moves to the top.
func (m *Document) moveTop() {
	m.moveLine(0)
}

// Go to the top line.
// Called from a EventKey.
func (root *Root) moveTop() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	root.Doc.moveTop()
}

// Go to the bottom line.
// Called from a EventKey.
func (root *Root) moveBottom() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	root.Doc.topLX, root.Doc.topLN = root.bottomLineNum(root.Doc.BufEndNum())
}

// Move to the nth wrapping line of the specified line.
func (root *Root) moveLineNth(lN int, nTh int) (int, int) {
	lN = root.Doc.moveLine(lN)
	if !root.Doc.WrapMode {
		return lN, 0
	}
	listX := root.leftMostX(lN + root.Doc.firstLine())
	if len(listX) == 0 {
		return lN, 0
	}

	if nTh >= len(listX) {
		nTh = len(listX) - 1
	}

	root.Doc.topLN = lN
	root.Doc.topLX = listX[nTh]

	return lN, nTh
}

// Move up one screen.
// Called from a EventKey.
func (root *Root) movePgUp() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	root.moveNumUp(root.statusPos - root.headerLen)
	if root.Doc.topLN < 0 {
		root.Doc.moveTop()
	}
}

// Moves down one screen.
// Called from a EventKey.
func (root *Root) movePgDn() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	m := root.Doc
	y := m.bottomLN - m.firstLine()
	x := m.bottomLX
	root.limitMoveDown(x, y)
}

func (root *Root) limitMoveDown(x int, y int) {
	m := root.Doc
	if y+root.vHight >= m.BufEndNum()-m.SkipLines {
		tx, tn := root.bottomLineNum(m.BufEndNum())
		if y > tn || (y == tn && x > tx) {
			if m.topLN < tn || (m.topLN == tn && m.topLX < tx) {
				m.topLN = tn
				m.topLX = tx
			}
			return
		}
	}
	m.topLN = y
	m.topLX = x
}

// Moves up half a screen.
// Called from a EventKey.
func (root *Root) moveHfUp() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	root.moveNumUp((root.statusPos - root.headerLen) / 2)
	if root.Doc.topLN < 0 {
		root.Doc.moveTop()
	}
}

// Moves down half a screen.
// Called from a EventKey.
func (root *Root) moveHfDn() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	root.moveNumDown((root.statusPos - root.headerLen) / 2)
}

// numOfWrap returns the number of wrap from lX and lY.
func (root *Root) numOfWrap(lX int, lY int) int {
	m := root.Doc
	listX := root.leftMostX(m.topLN + m.firstLine() + lY)
	if len(listX) == 0 {
		return 0
	}
	return numOfSlice(listX, lX)
}

// numOfSlice returns what number x is in slice.
func numOfSlice(listX []int, x int) int {
	for n, v := range listX {
		if v >= x {
			return n
		}
	}
	return len(listX) - 1
}

// numOfReverseSlice returns what number x is from the back of slice.
func numOfReverseSlice(listX []int, x int) int {
	for n := len(listX) - 1; n >= 0; n-- {
		if listX[n] <= x {
			return n
		}
	}
	return 0
}

// Moves up by the specified number of y.
func (root *Root) moveNumUp(moveY int) {
	m := root.Doc
	if !m.WrapMode {
		m.topLN -= moveY
		return
	}

	// WrapMode
	num := m.topLN + m.firstLine()
	m.topLX, num = root.findNumUp(m.topLX, num, moveY)
	m.topLN = num - m.firstLine()
}

// Moves down by the specified number of y.
func (root *Root) moveNumDown(moveY int) {
	m := root.Doc
	num := m.topLN + m.firstLine()
	if !m.WrapMode {
		root.limitMoveDown(0, num+moveY)
		return
	}

	// WrapMode
	x := m.topLX
	listX := root.leftMostX(num)
	n := numOfReverseSlice(listX, x)

	for y := 0; y < moveY; y++ {
		if n >= len(listX) {
			num++
			if num > m.BufEndNum() {
				break
			}
			listX = root.leftMostX(num)
			n = 0
		}
		x = 0
		if len(listX) > 0 && n < len(listX) {
			x = listX[n]
		}
		n++
	}

	root.limitMoveDown(x, num-m.Header)
}

// Move up one line.
// Called from a EventKey.
func (root *Root) moveUp() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	m := root.Doc
	if m.topLN == 0 && m.topLX == 0 {
		return
	}

	if !m.WrapMode {
		m.topLN--
		m.topLX = 0
		return
	}

	// WrapMode.
	// Same line.
	if m.topLX > 0 {
		listX := root.leftMostX(m.topLN + m.firstLine())
		for n, x := range listX {
			if x >= m.topLX {
				m.topLX = listX[n-1]
				return
			}
		}
	}

	// Previous line.
	m.topLN--
	if m.topLN < 0 {
		m.topLN = 0
		m.topLX = 0
		return
	}
	listX := root.leftMostX(m.topLN + m.firstLine())
	if len(listX) > 0 {
		m.topLX = listX[len(listX)-1]
		return
	}
	m.topLX = 0
}

// Move down one line.
// Called from a EventKey.
func (root *Root) moveDown() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	m := root.Doc
	num := m.topLN

	if !m.WrapMode {
		num++
		root.limitMoveDown(0, num)
		return
	}

	// WrapMode
	listX := root.leftMostX(m.topLN + m.firstLine())

	for _, x := range listX {
		if x > m.topLX {
			root.limitMoveDown(x, num)
			return
		}
	}

	// Next line.
	num++
	m.topLX = 0
	root.limitMoveDown(m.topLX, num)
}

func (root *Root) nextSction() {
	// Move by page, if there is no section delimiter.
	if root.Doc.SectionDelimiter == "" {
		root.movePgDn()
		return
	}

	root.resetSelect()
	defer root.releaseEventBuffer()

	m := root.Doc
	num := m.topLN + m.firstLine() + (1 - m.SectionStartPosition)
	searcher := NewSearcher(m.SectionDelimiter, m.SectionDelimiterReg, true, true)
	ctx := context.Background()
	defer ctx.Done()
	n, err := m.SearchLine(ctx, searcher, num)
	if err != nil {
		// Last section or no section.
		root.setMessage("no next section")
		root.movePgDn()
		return
	}
	root.Doc.moveLine((n - m.firstLine()) + m.SectionStartPosition)
}

func (root *Root) prevSection() {
	// Move by page, if there is no section delimiter.
	if root.Doc.SectionDelimiter == "" {
		root.movePgUp()
		return
	}

	root.resetSelect()
	defer root.releaseEventBuffer()

	m := root.Doc
	num := m.topLN + m.firstLine() - (1 + m.SectionStartPosition)
	searcher := NewSearcher(m.SectionDelimiter, m.SectionDelimiterReg, true, true)
	ctx := context.Background()
	defer ctx.Done()
	n, err := m.BackSearchLine(ctx, searcher, num)
	if err != nil {
		m.moveTop()
		return
	}
	n = (n - m.firstLine()) + m.SectionStartPosition
	n = max(n, 0)
	m.moveLine(n)
}

func (root *Root) lastSection() {
	root.resetSelect()

	m := root.Doc
	// +1 to avoid if the bottom line is a session delimiter.
	num := m.BufEndNum() - 2
	searcher := NewSearcher(m.SectionDelimiter, m.SectionDelimiterReg, true, true)
	ctx := context.Background()
	defer ctx.Done()
	n, err := m.BackSearchLine(ctx, searcher, num)
	if err != nil {
		log.Printf("last section:%v", err)
		return
	}
	n = (n - m.firstLine()) + m.SectionStartPosition
	m.moveLine(n)
}

// Move to the left.
// Called from a EventKey.
func (root *Root) moveLeft() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	m := root.Doc
	if m.ColumnMode {
		if m.columnNum > 0 {
			m.columnNum--
			m.x = root.columnModeX()
		}
		return
	}
	if m.WrapMode {
		return
	}
	m.x--
	if m.x < root.minStartX {
		m.x = root.minStartX
	}
}

// Move to the right.
// Called from a EventKey.
func (root *Root) moveRight() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	m := root.Doc
	if m.ColumnMode {
		m.columnNum++
		m.x = root.columnModeX()
		return
	}
	if m.WrapMode {
		return
	}
	m.x++
}

// columnModeX returns the actual x from m.columnNum.
func (root *Root) columnModeX() int {
	m := root.Doc
	maxNum := 0
	// m.firstLine()+10 = Maximum columnMode target.
	for i := 0; i < m.firstLine()+10; i++ {
		lc, err := m.contentsLN(m.topLN+m.firstLine()+i, m.TabWidth)
		if err != nil {
			continue
		}
		lineStr, posCV := ContentsToStr(lc)
		idxs := allIndex(lineStr, m.ColumnDelimiter)
		maxNum = max(maxNum, len(idxs))
		if len(idxs) < m.columnNum {
			continue
		}

		sx, ex := 0, 0
		if len(idxs) == m.columnNum {
			pos := idxs[m.columnNum-1]
			sx = posCV[pos[1]+len(m.ColumnDelimiter)]
			ex = posCV[len(lineStr)]
		} else {
			pos := idxs[m.columnNum]
			sx = posCV[pos[0]]
			ex = posCV[pos[1]]
		}
		if root.vWidth > ex {
			return 0
		}
		if ex-root.vWidth > 0 {
			return ex - root.vWidth
		}
		return sx
	}
	if maxNum > 0 {
		m.columnNum--
		return m.x
	}
	m.columnNum = 0
	return 0
}

// Move to the left by half a screen.
// Called from a EventKey.
func (root *Root) moveHfLeft() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	m := root.Doc

	if m.WrapMode {
		return
	}
	moveSize := (root.vWidth / 2)
	if m.x > 0 && (m.x-moveSize) < 0 {
		m.x = 0
		return
	}

	m.x -= moveSize
	if m.x < root.minStartX {
		m.x = root.minStartX
	}
}

// Move to the right by half a screen.
// Called from a EventKey.
func (root *Root) moveHfRight() {
	root.resetSelect()
	defer root.releaseEventBuffer()

	m := root.Doc

	if m.WrapMode {
		return
	}
	if m.x < 0 {
		m.x = 0
		return
	}

	m.x += (root.vWidth / 2)
}

// bottomLineNum returns the display start line
// when the last line number as an argument.
func (root *Root) bottomLineNum(lN int) (int, int) {
	if lN < root.headerLen {
		return 0, 0
	}

	hight := (root.vHight - root.headerLen) - 2
	if !root.Doc.WrapMode {
		return 0, lN - (hight + root.Doc.firstLine())
	}
	// WrapMode
	lX, lN := root.findNumUp(0, lN, hight)
	return lX, lN - root.Doc.firstLine()
}

// findNumUp finds lX, lN when the number of lines is moved up from lX, lN.
func (root *Root) findNumUp(lX int, lN int, upY int) (int, int) {
	listX := root.leftMostX(lN)
	n := numOfSlice(listX, lX)

	for y := upY; y > 0; y-- {
		if n <= 0 {
			lN--
			listX = root.leftMostX(lN)
			n = len(listX)
		}
		if n > 0 {
			lX = listX[n-1]
		} else {
			lX = 0
		}
		n--
	}
	return lX, lN
}

// leftMostX returns a list of left - most x positions when wrapping.
// Returns nil if there is no line number.
func (root *Root) leftMostX(lN int) []int {
	lc, err := root.Doc.contentsLN(lN, root.Doc.TabWidth)
	if err != nil {
		return nil
	}

	listX := make([]int, 0, (len(lc)/root.vWidth)+1)
	width := (root.vWidth - root.startX)

	listX = append(listX, 0)
	for n := width; n < len(lc); n += width {
		if lc[n-1].width == 2 {
			n--
		}
		listX = append(listX, n)
	}
	return listX
}
