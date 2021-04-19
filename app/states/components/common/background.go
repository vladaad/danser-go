package common

import (
	"github.com/EdlinOrg/prominentcolor"
	"github.com/faiface/mainthread"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/graphics/gui/drawables"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/storyboard"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/effects"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/graphics/viewport"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/scaling"
	"github.com/wieku/danser-go/framework/math/vector"
	"log"
	"math"
	"path/filepath"
)

type Background struct {
	blur       *effects.BlurEffect
	scale      vector.Vector2d
	position   vector.Vector2d
	background *texture.TextureSingle
	storyboard *storyboard.Storyboard
	lastTime   float64
	//bMap           *beatmap.BeatMap
	triangles      *drawables.Triangles
	blurVal        float64
	blurredTexture texture.Texture
	scaling        scaling.Scaling
	forceRedraw    bool
}

func NewBackground() *Background {
	bg := new(Background)
	bg.blurVal = -1
	bg.blur = effects.NewBlurEffect(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))

	image, err := assets.GetPixmap("assets/textures/background-1.png")
	if err != nil {
		panic(err)
	}

	bg.background = texture.LoadTextureSingle(image.RGBA(), 0)

	bg.triangles = drawables.NewTriangles(bg.getColors(image))

	if image != nil {
		image.Dispose()
	}

	return bg
}

func (bg *Background) SetBeatmap(beatMap *beatmap.BeatMap, loadStoryboards bool) {
	go func() {
		image, err := texture.NewPixmapFileString(filepath.Join(settings.General.OsuSongsDir, beatMap.Dir, beatMap.Bg))
		if err != nil {
			image, err = assets.GetPixmap("assets/textures/background-1.png")
			if err != nil {
				panic(err)
			}
		}

		bg.triangles.SetColors(bg.getColors(image))

		if image != nil {
			mainthread.CallNonBlock(func() {
				bg.background = texture.LoadTextureSingle(image.RGBA(), 0)
				image.Dispose()

				bg.forceRedraw = true
			})
		}
	}()

	if loadStoryboards {
		bg.storyboard = storyboard.NewStoryboard(beatMap)

		if bg.storyboard == nil {
			log.Println("Storyboard not found!")
		}
	}

}

func (bg *Background) SetTrack(track *bass.Track) {
	bg.triangles.SetTrack(track)
}

func (bg *Background) Update(time float64, x, y float64) {
	if bg.lastTime == 0 {
		bg.lastTime = time
	}

	if bg.storyboard != nil {
		if settings.RECORD {
			bg.storyboard.Update(time)
		} else {
			if !bg.storyboard.IsThreadRunning() {
				bg.storyboard.StartThread()
			}
			bg.storyboard.UpdateTime(time)
		}
	}

	bg.triangles.Update(time)

	pX := 0.0
	pY := 0.0

	if math.Abs(settings.Playfield.Background.Parallax.Amount) > 0.0001 && !math.IsNaN(x) && !math.IsNaN(y) && settings.DIVIDES == 1 {
		pX = bmath.ClampF64(x, -1, 1) * settings.Playfield.Background.Parallax.Amount
		pY = bmath.ClampF64(y, -1, 1) * settings.Playfield.Background.Parallax.Amount
	}

	delta := math.Abs(time - bg.lastTime)

	p := math.Pow(1-settings.Playfield.Background.Parallax.Speed, delta/100)

	bg.position.X = pX*(1-p) + p*bg.position.X
	bg.position.Y = pY*(1-p) + p*bg.position.Y

	bg.lastTime = time
}

func project(pos vector.Vector2d, camera mgl32.Mat4) vector.Vector2d {
	res := camera.Mul4x1(mgl32.Vec4{pos.X32(), pos.Y32(), 0.0, 1.0})
	return vector.NewVec2d((float64(res[0])/2+0.5)*settings.Graphics.GetWidthF(), float64((res[1])/2+0.5)*settings.Graphics.GetWidthF())
}

func (bg *Background) Draw(time float64, batch *batch.QuadBatch, blurVal, bgAlpha float64, camera mgl32.Mat4) {
	if bgAlpha < 0.01 {
		return
	}

	batch.Begin()

	needsRedraw := bg.forceRedraw || bg.storyboard != nil || !settings.Playfield.Background.Blur.Enabled || (settings.Playfield.Background.Triangles.Enabled && !settings.Playfield.Background.Triangles.DrawOverBlur)

	bg.forceRedraw = false

	if math.Abs(bg.blurVal-blurVal) > 0.001 {
		needsRedraw = true
		bg.blurVal = blurVal
	}

	var clipX, clipY, clipW, clipH int
	widescreen := true

	bg.scaling = scaling.Fill

	if bg.storyboard != nil && !bg.storyboard.IsWideScreen() {
		widescreen = false

		v1 := project(vector.NewVec2d(256-320, 192+240), camera)
		v2 := project(vector.NewVec2d(256+320, 192-240), camera)

		clipX, clipY, clipW, clipH = int(v1.X32()), int(v1.Y32()), int(v2.X32()-v1.X32()), int(v2.Y32()-v1.Y32())

		bg.scaling = scaling.Fit
	}

	if needsRedraw {
		batch.ResetTransform()
		batch.SetAdditive(false)

		if settings.Playfield.Background.Blur.Enabled {
			batch.SetColor(1, 1, 1, 1)

			bg.blur.SetBlur(blurVal, blurVal)
			bg.blur.Begin()
		} else {
			batch.SetColor(bgAlpha, bgAlpha, bgAlpha, 1)
		}

		if !widescreen && !settings.Playfield.Background.Blur.Enabled {
			viewport.PushScissorPos(clipX, clipY, clipW, clipH)
		}

		if bg.background != nil && (bg.storyboard == nil || !bg.storyboard.BGFileUsed()) {
			batch.SetCamera(mgl32.Ortho(float32(-settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetHeightF()/2), float32(-settings.Graphics.GetHeightF()/2), 1, -1))
			size := bg.scaling.Apply(float32(bg.background.GetWidth()), float32(bg.background.GetHeight()), float32(settings.Graphics.GetWidthF()), float32(settings.Graphics.GetHeightF())).Scl(0.5)

			if !settings.Playfield.Background.Blur.Enabled {
				batch.SetTranslation(bg.position.Mult(vector.NewVec2d(1, -1)).Mult(vector.NewVec2d(settings.Graphics.GetSizeF()).Scl(0.5)))
				size = size.Scl(float32(1 + math.Abs(settings.Playfield.Background.Parallax.Amount)))
			}

			batch.SetScale(size.X64(), size.Y64())
			batch.DrawUnit(bg.background.GetRegion())
		}

		if bg.storyboard != nil {
			batch.SetScale(1, 1)
			batch.SetTranslation(vector.NewVec2d(0, 0))

			cam := camera
			if !settings.Playfield.Background.Blur.Enabled {
				scale := float32(1 + math.Abs(settings.Playfield.Background.Parallax.Amount))
				cam = mgl32.Translate3D(bg.position.X32(), bg.position.Y32(), 0).Mul4(mgl32.Scale3D(scale, scale, 1)).Mul4(cam)
			}

			batch.SetCamera(cam)

			bg.storyboard.Draw(time, batch)
		}

		if settings.Playfield.Background.Triangles.Enabled && !settings.Playfield.Background.Triangles.DrawOverBlur {
			bg.drawTriangles(batch, bgAlpha, settings.Playfield.Background.Blur.Enabled)
		}

		batch.Flush()
		batch.SetColor(1, 1, 1, 1)
		batch.ResetTransform()

		if !widescreen && !settings.Playfield.Background.Blur.Enabled {
			viewport.PopScissor()
		}

		if settings.Playfield.Background.Blur.Enabled {
			bg.blurredTexture = bg.blur.EndAndProcess()
		}
	}

	if !widescreen {
		viewport.PushScissorPos(clipX, clipY, clipW, clipH)
	}

	if settings.Playfield.Background.Blur.Enabled && bg.blurredTexture != nil {
		batch.ResetTransform()
		batch.SetAdditive(false)
		batch.SetColor(1, 1, 1, bgAlpha)
		batch.SetCamera(mgl32.Ortho(-1, 1, -1, 1, 1, -1))
		batch.SetTranslation(bg.position)
		batch.SetScale(1+math.Abs(settings.Playfield.Background.Parallax.Amount), 1+math.Abs(settings.Playfield.Background.Parallax.Amount))
		batch.DrawUnit(bg.blurredTexture.GetRegion())
		batch.Flush()
		batch.SetColor(1, 1, 1, 1)
		batch.ResetTransform()
	}

	if settings.Playfield.Background.Triangles.Enabled && settings.Playfield.Background.Triangles.DrawOverBlur {
		bg.drawTriangles(batch, bgAlpha, false)
	}

	batch.End()

	if !widescreen {
		viewport.PopScissor()
	}
}

func (bg *Background) drawTriangles(batch *batch.QuadBatch, bgAlpha float64, blur bool) {
	batch.ResetTransform()
	cam := mgl32.Ortho(float32(-settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetHeightF()/2), float32(-settings.Graphics.GetHeightF()/2), 1, -1)

	if !blur {
		batch.SetColor(bgAlpha, bgAlpha, bgAlpha, 1)

		subScale := float32(settings.Playfield.Background.Triangles.ParallaxMultiplier)
		scale := 1 + math32.Abs(float32(settings.Playfield.Background.Parallax.Amount))*math32.Abs(subScale)
		cam = mgl32.Translate3D(bg.position.X32()*subScale, bg.position.Y32()*subScale, 0).Mul4(mgl32.Scale3D(scale, scale, 1)).Mul4(cam)
	} else {
		batch.SetColor(1, 1, 1, 1)
	}

	batch.SetCamera(cam)
	bg.triangles.Draw(bg.lastTime, batch)

	batch.SetAdditive(false)
	batch.ResetTransform()
}

func (bg *Background) DrawOverlay(time float64, batch *batch.QuadBatch, bgAlpha float64, camera mgl32.Mat4) {
	if bgAlpha < 0.01 || bg.storyboard == nil {
		return
	}

	if !bg.storyboard.IsWideScreen() {
		v1 := project(vector.NewVec2d(256-320, 192+240), camera)
		v2 := project(vector.NewVec2d(256+320, 192-240), camera)

		viewport.PushScissorPos(int(v1.X32()), int(v1.Y32()), int(v2.X32()-v1.X32()), int(v2.Y32()-v1.Y32()))
	}

	batch.Begin()

	batch.SetColor(bgAlpha, bgAlpha, bgAlpha, 1)
	batch.ResetTransform()
	batch.SetAdditive(false)

	scale := float32(1 + math.Abs(settings.Playfield.Background.Parallax.Amount))
	cam := mgl32.Translate3D(bg.position.X32(), -bg.position.Y32(), 0).Mul4(mgl32.Scale3D(scale, scale, 1)).Mul4(camera)
	batch.SetCamera(cam)

	bg.storyboard.DrawOverlay(time, batch)

	batch.End()

	if !bg.storyboard.IsWideScreen() {
		viewport.PopScissor()
	}

	batch.SetColor(1, 1, 1, 1)
	batch.ResetTransform()
}

func (bg *Background) GetStoryboard() *storyboard.Storyboard {
	return bg.storyboard
}

func (bg *Background) getColors(image *texture.Pixmap) []color2.Color {
	newCol := make([]color2.Color, 0)

	var err error = nil

	if image != nil {
		cItems, err1 := prominentcolor.KmeansWithAll(10, image.NRGBA(), prominentcolor.ArgumentDefault, prominentcolor.DefaultSize, prominentcolor.GetDefaultMasks())
		newCol = make([]color2.Color, 0)

		err = err1

		if err1 == nil {
			for i := 0; i < len(cItems); i++ {
				if cItems[i].Color.R+cItems[i].Color.G+cItems[i].Color.B == 0 {
					continue
				}

				newCol = append(newCol, color2.NewIRGB(uint8(cItems[i].Color.R), uint8(cItems[i].Color.G), uint8(cItems[i].Color.B)) /*.Lighten2(0.15)*/)
			}
		}
	}

	if err != nil {
		color1 := color2.NewL(0.054)
		color2 := color2.NewL(0.3)
		for i := 0; i <= 10; i++ {
			newCol = append(newCol, color1.Mix(color2, float32(i)/10))
		}
	}

	return newCol
}
