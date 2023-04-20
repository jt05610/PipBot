package pipbot

const Port = "/dev/cu.usbserial-1120"
const Baud = 115200

func MakeGrid() *Layout {
	ret := &Layout{
		Matrices: make([]*Matrix, 4),
	}

	ret.Matrices[0] = NewMatrix("Purp", &Position{X: 29, Y: 17, Z: 80}, 42.5-29,
		42.5-29, 5, 16)

	ret.Matrices[1] = NewMatrix("96", &Position{X: 29, Y: 17, Z: 80}, 42.5-29,
		42.5-29, 5, 16)

	ret.Matrices[2] = NewMatrix("12", &Position{X: 29, Y: 17, Z: 80}, 42.5-29,
		42.5-29, 5, 16)

	ret.Matrices[3] = NewMatrix("tips", &Position{X: 29, Y: 17, Z: 80}, 42.5-29,
		42.5-29, 5, 16)

	return ret
}
