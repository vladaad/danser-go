package settings

var Objects = initObjects()

func initObjects() *objects {
	return &objects{
		DrawApproachCircles: true,
		DrawComboNumbers:    true,
		DrawFollowPoints:    true,
		LoadSpinners:        true,
		ScaleToTheBeat:      false,
		StackEnabled:        true,
		Sliders: &sliders{
			ForceSliderBallTexture: true,
			DrawEndCircles:         true,
			DrawSliderFollowCircle: true,
			DrawScorePoints:        true,
			SliderMerge:            false,
			SliderDistortions:      true,
			BorderWidth:            1.0,
			Quality: &quality{
				CircleLevelOfDetail: 50,
				PathLevelOfDetail:   50,
			},
			Snaking: &snaking{
				In:                 true,
				Out:                true,
				DurationMultiplier: 0,
			},
		},
		Colors: &objectcolors{
			MandalaTexturesTrigger: 5,
			MandalaTexturesAlpha:   0.3,
			Color: &color{
				EnableRainbow: true,
				RainbowSpeed:  8,
				BaseColor: &hsv{
					0,
					1.0,
					1.0},
				EnableCustomHueOffset: false,
				HueOffset:             0,
				FlashToTheBeat:        false,
				FlashAmplitude:        100,
				currentHue:            0,
			},
			UseComboColors: false,
			ComboColors:    []*hsv{{Hue: 0, Saturation: 1, Value: 1}},
			Sliders: &sliderColors{
				WhiteScorePoints:      true,
				ScorePointColorOffset: 0,
				SliderBallTint:        false,
				Border: &borderColors{
					UseHitCircleColor: false,
					Color: &color{
						EnableRainbow: false,
						RainbowSpeed:  8,
						BaseColor: &hsv{
							0,
							0.0,
							1.0},
						EnableCustomHueOffset: false,
						HueOffset:             0,
						FlashToTheBeat:        false,
						FlashAmplitude:        100,
						currentHue:            0,
					},
					EnableCustomGradientOffset: true,
					CustomGradientOffset:       0,
				},
				Body: &bodyColors{
					UseHitCircleColor: true,
					Color: &color{
						EnableRainbow: false,
						RainbowSpeed:  8,
						BaseColor: &hsv{
							0,
							1.0,
							0.0},
						EnableCustomHueOffset: false,
						HueOffset:             0,
						FlashToTheBeat:        true,
						FlashAmplitude:        100,
						currentHue:            0,
					},
					InnerOffset: -0.5,
					OuterOffset: -0.05,
					InnerAlpha:  0.8,
					OuterAlpha:  0.8,
				},
			},
		},
	}
}

type objects struct {
	DrawApproachCircles bool //true
	DrawComboNumbers    bool
	DrawFollowPoints    bool
	LoadSpinners        bool
	ScaleToTheBeat      bool //true, objects size is changing with music peak amplitude
	StackEnabled        bool //true, stack leniency
	Sliders             *sliders
	Colors              *objectcolors
}

type sliders struct {
	ForceSliderBallTexture bool
	DrawEndCircles         bool
	DrawSliderFollowCircle bool
	DrawScorePoints        bool //true
	SliderMerge            bool
	SliderDistortions      bool //true, osu!stable slider distortions on aspire maps
	BorderWidth            float64
	Quality                *quality
	Snaking                *snaking
}

type quality struct {
	// Quality of slider unit circle, 50 means that circle will have 50 sides
	CircleLevelOfDetail int64 //30, number of triangles in a circle

	//Quality of slider path, 50 means that unit circle will be placed every 2 osu!pixels (100/PathLevelOfDetail)
	PathLevelOfDetail int64 //50, int(pixelLength*(PathLOD/100)) => number of slider path points
}

type snaking struct {
	In                 bool
	Out                bool
	DurationMultiplier float64
	FadeMultiplier     float64
}

type objectcolors struct {
	MandalaTexturesTrigger int     //5, minimum value of cursors needed to use more translucent texture
	MandalaTexturesAlpha   float64 //0.3
	Color                  *color
	UseComboColors         bool
	ComboColors            []*hsv
	Sliders                *sliderColors
}

type sliderColors struct {
	WhiteScorePoints      bool    //true
	ScorePointColorOffset float64 //0.0, hue offset of the followpoint
	SliderBallTint        bool
	Border                *borderColors
	Body                  *bodyColors
}

type borderColors struct {
	UseHitCircleColor          bool
	Color                      *color
	EnableCustomGradientOffset bool
	CustomGradientOffset       float64 //18, hue offset of slider outer border
}

type bodyColors struct {
	UseHitCircleColor bool
	Color             *color
	InnerOffset       float64
	OuterOffset       float64
	InnerAlpha        float64
	OuterAlpha        float64
}
