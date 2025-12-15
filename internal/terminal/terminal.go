package terminal

// Terminal represents a terminal emulator
// This package will handle terminal rendering and ANSI escape sequences
type Terminal struct {
	width  int
	height int
	buffer [][]rune
}

// NewTerminal creates a new terminal
func NewTerminal(width, height int) *Terminal {
	buffer := make([][]rune, height)
	for i := range buffer {
		buffer[i] = make([]rune, width)
		for j := range buffer[i] {
			buffer[i][j] = ' '
		}
	}

	return &Terminal{
		width:  width,
		height: height,
		buffer: buffer,
	}
}

// Resize resizes the terminal
func (t *Terminal) Resize(width, height int) {
	newBuffer := make([][]rune, height)
	for i := range newBuffer {
		newBuffer[i] = make([]rune, width)
		for j := range newBuffer[i] {
			if i < len(t.buffer) && j < len(t.buffer[i]) {
				newBuffer[i][j] = t.buffer[i][j]
			} else {
				newBuffer[i][j] = ' '
			}
		}
	}

	t.width = width
	t.height = height
	t.buffer = newBuffer
}

// GetSize returns terminal dimensions
func (t *Terminal) GetSize() (width, height int) {
	return t.width, t.height
}

// TODO: Implement ANSI escape sequence parsing and rendering
// TODO: Implement cursor management
// TODO: Implement scrollback buffer
