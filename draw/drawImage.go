package draw

import (
	"image/color"
	"synstation"
)
import "math"
import "geom"
import "image"
import "image/png"
//import "log"
import "os"
import "fmt"
//import "rand"

const Size = 600

func DrawReceptionField(dbs []synstation.DBS, name string) {

	var data [Size][Size]float64

	//dbs:= make([]synstation.DBS,1)
	//dbs[0].R.SetPos(geom.Pos{6000,6000})


	e := new(synstation.Emitter)
	e.Power[0] = 1

	var TMax, TMin float64
	TMax = -300
	TMin = +300

	sums := float64(0)
	for x := 0; x < Size; x++ {

		e.Pos.X = inField(x)
		sumy := float64(0)
		for y := 0; y < Size; y++ {
			e.Pos.Y = inField(y)
			Pr := float64(0.0)
			for k := 0; k < len(dbs); k++ {

				/*	p:= e.GetPos().Minus(*dbs[k].R.GetPos())
						theta := math.Atan2(p.Y,p.X) * 180/ math.Pi
						if theta<0 {theta=theta+360}	
					if theta <0 {fmt.Println(" theta <0!!!!")}	
						dbs[k].R.Orientation[ch]=theta //+ (dbs.Rgen.Float64()*30-15)
				*/

				if geom.Abs(inField(x)-dbs[k].R.GetPos().X) < 1500 &&
					geom.Abs(inField(y)-dbs[k].R.GetPos().Y) < 1500 {

					//TODO need to change here the call to fading to a paralle
					// call to GenFastFading 
					// on many emitters, and read GetPr
					//	p := dbs[k].R.Fading(e.Pos, e.Ch)
					p, _ := dbs[k].R.GetPr(0, 0)
					if p > Pr {
						Pr = p
					}

				}

			}
			Pr = 10 * math.Log10(Pr)
			data[x][y] = Pr
			sumy += Pr
			if Pr > TMax {
				TMax = Pr
			}
			if Pr < TMin {
				TMin = Pr
			}
		}
		sums += sumy
	}

	fmt.Println("mean ", sums/600/600)
	fmt.Println("TMax ", TMax)
	fmt.Println("TMin ", TMin)
	if TMin < -150 {
		TMin = -150
	}

	// -15 -125
	// -5 -115
	// -9 -117

	im := image.NewNRGBA(image.Rect(0, 0, Size, Size))
	for x := 0; x < Size; x++ {
		for y := 0; y < Size; y++ {
			var r, g, b float64
			v := (data[x][y] - TMin) / (TMax - TMin) * -8.0
			switch {
			case v > -2.0:
				b = 2.0 + v
				g = -v
			case v > -4:
				g = 1
				r = -v/2.0 - 1.0
			default:
				r = 1.0
				b = -v/2.0 - 2.0
				g = 1.0 - b
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
			im.Set(x, y, color.NRGBA{uint8(r), uint8(g), uint8(b), uint8(255)})
			//im.Pix[y*im.Stride+x] = image.NRGBAColor{uint8(v*255),uint8(v*255),uint8(v*255),255}
		}
	}

	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, 0666)

	if err = png.Encode(f, im); err != nil {
		os.Exit(-1)
	}

}

func inField(x int) (a float64) {
	a = synstation.Field / Size * float64(x)
	return
}
