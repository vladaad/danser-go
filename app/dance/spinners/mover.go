package spinners

import (
	"github.com/wieku/danser-go/framework/math/vector"
	"strings"
)

const rpms = 0.00795

var center = vector.NewVec2f(256, 192)

type SpinnerMover interface {
	Init(start, end float64)
	GetPositionAt(time float64) vector.Vector2f
}

func GetMoverByName(name string) SpinnerMover {
	switch strings.ToLower(name) {
	case "heart":
		return NewHeartMover()
	case "triangle":
		return NewTriangleMover()
	case "square":
		return NewSquareMover()
	case "cube":
		return NewCubeMover()
	default:
		return NewCircleMover()
	}
}

func GetMoverCtorByName(name string) func() SpinnerMover {
	return func() SpinnerMover {
		return GetMoverByName(name)
	}
}
