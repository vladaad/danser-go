package osu

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

type Buttons int64

const Left = Buttons(1)
const Right = Buttons(2)

type sliderstate struct {
	downButton  Buttons
	isStartHit  bool
	isHit       bool
	points      []tickpoint
	scored      int
	missed      int
	slideStart  int64
	sliding     bool
	startScored bool
}

type tickpoint struct {
	time       int64
	scoreGiven HitResult
	edgeNum    int
}

type Slider struct {
	ruleSet           *OsuRuleSet
	hitSlider         *objects.Slider
	players           []*difficultyPlayer
	state             map[*difficultyPlayer]*sliderstate
	fadeStartRelative float64

	lastSliderTime int64
	sliderPosition vector.Vector2f

	lastSliderTimeHR int64
	sliderPositionHR vector.Vector2f

	lastSliderTimeEZ int64
	sliderPositionEZ vector.Vector2f
}

func (slider *Slider) GetNumber() int64 {
	return slider.hitSlider.GetID()
}

func (slider *Slider) IsSliding(player *difficultyPlayer) bool {
	return slider.state[player].sliding
}

func (slider *Slider) Init(ruleSet *OsuRuleSet, object objects.IHitObject, players []*difficultyPlayer) {
	slider.ruleSet = ruleSet
	slider.hitSlider = object.(*objects.Slider)
	slider.players = players
	slider.state = make(map[*difficultyPlayer]*sliderstate)

	rSlider := object.(*objects.Slider)

	slider.lastSliderTime = math.MinInt64
	slider.lastSliderTimeEZ = math.MinInt64
	slider.lastSliderTimeHR = math.MinInt64
	slider.fadeStartRelative = 100000

	for _, player := range slider.players {
		slider.fadeStartRelative = math.Min(slider.fadeStartRelative, player.diff.Preempt)
		slider.state[player] = new(sliderstate)

		edgeNumber := 1

		for _, point := range rSlider.ScorePoints {
			if point.IsReverse {
				slider.state[player].points = append(slider.state[player].points, tickpoint{int64(point.Time), SliderRepeat, edgeNumber})
				edgeNumber++
			} else {
				slider.state[player].points = append(slider.state[player].points, tickpoint{int64(point.Time), SliderPoint, -1})
			}
		}

		if len(slider.state[player].points) > 0 {
			slider.state[player].points[len(slider.state[player].points)-1].time = bmath.MaxI64(int64(slider.hitSlider.GetStartTime())+int64(slider.hitSlider.GetEndTime()-slider.hitSlider.GetStartTime())/2, int64(slider.hitSlider.GetEndTime())-36) //slider ends 36ms before the real end for scoring
			slider.state[player].points[len(slider.state[player].points)-1].scoreGiven = SliderEnd
		}
	}
}

func (slider *Slider) UpdateClickFor(player *difficultyPlayer, time int64) bool {
	state := slider.state[player]

	position := slider.hitSlider.GetStackedStartPositionMod(player.diff.Mods)

	clicked := player.leftCondE || player.rightCondE

	radius := float32(player.diff.CircleRadius)
	if player.diff.CheckModActive(difficulty.Relax2) {
		radius = 100
	}

	inRadius := player.cursor.RawPosition.Dst(position) <= radius

	if clicked && inRadius && !state.isStartHit && !state.isHit {
		if slider.ruleSet.CanBeHit(time, slider, player) == Click {
			if player.leftCondE {
				player.leftCondE = false
			} else if player.rightCondE {
				player.rightCondE = false
			}

			if player.leftCond {
				state.downButton = Left
			} else if player.rightCond {
				state.downButton = Right
			} else {
				state.downButton = player.mouseDownButton
			}

			hit := SliderMiss
			combo := ComboResults.Reset

			relative := int64(math.Abs(float64(time) - slider.hitSlider.GetStartTime()))

			if relative < player.diff.Hit50 {
				hit = SliderStart
				state.startScored = true
				combo = ComboResults.Increase
			}

			if hit != Ignore {
				if len(slider.players) == 1 {
					slider.hitSlider.HitEdge(0, float64(time), hit != SliderMiss)
				}

				slider.ruleSet.SendResult(time, player.cursor, slider.hitSlider.GetID(), position.X, position.Y, hit, true, combo)

				state.isStartHit = true
			}
		} else {
			player.leftCondE = false
			player.rightCondE = false
		}
	}

	return state.isStartHit
}

func (slider *Slider) UpdateFor(player *difficultyPlayer, time int64) bool {
	state := slider.state[player]

	var sliderPosition vector.Vector2f

	switch {
	case player.diff.Mods&difficulty.HardRock > 0:
		if time != slider.lastSliderTimeHR {
			slider.sliderPositionHR = slider.hitSlider.GetStackedPositionAtMod(float64(time), difficulty.HardRock)
			slider.lastSliderTimeHR = time
		}

		sliderPosition = slider.sliderPositionHR
	case player.diff.Mods&difficulty.Easy > 0:
		if time != slider.lastSliderTimeEZ {
			slider.sliderPositionEZ = slider.hitSlider.GetStackedPositionAtMod(float64(time), difficulty.Easy)
			slider.lastSliderTimeEZ = time
		}

		sliderPosition = slider.sliderPositionEZ
	default:
		if time != slider.lastSliderTime {
			slider.sliderPosition = slider.hitSlider.GetStackedPositionAt(float64(time))
			slider.lastSliderTime = time
		}

		sliderPosition = slider.sliderPosition
	}

	if time >= int64(slider.hitSlider.GetStartTime()) && !state.isHit {
		mouseDownAcceptable := false
		mouseDownAcceptableSwap := player.gameDownState &&
			!(player.lastButton == (Left|Right) &&
				player.lastButton2 == player.mouseDownButton)

		if player.gameDownState {
			if state.downButton == Buttons(0) || (player.mouseDownButton != (Left|Right) && mouseDownAcceptableSwap) {
				state.downButton = Buttons(0)
				if player.leftCond {
					state.downButton = Left
				} else if player.rightCond {
					state.downButton = Right
				} else {
					state.downButton = player.mouseDownButton
				}

				mouseDownAcceptable = true
			} else if (player.mouseDownButton & state.downButton) > 0 {
				mouseDownAcceptable = true
			}
		} else {
			state.downButton = Buttons(0)
		}

		mouseDownAcceptable = mouseDownAcceptable || mouseDownAcceptableSwap || player.diff.CheckModActive(difficulty.Relax)

		radiusNeeded := player.diff.CircleRadius
		if state.sliding {
			radiusNeeded *= 2.4
		}

		allowable := mouseDownAcceptable && player.cursor.RawPosition.Dst(sliderPosition) <= float32(radiusNeeded)

		if allowable && !state.sliding {
			state.sliding = true
			state.slideStart = time

			if len(slider.players) == 1 {
				slider.hitSlider.InitSlide(float64(time))
			}
		}

		pointsPassed := 0

		for _, point := range state.points {
			if point.time > time {
				break
			}

			pointsPassed++
		}

		if state.scored+state.missed < pointsPassed {
			index := state.scored + state.missed
			point := state.points[index]

			if allowable && state.slideStart <= point.time {
				state.scored++

				var scoreGiven HitResult
				if pointsPassed == len(state.points) {
					scoreGiven = SliderEnd
				} else if pointsPassed%(len(state.points)/len(slider.hitSlider.TickReverse)) == 0 {
					scoreGiven = SliderRepeat
				} else {
					scoreGiven = SliderPoint
				}

				slider.ruleSet.SendResult(time, player.cursor, slider.hitSlider.GetID(), slider.hitSlider.GetPosition().X, slider.hitSlider.GetPosition().Y, scoreGiven, true, ComboResults.Increase)
			} else {
				state.missed++

				combo := ComboResults.Reset
				if state.scored+state.missed == len(state.points) {
					combo = ComboResults.Hold
				}

				slider.ruleSet.SendResult(time, player.cursor, slider.hitSlider.GetID(), slider.hitSlider.GetPosition().X, slider.hitSlider.GetPosition().Y, SliderMiss, true, combo)
			}
		}

		if !allowable && state.sliding && state.scored+state.missed < len(state.points) {
			if len(slider.players) == 1 {
				slider.hitSlider.KillSlide(float64(time))
			}

			state.sliding = false
		}
	}

	return true
}

func (slider *Slider) UpdatePostFor(player *difficultyPlayer, time int64) bool {
	state := slider.state[player]

	if time > int64(slider.hitSlider.GetStartTime())+player.diff.Hit50 && !state.isStartHit {
		if len(slider.players) == 1 {
			slider.hitSlider.ArmStart(false, float64(time))
		}

		slider.ruleSet.SendResult(time, player.cursor, slider.hitSlider.GetID(), slider.hitSlider.GetPosition().X, slider.hitSlider.GetPosition().Y, SliderMiss, true, ComboResults.Reset)

		if player.leftCond {
			state.downButton = Left
		} else if player.rightCond {
			state.downButton = Right
		} else {
			state.downButton = player.mouseDownButton
		}

		state.isStartHit = true
	}

	if time >= int64(slider.hitSlider.GetEndTime()) && !state.isHit {
		if state.startScored {
			state.scored++
		}

		hit := Miss
		combo := ComboResults.Reset

		rate := float64(state.scored) / float64(len(state.points)+1)

		if rate > 0 && len(slider.players) == 1 {
			slider.hitSlider.HitEdge(len(slider.hitSlider.TickReverse), float64(time), true)
		}

		if rate == 1.0 {
			hit = Hit300
		} else if rate >= 0.5 {
			hit = Hit100
		} else if rate > 0 {
			hit = Hit50
		}

		if hit != Miss {
			combo = ComboResults.Hold
		}

		slider.ruleSet.SendResult(time, player.cursor, slider.hitSlider.GetID(), slider.hitSlider.GetPosition().X, slider.hitSlider.GetPosition().Y, hit, false, combo)

		state.isHit = true
	}

	return state.isHit
}

func (slider *Slider) UpdatePost(_ int64) bool {
	numFinishedTotal := 0

	for _, player := range slider.players {
		state := slider.state[player]

		if !state.isHit || !state.isStartHit {
			numFinishedTotal++
		}
	}

	return numFinishedTotal == 0
}

func (slider *Slider) IsHit(pl *difficultyPlayer) bool {
	return slider.state[pl].isHit
}

func (slider *Slider) IsStartHit(pl *difficultyPlayer) bool {
	return slider.state[pl].isStartHit
}

func (slider *Slider) GetFadeTime() int64 {
	return int64(slider.hitSlider.GetStartTime() - slider.fadeStartRelative)
}
