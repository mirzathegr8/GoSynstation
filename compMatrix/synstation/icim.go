package synstation

import "geom"
import "math"

func ICIMNone(r *geom.Pos, E *Emitter, Metric []float64, Color int) {
}

func ICIM(r *geom.Pos, E *Emitter, Metric []float64, Color int) {

	distRatio := r.DistanceSquare(E.Pos) / (IntereNodeBDist * IntereNodeBDist)

	if distRatio > ICIMdistRatio {

		step := float64(NCh) / 3.0

		if ICIMTheta || ICIMtype == ICIMc {
			p := geom.Pos{E.X - r.X, E.Y - r.Y}
			theta := math.Atan2(p.Y, p.X)

			Color = 0
			if geom.Abs(theta) < math.Pi/3.0 {
				Color = 0
			} else if theta > math.Pi/3 {
				Color = 1
			} else if theta < -math.Pi/3 {
				Color = 2
			}
		}

		jMin := NChRes
		jMax := NCh
		jMin += int(step * float64(Color))
		jMax -= int(step * (2.0 - float64(Color)))
		for rb := 0; rb < jMin; rb++ {
			Metric[rb] = 0.0
		}
		for rb := jMax; rb < NCh; rb++ {
			Metric[rb] = 0.0
		}

	}
}

func ICIMSplitEdgeCenter(r *geom.Pos, E *Emitter, Metric []float64, Color int) {

	distRatio := r.DistanceSquare(E.Pos) / (IntereNodeBDist * IntereNodeBDist)

	jMin := NChRes + 40 + Color*20
	jMax := jMin + 20

	if ICIMTheta || ICIMtype == ICIMc {
		p := geom.Pos{E.X - r.X, E.Y - r.Y}
		theta := math.Atan2(p.Y, p.X)

		Color = 0
		if geom.Abs(theta) < math.Pi/3.0 {
			Color = 0
		} else if theta > math.Pi/3 {
			Color = 1
		} else if theta < -math.Pi/3 {
			Color = 2
		}
	}

	if distRatio > ICIMdistRatio {

		for rb := 0; rb < jMin; rb++ {
			Metric[rb] = 0.0
		}
		for rb := jMax; rb < NCh; rb++ {
			Metric[rb] = 0.0
		}

	} else {

		for rb := jMin; rb < jMax; rb++ {
			Metric[rb] = 0.0
		}
		//		for rb:=jMax;rb<NCh;rb++ {Metric[rb]=0.0}

	}
}

func ICIMSplitEdgeCenter2(r *geom.Pos, E *Emitter, Metric []float64, Color int) {

	distRatio := r.DistanceSquare(E.Pos) / (IntereNodeBDist * IntereNodeBDist)

	jMin := NChRes + Color*33
	jMax := jMin + 33

	if ICIMTheta || ICIMtype == ICIMc {
		p := geom.Pos{E.X - r.X, E.Y - r.Y}
		theta := math.Atan2(p.Y, p.X)

		Color = 0
		if geom.Abs(theta) < math.Pi/3.0 {
			Color = 0
		} else if theta > math.Pi/3 {
			Color = 1
		} else if theta < -math.Pi/3 {
			Color = 2
		}
	}

	if distRatio > ICIMdistRatio {

		for rb := 0; rb < jMin; rb++ {
			Metric[rb] = 0.0
		}
		for rb := jMax; rb < NCh; rb++ {
			Metric[rb] = 0.0
		}

	} else {

		for rb := jMin; rb < jMax; rb++ {
			Metric[rb] = 0.0
		}
		//		for rb:=jMax;rb<NCh;rb++ {Metric[rb]=0.0}

	}
}


