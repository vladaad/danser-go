package objects

import (
	"github.com/wieku/danser-go/app/audio"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
	"strconv"
)

const defaultCircleName = "hit"

type Circle struct {
	*HitObject

	sample  int
	Timings *Timings

	textFade *animation.Glider

	hitCircleTexture *texture.TextureRegion
	fullTexture      *texture.TextureRegion
	hitCircle        *sprite.Sprite
	hitCircleOverlay *sprite.Sprite
	approachCircle   *sprite.Sprite
	reverseArrow     *sprite.Sprite
	sprites          []*sprite.Sprite
	diff             *difficulty.Difficulty
	lastTime         float64
	silent           bool
	firstEndCircle   bool
	textureName      string
	appearTime       float64
	ArrowRotation    float64

	SliderPoint      bool
	SliderPointStart bool
	SliderPointEnd   bool
}

func NewCircle(data []string) *Circle {
	circle := &Circle{
		HitObject: commonParse(data, 5),
	}

	f, _ := strconv.ParseInt(data[4], 10, 64)
	circle.sample = int(f)

	circle.textureName = defaultCircleName

	return circle
}

func DummyCircle(pos vector.Vector2f, time float64) *Circle {
	return DummyCircleInherit(pos, time, false, false, false)
}

func DummyCircleInherit(pos vector.Vector2f, time float64, inherit bool, inheritStart bool, inheritEnd bool) *Circle {
	circle := &Circle{HitObject: &HitObject{}}
	circle.StartPosRaw = pos
	circle.EndPosRaw = pos
	circle.StartTime = time
	circle.EndTime = time
	circle.SliderPoint = inherit
	circle.SliderPointStart = inheritStart
	circle.SliderPointEnd = inheritEnd
	circle.silent = true
	circle.textureName = "sliderstart"

	return circle
}

func NewSliderEndCircle(pos vector.Vector2f, appearTime, time float64, first, last bool) *Circle {
	circle := &Circle{HitObject: &HitObject{}}
	circle.StartPosRaw = pos
	circle.EndPosRaw = pos
	circle.StartTime = time
	circle.EndTime = time
	circle.SliderPoint = true
	circle.SliderPointEnd = last
	circle.firstEndCircle = first
	circle.silent = true
	circle.textureName = "sliderend"
	circle.appearTime = appearTime

	return circle
}

func (circle *Circle) Update(time float64) bool {
	if !circle.silent && ((!settings.PLAY && !settings.KNOCKOUT) || settings.PLAYERS > 1) && (circle.lastTime < circle.StartTime && time >= circle.StartTime) {
		circle.Arm(true, circle.StartTime)
		circle.PlaySound()
	}

	for _, s := range circle.sprites {
		s.Update(time)
	}

	if circle.textFade != nil {
		circle.textFade.Update(time)
	}

	circle.lastTime = time

	return true
}

func (circle *Circle) PlaySound() {
	if circle.audioSubmissionDisabled {
		return
	}

	point := circle.Timings.GetPoint(circle.StartTime)

	index := circle.BasicHitSound.CustomIndex
	sampleSet := circle.BasicHitSound.SampleSet

	if index == 0 {
		index = point.SampleIndex
	}

	if sampleSet == 0 {
		sampleSet = point.SampleSet
	}

	audio.PlaySample(sampleSet, circle.BasicHitSound.AdditionSet, circle.sample, index, point.SampleVolume, circle.HitObjectID, circle.GetStackedStartPosition().X64())
}

func (circle *Circle) SetTiming(timings *Timings) {
	circle.Timings = timings
}

func (circle *Circle) SetDifficulty(diff *difficulty.Difficulty) {
	circle.diff = diff

	startTime := circle.StartTime - diff.Preempt

	if circle.SliderPoint {
		startTime = circle.appearTime
	}

	endTime := circle.StartTime

	circle.textFade = animation.NewGlider(0)

	defaul := skin.GetTexture(defaultCircleName + "circle")
	named := skin.GetTexture(circle.textureName + "circle")

	name := circle.textureName + "circle"

	if named == nil || skin.GetMostSpecific(named, defaul) == defaul {
		name = defaultCircleName + "circle"
	}

	circle.hitCircleTexture = skin.GetTexture(name)
	circle.fullTexture = skin.GetTexture("hitcircle-full")

	circle.hitCircle = sprite.NewSpriteSingle(circle.hitCircleTexture, 0, vector.NewVec2d(0, 0), bmath.Origin.Centre)
	circle.hitCircleOverlay = sprite.NewSpriteSingle(skin.GetTextureSource(name+"overlay", skin.GetSource(name)), 0, vector.NewVec2d(0, 0), bmath.Origin.Centre)
	circle.approachCircle = sprite.NewSpriteSingle(skin.GetTexture("approachcircle"), 0, vector.NewVec2d(0, 0), bmath.Origin.Centre)
	circle.reverseArrow = sprite.NewSpriteSingle(skin.GetTexture("reversearrow"), 0, vector.NewVec2d(0, 0), bmath.Origin.Centre)

	circle.sprites = append(circle.sprites, circle.hitCircle, circle.hitCircleOverlay, circle.approachCircle, circle.reverseArrow)

	circle.hitCircle.SetAlpha(0)
	circle.hitCircleOverlay.SetAlpha(0)
	circle.approachCircle.SetAlpha(0)
	circle.reverseArrow.SetAlpha(0)

	circles := []*sprite.Sprite{circle.hitCircle, circle.hitCircleOverlay}

	for _, t := range circles {
		if diff.CheckModActive(difficulty.Hidden) {
			if !circle.SliderPoint || circle.SliderPointStart || circle.firstEndCircle {
				t.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, startTime, startTime+diff.Preempt*0.4, 0.0, 1.0))
				t.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, startTime+diff.Preempt*0.4, startTime+diff.Preempt*0.7, 1.0, 0.0))
			}
		} else {
			t.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, startTime, startTime+difficulty.HitFadeIn, 0.0, 1.0))
			if !circle.SliderPoint || circle.SliderPointStart {
				t.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, endTime+float64(diff.Hit100), endTime+float64(diff.Hit50), 1.0, 0.0))
			} else {
				t.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, endTime, endTime, 1.0, 0.0))
			}
		}
	}

	circle.reverseArrow.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, startTime, math.Min(endTime, startTime+150), 0.0, 1.0))
	circle.reverseArrow.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, endTime, endTime, 1.0, 0.0))

	if diff.CheckModActive(difficulty.Hidden) {
		circle.textFade.AddEventS(startTime, startTime+diff.Preempt*0.4, 0.0, 1.0)
		circle.textFade.AddEventS(startTime+diff.Preempt*0.4, startTime+diff.Preempt*0.7, 1.0, 0.0)
	} else {
		circle.textFade.AddEventS(startTime, startTime+difficulty.HitFadeIn, 0.0, 1.0)
		circle.textFade.AddEventS(endTime+float64(diff.Hit100), endTime+float64(diff.Hit50), 1.0, 0.0)
	}

	if !diff.CheckModActive(difficulty.Hidden) || circle.HitObjectID == 0 {
		circle.approachCircle.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, startTime, math.Min(endTime, endTime-diff.Preempt+difficulty.HitFadeIn*2), 0.0, 0.9))
		circle.approachCircle.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, endTime, endTime, 0.0, 0.0))

		circle.approachCircle.AddTransform(animation.NewSingleTransform(animation.Scale, easing.Linear, startTime, endTime, 4.0, 1.0))
	}

	for t := startTime; t < endTime; t += 300 {
		length := math.Min(300, endTime-t)
		circle.reverseArrow.AddTransform(animation.NewSingleTransform(animation.Scale, easing.OutQuad, t, t+length, 1.3, 1.0))
	}
}

func (circle *Circle) Arm(clicked bool, time float64) {
	circle.hitCircle.ClearTransformations()
	circle.hitCircleOverlay.ClearTransformations()
	circle.textFade.Reset()

	startTime := time

	circle.approachCircle.ClearTransformations()
	circle.approachCircle.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, startTime, startTime, 0.0, 0.0))

	if clicked && !circle.diff.CheckModActive(difficulty.Hidden) {
		endTime := startTime + difficulty.HitFadeOut
		circle.hitCircle.AddTransform(animation.NewSingleTransform(animation.Scale, easing.OutQuad, startTime, endTime, 1.0, 1.4))
		circle.hitCircleOverlay.AddTransform(animation.NewSingleTransform(animation.Scale, easing.OutQuad, startTime, endTime, 1.0, 1.4))
		circle.reverseArrow.AddTransform(animation.NewSingleTransform(animation.Scale, easing.OutQuad, startTime, endTime, 1.0, 1.4))

		circle.hitCircle.AddTransform(animation.NewSingleTransform(animation.Fade, easing.OutQuad, startTime, endTime, 1.0, 0.0))
		circle.hitCircleOverlay.AddTransform(animation.NewSingleTransform(animation.Fade, easing.OutQuad, startTime, endTime, 1.0, 0.0))
		circle.reverseArrow.AddTransform(animation.NewSingleTransform(animation.Fade, easing.OutQuad, startTime, endTime, 1.0, 0.0))
		circle.textFade.AddEventS(startTime, startTime+60, 1.0, 0.0)
	} else {
		endTime := startTime + 60
		circle.hitCircle.AddTransform(animation.NewSingleTransform(animation.Fade, easing.OutQuad, startTime, endTime, circle.hitCircle.GetAlpha(), 0.0))
		circle.hitCircleOverlay.AddTransform(animation.NewSingleTransform(animation.Fade, easing.OutQuad, startTime, endTime, circle.hitCircleOverlay.GetAlpha(), 0.0))
		circle.textFade.AddEventS(startTime, endTime, circle.textFade.GetValue(), 0.0)
	}
}

func (circle *Circle) Shake(time float64) {
	for _, s := range circle.sprites {
		s.ClearTransformationsOfType(animation.MoveX)
		s.AddTransform(animation.NewSingleTransform(animation.MoveX, easing.Linear, time, time+20, 0, 8))
		s.AddTransform(animation.NewSingleTransform(animation.MoveX, easing.Linear, time+20, time+40, 8, -8))
		s.AddTransform(animation.NewSingleTransform(animation.MoveX, easing.Linear, time+40, time+60, -8, 8))
		s.AddTransform(animation.NewSingleTransform(animation.MoveX, easing.Linear, time+60, time+80, 8, -8))
		s.AddTransform(animation.NewSingleTransform(animation.MoveX, easing.Linear, time+80, time+100, -8, 8))
		s.AddTransform(animation.NewSingleTransform(animation.MoveX, easing.Linear, time+100, time+120, 8, 0))
	}
}

func (circle *Circle) UpdateStacking() {

}

func (circle *Circle) Draw(time float64, color color2.Color, batch *batch.QuadBatch) bool {
	position := circle.GetStackedPositionAtMod(time, circle.diff.Mods)

	batch.SetSubScale(1, 1)
	batch.SetTranslation(position.Copy64())

	alpha := float64(color.A)

	if settings.DIVIDES >= settings.Objects.Colors.MandalaTexturesTrigger {
		alpha *= settings.Objects.Colors.MandalaTexturesAlpha
		circle.hitCircle.Textures[0] = circle.fullTexture
	} else {
		circle.hitCircle.Textures[0] = circle.hitCircleTexture
	}

	batch.SetColor(1, 1, 1, alpha)

	//TODO: REDO THIS
	if settings.Skin.UseColorsFromSkin && len(skin.GetInfo().ComboColors) > 0 {
		color := skin.GetInfo().ComboColors[int(circle.ComboSet)%len(skin.GetInfo().ComboColors)]
		circle.hitCircle.SetColor(color2.NewRGB(color.R, color.G, color.B))
	} else if settings.Objects.Colors.UseComboColors && len(settings.Objects.Colors.ComboColors) > 0 {
		cHSV := settings.Objects.Colors.ComboColors[int(circle.ComboSet)%len(settings.Objects.Colors.ComboColors)]
		r, g, b := color2.HSVToRGB(float32(cHSV.Hue), float32(cHSV.Saturation), float32(cHSV.Value))
		circle.hitCircle.SetColor(color2.NewRGB(r, g, b))
	} else {
		circle.hitCircle.SetColor(color2.NewRGB(color.R, color.G, color.B))
	}

	circle.hitCircle.Draw(time, batch)

	if settings.DIVIDES < settings.Objects.Colors.MandalaTexturesTrigger {
		if !skin.GetInfo().HitCircleOverlayAboveNumber {
			circle.hitCircleOverlay.Draw(time, batch)
		}

		if !circle.SliderPoint || circle.SliderPointStart {
			if settings.DIVIDES < 2 && settings.Objects.DrawComboNumbers {
				fnt := skin.GetFont("default")
				batch.SetColor(1, 1, 1, alpha*circle.textFade.GetValue())
				fnt.DrawOriginV(batch, position.Copy64(), bmath.Origin.Centre, 0.8*fnt.GetSize(), false, strconv.Itoa(int(circle.ComboNumber)))
			}
		} else if !circle.SliderPointEnd {
			circle.reverseArrow.SetRotation(circle.ArrowRotation)
			circle.reverseArrow.Draw(time, batch)
		}

		batch.SetSubScale(1, 1)
		batch.SetTranslation(position.Copy64())
		batch.SetColor(1, 1, 1, alpha)
		if skin.GetInfo().HitCircleOverlayAboveNumber {
			circle.hitCircleOverlay.Draw(time, batch)
		}
	}

	batch.SetSubScale(1, 1)
	batch.SetTranslation(vector.NewVec2d(0, 0))

	if time >= circle.StartTime && circle.hitCircle.GetAlpha() <= 0.001 {
		return true
	}
	return false
}

func (circle *Circle) DrawApproach(time float64, color color2.Color, batch *batch.QuadBatch) {
	position := circle.GetStackedPositionAtMod(time, circle.diff.Mods)

	batch.SetSubScale(1, 1)
	batch.SetTranslation(position.Copy64())
	batch.SetColor(1, 1, 1, float64(color.A))

	if settings.Skin.UseColorsFromSkin && len(skin.GetInfo().ComboColors) > 0 {
		color := skin.GetInfo().ComboColors[int(circle.ComboSet)%len(skin.GetInfo().ComboColors)]
		circle.approachCircle.SetColor(color2.NewRGB(color.R, color.G, color.B))
	} else if settings.Objects.Colors.UseComboColors && len(settings.Objects.Colors.ComboColors) > 0 {
		cHSV := settings.Objects.Colors.ComboColors[int(circle.ComboSet)%len(settings.Objects.Colors.ComboColors)]
		r, g, b := color2.HSVToRGB(float32(cHSV.Hue), float32(cHSV.Saturation), float32(cHSV.Value))
		circle.approachCircle.SetColor(color2.NewRGB(r, g, b))
	} else {
		circle.approachCircle.SetColor(color2.NewRGB(color.R, color.G, color.B))
	}

	circle.approachCircle.Draw(time, batch)
}

func (circle *Circle) GetType() Type {
	return CIRCLE
}
