package draw

import "gocairo"
//import "time"
import "synstation"
import "fmt"
import "math"
import "container/vector"
import "geom"
import "runtime"

const numImThread = 1


//the dataset structure to draw, with its surface/context to draw to
type DataSave struct {
	t       *synstation.Trace
	ackChan chan int
	cv      gocairo.Canvas
}


var sentData chan *DataSave

//common cairo object to draw to


//initialises the drawing threads wiht according datasets memory allocated for the size of the data to save (mobilenumber and conneciton Number)
func Init(MobileNumber int, ConnectionNumber int) {

	sentData = make(chan *DataSave, numImThread)
	for i := 0; i < numImThread; i++ {
		go drawingThread(MobileNumber, ConnectionNumber, drawP)
		//go drawingThread(MobileNumber, ConnectionNumber,drawP)
	}
}

//Final function that waits for all drawing thread to finish execution and closes the datasets (ie calls a function to do something of the last dataset and drawing context/surface)
func Close() {
	d := make([]*DataSave, numImThread)
	for i := 0; i < numImThread; i++ {
		d[i] = <-sentData //get allwait for all threads to end
	}
	for i := 0; i < numImThread; i++ {
		d[i].Close()
		close(d[i].ackChan)
	}
	close(sentData)

}


//function called to copy data to a dataset when one is available and let a drawing thread do some drawing in background
func Draw(t *synstation.Trace) {

	data := <-sentData

	data.t = t

	data.ackChan <- 1

}


// function to initialise a dataset, allocate memory, create surface/context and communication channel
func (d *DataSave) Init(MobileNumber int, ConnectionNumber int) {
	d.ackChan = make(chan int)
	d.cv.Create()
}

//what to do with drawing context/surface when simulation finishes
func (d *DataSave) Close() {

	for p := 0; p < d.t.NumConn; p++ {
		c := &d.t.Connec[p]
		d.cv.SetColor(0.0, 0.0, 0.0, 1.0)
		d.cv.DrawCircle(float(c.B.X)/2.0, float(c.B.Y)/2.0, 5.0)
		d.cv.Stroke()
	}

	d.cv.Save("FinalOutput", d.t.K)
	d.cv.Close()
}

//start a drawing thread, lock it to os thread, makes a dataset
func drawingThread(MobileNumber, ConnectionNumber int, drawIt func(*DataSave)) {

	runtime.LockOSThread()

	d := new(DataSave)

	d.Init(MobileNumber, ConnectionNumber)

	sentData <- d // we are ready to draw	

	d.cv.Create()

	for _ = range d.ackChan {
		drawIt(d)
		sentData <- d //we are done
	}

}

func drawP(d *DataSave) {

	/*d.cv.Clear()

	for p := 0; p < d.t.NumConn; p++ {
			c := &d.t.Connec[p]
			d.cv.SetColor(0.5, .5, 0.5, 0.1)

			d.cv.DrawLine(float(c.A.X),
					float(c.A.Y),
					float(c.B.X),
					float(c.B.Y))
			d.cv.Stroke()	
		d.cv.SetColor(0.0, 0.0, 0.0, 1.0)
		d.cv.DrawCircle(float(c.B.X), float(c.B.Y), 5.0)
		d.cv.Stroke()
	}*/

	for i := range d.t.Mobs {

		p := &d.t.Mobs[i]

		var r, g, b, v float

		if p.BERtotal < synstation.BERThres && p.Ch > 0 {
			/*	r = 1.0
				g = 0.0
				b = 0.0
			} else {*/
			//v=float(-p.Diversity)*2.0 +2	
			//v = float((p.BERtotal - p.MaxBER) / p.BERtotal * -2.0)
			v = float(p.BERtotal) //-  (float(-p.Diversity)*2.0 +2)
			//v=float(p.MaxBER)
			//v= -3*float(math.Log10(p.SNRb))
			//v= -6*float(p.Power)

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
		}

		d.cv.SetColor(r, g, b, 0.02)

		d.cv.DrawCircle(float(p.X)/2.0, float(p.Y)/2.0, 10.0)
		d.cv.Stroke()

	}

	//d.cv.Save("test", d.t.K)


}


func drawChan(d *DataSave) {

	v := float64(d.t.K) / 1000.0 * -8.0

	var r, g, b float

	switch {

	case v > -2.0:
		b = 2.0 + float(v)/2.0
		g = -float(v)
	case v > -4:
		g = 1
		r = -float(v)/2.0 - 1.0
	case v > -6:
		r = 1.0
		b = -float(v)/2.0 - 2.0
		g = 1.0 - b
	default:
		r = 1.0
		b = 4.0 + float(v)/2.0
		g = 0.0
	}

	for i := range d.t.Mobs {

		p := &d.t.Mobs[i]

		if p.Ch == 0 || p.BERtotal > synstation.BERThres { // >= synstation.NCh-5{

			d.cv.SetColor(r, g, b, 0.2)

			r := 15 * math.Log(-30.0/p.BERtotal)

			d.cv.DrawCircle(float(p.X)/2.0, float(p.Y)/2.0, float(r))
			d.cv.Stroke()

		}

	}

}


func drawVoronoi(d *DataSave) {

	//	d.cv.Clear()

	L1 := geom.GeomLine{geom.Pos{-1000, -1000}, geom.Pos{10000, -0.001}}
	// delta to prevent some division by zero apparently happens in 
	//the top left corner of the voronoi graph
	L2 := geom.GeomLine{geom.Pos{-1000, -1000}, geom.Pos{-0.0001, 10000}}
	L3 := geom.GeomLine{geom.Pos{synstation.Field + 1000, synstation.Field + 1000}, geom.Pos{10000, 0.0001}}
	L4 := geom.GeomLine{geom.Pos{synstation.Field + 1000, synstation.Field + 1000}, geom.Pos{0.0001, 10000}}

	var neighB vector.Vector
	var Borders vector.Vector

	//channels := []int{ 10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,29,25,30}
	channels := []int{0}

	for _, Ch := range channels {

		for m := range d.t.Mobs {

			m1 := &d.t.Mobs[m]

			if m1.Ch == Ch {

				d.cv.SetColor(0, 0, 0, 1)
				neighB = neighB[0:0]
				Borders = Borders[0:0]

				for j := range d.t.Mobs {

					if j != m {
						m2 := &d.t.Mobs[j]

						if m2.Ch == Ch {

							if m1.Pos.Distance(&m2.Pos) < 8000 {

								//neighB.Push(m2)
								l := new(geom.GeomLine)
								//mp1:=float64(1.0)/m1.Power
								//mp2:=float64(1.0)/m2.Power
								//l.Org= m1.Times(mp1).Plus(m2.Times(mp2)).Times(1/(mp1+mp2))

								vd := m2.Pos.Minus(m1.Pos)
								dist := vd.Len()
								al := math.Pow(m1.Power/m2.Power, .25)
								q := (al*(dist+2) - 2) / (1 + al)
								l.Org = m1.Pos.Plus(vd.Times(q / dist))

								l.Vect = m2.Minus(m1.Pos).Rot90()
								Borders.Push(l)

								neighB.Push(m2)

							}
						}
					}
				}
				Borders.Push(&L1)
				Borders.Push(&L2)
				Borders.Push(&L3)
				Borders.Push(&L4)

				matP := make([][]geom.Pos, len(Borders))

				for i := range Borders {

					matP[i] = make([]geom.Pos, len(Borders))

					for j := range Borders {
						if i != j {
							matP[i][j] = Borders[i].(*geom.GeomLine).Intersect(Borders[j].(*geom.GeomLine))
						}
					}
				}

				//find clossest intersection
				max := float64(synstation.Field * 2.0)
				var I, J int

				for i := range Borders {
					gl := Borders[i].(*geom.GeomLine)
					d := m1.Pos.Distance(&gl.Org)
					if d < max {
						I = i
						max = d
					}
				}

				vOrg := Borders[I].(*geom.GeomLine).Org
				max = float64(synstation.Field * 2.0)
				for j := range Borders {
					d := matP[I][j].Distance(&vOrg)
					if d < max {
						J = j
						max = d
					}

				}

				var PathVoronoi vector.Vector

				PStart := &matP[I][J]
				PathVoronoi.Push(PStart)

				// try if we start on I or J

				vectO := m1.Pos.Minus(*PStart).Normalise()
				vI := Borders[I].(*geom.GeomLine).Vect
				vJ := Borders[J].(*geom.GeomLine).Vect

				c1 := vectO.Scalar(vI) / vI.Len()
				s1 := vectO.Prodv(vI) / vI.Len()
				a1 := math.Atan2(s1, c1)
				c2 := vectO.Scalar(vJ) / vJ.Len()
				s2 := vectO.Prodv(vJ) / vJ.Len()
				a2 := math.Atan2(s2, c2)

				if a1 > math.Pi/2 {
					a1 = -math.Pi + a1
				}
				if a1 < -math.Pi/2 {
					a1 = a1 + math.Pi
				}
				if a2 > math.Pi/2 {
					a2 = -math.Pi + a2
				}
				if a2 < -math.Pi/2 {
					a2 = a2 + math.Pi
				}

				if a1 < 0 && a2 < 0 && a1 > a2 {
					K := I
					I = J
					J = K
				} else if a1 < 0 && a2 > 0 {
					K := I
					I = J
					J = K
				} else if a2 < a1 && a2 > 0 {
					K := I
					I = J
					J = K
				}

				jnext := -1

				for i, j := I, J; jnext != I && len(PathVoronoi) < 10; {
					//find clossest in direction to center

					dm := float64(synstation.Field * 2.0)
					jnext = -1
					vectO := m1.Pos.Minus(*PStart)
					for k := range Borders {
						if k != i && k != j {

							vect2 := matP[i][k].Minus(*PStart)
							dist := vect2.Len()
							scalY := vectO.Prodv(vect2) / vectO.Len() / vect2.Len()
							if scalY > 0 && dist < dm {
								dm = dist
								jnext = k
							}

						}
					}

					if jnext < 0 {
						fmt.Println("index out of range")
						d.cv.SetColor(1, 0, 0, 1)
						d.cv.DrawCircle(float(m1.Pos.X)/2.0, float(m1.Pos.Y)/2.0, 40.0)
						break
					}
					a := i
					i = jnext
					j = a

					PStart = &matP[i][j]
					PathVoronoi.Push(PStart)

				}

				d.cv.MoveTo(PathVoronoi[0].(*geom.Pos).Times(1 / 20.0).ToFloat())
				for i := 1; i < len(PathVoronoi); i++ {
					d.cv.LineTo(PathVoronoi[i].(*geom.Pos).Times(1 / 20.0).ToFloat())

				}
				d.cv.ClosePath()

				p := m1
				var r, g, b float
				//v=float(-p.Diversity)*2.0 +2	
				//v=float((p.BERtotal-p.MaxBER)/p.BERtotal* -2.0)
				//v := float(p.BERtotal) //-  (float(-p.Diversity)*2.0 +2)
				//v=float(p.MaxBER)
				v := -3 * float(math.Log10(p.SNRb))
				//v= -6*float(p.Power)

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

				d.cv.SetColor(r, g, b, 0.05)
				//d.cv.FillPreserve()
				d.cv.Fill()

				//d.cv.SetColor(0, 0, 0, 0.01)
				//d.cv.Stroke()

				/*d.cv.SetColor(0, 0, 0, 1)
				d.cv.DrawCircle(float(m1.Pos.X)/2.0, float(m1.Pos.Y)/2.0, 10.0)
				d.cv.Stroke()*/

			}
		}

		/*for p := 0; p < d.NumConn; p++ {
			c := &d.Connec[p]
			if c.Ch == Ch {

				if c.BER < math.Log10(0.5) {
					d.cv.SetColor(0.5, .5, 0.5, 1)
				} else {
					d.cv.SetColor(0.0, .5, 0.5, 1)
				}

				d.cv.DrawLine(float(c.A.X)/2,
					float(c.A.Y)/2,
					float(c.B.X)/2,
					float(c.B.Y)/2)
				d.cv.Stroke()
			}
			d.cv.SetColor(0.5, 0.5, 1.0, 1.0)
			d.cv.DrawCircle(float(c.B.X)/2, float(c.B.Y)/2, 15.0)
			d.cv.Stroke()
		}*/

	}
	//d.cv.Save("0vornoi", d.t.K)

}

