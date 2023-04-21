package pipbot

const Port = "COM1"
const Baud = 115200

func MakeGrid() *Layout {
	ret := &Layout{
		Matrices: make([]*Matrix, 4),
	}

	ret.Matrices[0] = NewMatrix(Unknown, "Purp", &Position{X: 29, Y: 17, Z: 80}, 42.5-29,
		42.5-29, 5, 16)

	ret.Matrices[1] = NewMatrix(Unknown, "96", &Position{X: 35.5, Y: 86.5, Z: 74.5},
		9,
		9, 8, 12)

	ret.Matrices[2] = NewMatrix(Stock, "12", &Position{X: 46, Y: 178.5, Z: 75},
		72-46,
		72-46, 3, 4)
	ret.Matrices[3] = NewMatrix(Tip, "tips", &Position{
		X: 165,
		Y: 103.5,
		Z: 73.5,
	}, 173.5-165, 173.5-165, 12, 8,
	)

	return ret
}
