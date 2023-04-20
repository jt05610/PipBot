package pipbot

type CellType uint8

const (
	Tip CellType = iota
	Stock
	Standard
	Unknown
)

// Mixture describes what is in a cell.
// Mixtures can be stocks, mixtures of mixtures, dilutions, etc.
type Mixture struct {
	Contents map[Cell]float32
}

// Cell is the fundamental discrete addressable unit in the system.
// A cell can be a pipette tip position, an individual well of a plate, etc.
type Cell struct {
	Kind  CellType
	Empty bool
	*Position
	Content *Mixture
}

// Matrix is an aggregate of Cells. This can be a well plate, pipette tip box,
// tube rack, etc.
type Matrix struct {
	Name    string
	Cells   [][]*Position
	Home    *Position
	Rows    int
	Columns int
}

// Layout describes how individual Matrix units are arranged on the build plate.
type Layout struct {
	Matrices []*Matrix
}

func NewMatrix(name string, home *Position, rowSpace, colSpace float32, nRow,
	nCol int) *Matrix {
	m := &Matrix{
		Name:    name,
		Cells:   make([][]*Position, nRow),
		Home:    home,
		Rows:    nRow,
		Columns: nCol,
	}
	for row := 0; row < nRow; row++ {
		m.Cells[row] = make([]*Position, nCol)
		for col := 0; col < nCol; col++ {
			m.Cells[row][col] = &Position{
				X: home.X + (float32(col) * colSpace),
				Y: home.Y + (float32(row) * rowSpace),
				Z: home.Z,
			}
		}
	}
	return m
}
