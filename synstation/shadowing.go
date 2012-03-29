package synstation

import "math"
import "geom"
import rand "math/rand"
import "image"
import "image/png"
import "image/color"
import "os"
import "fmt"

type shadowMapInt interface {
	Init(corr float64, Rgen *rand.Rand)
	evalShadowFading(d geom.Pos) (val float64)
}

type noshadow struct{}

func (s *noshadow) Init(f float64, rgen *rand.Rand) {
}
func (s *noshadow) evalShadowFading(d geom.Pos) (val float64) {
	return 1.0
}

type shadowMap struct {
	xcos []float64
	ycos []float64

	xsin []float64
	ysin []float64

	power float64

	smap [][]float32
}

var mapres = mapsize / float64(maplength)

func (s *shadowMap) Init(corr_dist float64, Rgen2 *rand.Rand) {

	nval := int(Field / corr_dist / shadow_sampling)

	s.xcos = make([]float64, nval)
	s.ycos = make([]float64, nval)
	s.xsin = make([]float64, nval)
	s.ysin = make([]float64, nval)

	for i := 0; i < nval; i++ {
		s.xcos[i] = Rgen2.NormFloat64()
		//s.xcos[i] *= s.xcos[i]
		s.ycos[i] = Rgen2.NormFloat64()

		s.xsin[i] = Rgen2.Float64() * 2 * math.Pi
		s.ysin[i] = Rgen2.Float64() * 2 * math.Pi

		//if !(s.xcos[i] > mval || s.xcos[i]< -mval) {
		if s.xcos[i] < mval {
			s.xcos[i] = 0
		}
		//if !(s.ycos[i] > mval || s.ycos[i]< -mval) {
		if s.ycos[i] < mval {
			s.ycos[i] = 0
		}
		s.power += s.xcos[i] * s.xcos[i]
		s.power += s.ycos[i] * s.ycos[i]

	}

	s.power = math.Sqrt(s.power) / shadow_deviance

	for i := 0; i < nval; i++ {
		s.xcos[i] /= s.power
		s.ycos[i] /= s.power
	}

	var ValMax float32
	s.smap = make([][]float32, mapsize)
	for i := 0; i < mapsize; i++ {
		s.smap[i] = make([]float32, mapsize)
		x := (float64(i) - mapsize/2) / mapres
		for j := 0; j < mapsize; j++ {
			d := geom.Pos{x, (float64(j) - mapsize/2) / mapres}
			s.smap[i][j] = float32(s.evalShadowFadingDirect(d))
			//lets not have -Inf here			
			if s.smap[i][j] < 0.0000001 {
				s.smap[i][j] = 0.0000001
			}
			if s.smap[i][j] > ValMax {ValMax=s.smap[i][j]}
		}
	}

	m := image.NewRGBA(image.Rect(0, 0, mapsize, mapsize))
	var r,g,b float64
	for i := 0; i < mapsize; i++ {
		for j := 0; j < mapsize; j++ {
			v:=float64(s.smap[i][j]/ValMax)
			switch {
			case v > .5:
				b = 2*(1-b) 
				g = 2*v-1
			default:
				r = 1-2*v
				b = 2*v
			}
			r = r * 255
			g = g * 255
			b = b * 255
			if r < 0 {
				r = 0
			}
			if r > 255 {
				r = 255
			}
			if g < 0 {
				g = 0
			}
			if g > 255 {
				g = 255
			}
			if b < 0 {
				b = 0
			}
			if b > 255 {
				b = 255
			}
			m.Set(i, j, color.NRGBA{uint8(r), uint8(g), uint8(b), uint8(255)})
	}
	}

	shadowint++
	f, err := os.OpenFile(fmt.Sprint("shadow",shadowint,".png" ), os.O_CREATE|os.O_WRONLY, 0666)

	if err = png.Encode(f, m); err != nil {
		os.Exit(-1)
	}
}
 var shadowint int

func (s *shadowMap) interpolFading(d geom.Pos) (val float64) {

	x := int(d.X*mapres + mapsize/2)
	y := int(d.Y*mapres + mapsize/2)

	if x < 0 || x >= mapsize || y < 0 || y >= mapsize {
		return 1.0
	}

	return float64(s.smap[x][y])

}

var facr = float64(2.0 * math.Pi / Field)

func (s *shadowMap) evalShadowFading(d geom.Pos) (val float64) {
	return s.interpolFading(d)
	//return s.evalShadowFadingDirect(d)
}

func (s *shadowMap) evalShadowFadingDirect(d geom.Pos) float64 {

	posx := float64(d.X) * facr * shadow_sampling
	posy := float64(d.Y) * facr * shadow_sampling
	var rx, ry, val float64
	for i := 0; i < len(s.xcos); i++ {
		rx += posx
		ry += posy
		val += s.xcos[i]*(math.Cos((rx + s.xsin[i]))) + s.ycos[i]*(math.Cos((ry + s.ysin[i])))
	}

	return math.Pow(10, val/10)
}
