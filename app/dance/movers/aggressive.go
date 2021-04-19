package movers

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

type AggressiveMover struct {
	lastAngle          float32
	bz                 *curves.Bezier
	startTime, endTime float64
	mods               difficulty.Modifier
}

func NewAggressiveMover() MultiPointMover {
	return &AggressiveMover{lastAngle: 0}
}

func (bm *AggressiveMover) Reset(mods difficulty.Modifier) {
	bm.mods = mods
	bm.lastAngle = 0
}

func (bm *AggressiveMover) SetObjects(objs []objects.IHitObject) int {
	end := objs[0]
	start := objs[1]

	endPos := end.GetStackedEndPositionMod(bm.mods)
	endTime := end.GetEndTime()
	startPos := start.GetStackedStartPositionMod(bm.mods)
	startTime := start.GetStartTime()

	scaledDistance := float32(startTime - endTime)

	newAngle := bm.lastAngle + math.Pi
	if s, ok := end.(objects.ILongObject); ok {
		newAngle = s.GetEndAngleMod(bm.mods)
	}

	points := []vector.Vector2f{endPos, vector.NewVec2fRad(newAngle, scaledDistance).Add(endPos)}

	if scaledDistance > 1 {
		bm.lastAngle = points[1].AngleRV(startPos)
	}

	if s, ok := start.(objects.ILongObject); ok {
		points = append(points, vector.NewVec2fRad(s.GetStartAngleMod(bm.mods), scaledDistance).Add(startPos))
	}

	points = append(points, startPos)

	bm.bz = curves.NewBezierNA(points)
	bm.endTime = endTime
	bm.startTime = startTime

	return 2
}

func (bm *AggressiveMover) Update(time float64) vector.Vector2f {
	t := bmath.ClampF32(float32(time-bm.endTime)/float32(bm.startTime-bm.endTime), 0, 1)
	return bm.bz.PointAt(t)
}

func (bm *AggressiveMover) GetEndTime() float64 {
	return bm.startTime
}
