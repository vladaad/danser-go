package play

import (
	"fmt"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/font"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

const errorBase = 4.8

var colors = []color2.Color{{0.2, 0.8, 1, 1}, {0.44, 0.98, 0.18, 1}, {0.85, 0.68, 0.27, 1}}

type HitErrorMeter struct {
	diff             *difficulty.Difficulty
	errorDisplay     *sprite.SpriteManager
	errorCurrent     float64
	triangle         *sprite.Sprite
	errorDisplayFade *animation.Glider

	Width    float64
	Height   float64
	lastTime float64

	errors       []float64
	unstableRate float64
	avgPos       float64
	avgNeg       float64
}

func NewHitErrorMeter(width, height float64, diff *difficulty.Difficulty) *HitErrorMeter {
	meter := new(HitErrorMeter)
	meter.Width = width
	meter.Height = height

	meter.diff = diff
	meter.errorDisplay = sprite.NewSpriteManager()
	meter.errorDisplayFade = animation.NewGlider(0.0)

	sum := float64(meter.diff.Hit50) * 0.8

	scale := settings.Gameplay.HitErrorMeter.Scale

	pixel := graphics.Pixel.GetRegion()
	bg := sprite.NewSpriteSingle(&pixel, 0.0, vector.NewVec2d(meter.Width/2, meter.Height-errorBase*2*scale), bmath.Origin.Centre)
	bg.SetScaleV(vector.NewVec2d(sum*2, errorBase*4).Scl(scale))
	bg.SetColor(color2.NewL(0))
	bg.SetAlpha(0.6)
	meter.errorDisplay.Add(bg)

	vals := []float64{float64(meter.diff.Hit300) * 0.8, float64(meter.diff.Hit100) * 0.8, float64(meter.diff.Hit50) * 0.8}

	for i, v := range vals {
		pos := 0.0
		width := v

		if i > 0 {
			pos = vals[i-1]
			width -= vals[i-1]
		}

		left := sprite.NewSpriteSingle(&pixel, 1.0, vector.NewVec2d(meter.Width/2-pos*scale, meter.Height-errorBase*2*scale), bmath.Origin.CentreRight)
		left.SetScaleV(vector.NewVec2d(width, errorBase).Scl(scale))
		left.SetColor(colors[i])

		meter.errorDisplay.Add(left)

		right := sprite.NewSpriteSingle(&pixel, 1.0, vector.NewVec2d(meter.Width/2+pos*scale, meter.Height-errorBase*2*scale), bmath.Origin.CentreLeft)
		right.SetScaleV(vector.NewVec2d(width, errorBase).Scl(scale))
		right.SetColor(colors[i])

		meter.errorDisplay.Add(right)
	}

	middle := sprite.NewSpriteSingle(&pixel, 2.0, vector.NewVec2d(meter.Width/2, meter.Height-errorBase*2*scale), bmath.Origin.Centre)
	middle.SetScaleV(vector.NewVec2d(2, errorBase*4).Scl(scale))

	meter.errorDisplay.Add(middle)

	meter.triangle = sprite.NewSpriteSingle(graphics.TriangleSmall, 2.0, vector.NewVec2d(meter.Width/2, meter.Height-errorBase*2.5*scale), bmath.Origin.BottomCentre)
	meter.triangle.SetScaleV(vector.NewVec2d(scale/6, scale/6))
	meter.triangle.SetAlpha(1)

	meter.errorDisplay.Add(meter.triangle)

	return meter
}

func (meter *HitErrorMeter) Add(time, error float64) {
	errorA := int64(math.Abs(error))

	scale := settings.Gameplay.HitErrorMeter.Scale

	pixel := graphics.Pixel.GetRegion()

	middle := sprite.NewSpriteSingle(&pixel, 3.0, vector.NewVec2d(meter.Width/2+error*0.8*scale, meter.Height-errorBase*2*scale), bmath.Origin.Centre)
	middle.SetScaleV(vector.NewVec2d(3, errorBase*4).Scl(scale))
	middle.SetAdditive(true)

	var col color2.Color
	switch {
	case errorA < meter.diff.Hit300:
		col = colors[0]
	case errorA < meter.diff.Hit100:
		col = colors[1]
	case errorA < meter.diff.Hit50:
		col = colors[2]
	}

	middle.SetColor(col)

	middle.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, time, time+10000, 0.4, 0.0))
	middle.AdjustTimesToTransformations()

	meter.errorDisplay.Add(middle)

	meter.errorCurrent = meter.errorCurrent*0.8 + error*0.8*0.2

	meter.triangle.ClearTransformations()
	meter.triangle.AddTransform(animation.NewSingleTransform(animation.MoveX, easing.OutQuad, time, time+800, meter.triangle.GetPosition().X, meter.Width/2+meter.errorCurrent*scale))

	meter.errorDisplayFade.Reset()
	meter.errorDisplayFade.SetValue(1.0)
	meter.errorDisplayFade.AddEventSEase(time+4000, time+5000, 1.0, 0.0, easing.InQuad)

	meter.errors = append(meter.errors, error)

	averageN := 0.0
	countN := 0

	averageP := 0.0
	countP := 0
	for _, e := range meter.errors {
		if e >= 0 {
			averageP += e
			countP++
		} else {
			averageN += e
			countN++
		}
	}

	average := (averageN+averageP) / float64(countN+countP)

	urBase := 0.0
	for _, e := range meter.errors {
		urBase += math.Pow(e-average, 2)
	}

	urBase /= float64(len(meter.errors))

	meter.avgNeg = averageN / math.Max(float64(countN), 1)
	meter.avgPos = averageP / math.Max(float64(countP), 1)
	meter.unstableRate = math.Sqrt(urBase) * 10
}

func (meter *HitErrorMeter) Update(time float64) {
	meter.errorDisplayFade.Update(time)
	meter.errorDisplay.Update(time)

	meter.lastTime = time
}

func (meter *HitErrorMeter) Draw(batch *batch.QuadBatch, alpha float64) {
	batch.ResetTransform()

	meterAlpha := settings.Gameplay.HitErrorMeter.Opacity * meter.errorDisplayFade.GetValue() * alpha
	if meterAlpha > 0.001 && settings.Gameplay.HitErrorMeter.Show {
		batch.SetColor(1, 1, 1, meterAlpha)
		meter.errorDisplay.Draw(meter.lastTime, batch)

		if settings.Gameplay.HitErrorMeter.ShowUnstableRate {
			batch.SetScale(1, -1)

			pY := meter.Height - errorBase*2*settings.Gameplay.HitErrorMeter.Scale

			scale := settings.Gameplay.HitErrorMeter.UnstableRateScale

			fnt := font.GetFont("Exo 2 Bold")

			urText := fmt.Sprintf("%.0fUR", meter.GetUnstableRateConverted())
			fnt.DrawMonospaced(batch, meter.Width/2-fnt.GetWidthMonospaced(15*scale, urText)/2, pY-13.33, 15*scale, urText)
		}
	}

	batch.ResetTransform()
}

func (meter *HitErrorMeter) GetAvgNeg() float64 {
	return meter.avgNeg
}

func (meter *HitErrorMeter) GetAvgNegConverted() float64 {
	return meter.avgNeg / meter.diff.Speed
}

func (meter *HitErrorMeter) GetAvgPos() float64 {
	return meter.avgPos
}

func (meter *HitErrorMeter) GetAvgPosConverted() float64 {
	return meter.avgPos / meter.diff.Speed
}

func (meter *HitErrorMeter) GetUnstableRate() float64 {
	return meter.unstableRate
}

func (meter *HitErrorMeter) GetUnstableRateConverted() float64 {
	return meter.unstableRate / meter.diff.Speed
}