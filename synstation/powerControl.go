package synstation

import "math"
import "geom"
import "fmt"

func optimizePowerNone(dbs *DBS) {

}

func optimizePowerAllocationAgent(dbs *DBS) {

	var meanPtot, meanPd, meanPtotPd, meanPePd, meanPr, meanPe float64
	//var meanBER float64

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		M := c.E
		meanPtotPd += M.BERT() / M.Req()
		meanPePd += c.meanBER.Get() / M.Req()
		meanPtot += M.BERT()
		meanPe += c.meanBER.Get()
		meanPd += M.Req()
		meanPr += c.meanPr.Get()
	}

	nbconnec := float64(dbs.Connec.Len())

	meanPtot /= nbconnec
	meanPtotPd /= nbconnec
	meanPePd /= nbconnec
	meanPe /= nbconnec
	meanPd /= nbconnec
	meanPr /= nbconnec

	for rb := 1; rb < NCh; rb++ {

		for e := dbs.Connec.Front(); e != nil; e = e.Next() {
			c := e.Value.(*Connection)
			M := c.E

			if c.Status == 0 && !c.E.IsSetARB(0) { // if master connection	

				var b, need, delta float64

				b = M.BERT() / M.Req() // meanPtotPd
				//b = M.BERT() / M.Req() / meanPtotPd
				b = b * PowerAgentFact
				need = 2.0*math.Exp(-b)*(b+1.0) - 1.0

				//need = .5 - math.Atan(5*(b-1))/math.Pi

				delta = math.Pow(geom.Abs(need), 1) *
					math.Pow(M.GetPower(rb), 1) *
					geom.Sign(need-M.GetPower(rb)) * PowerAgentAlpha *
					math.Pow(geom.Abs(need-M.GetPower(rb)), 1.5)

				if math.IsNaN(delta) {
					fmt.Println("delta NAN", need, M.GetPower(rb), b, M.BERT())
					delta = -1
				}

				if delta > 0 {
					v := (1.0 - M.GetPower(rb)) / 2.0
					if delta > v {
						delta = v
					}
				} else {
					v := -M.GetPower(rb) / 2.0
					if delta < v {
						delta = v
					}
				}

				if delta > 1 || delta < -1 {
					fmt.Println("Power Error ", delta)
				}

				M.PowerDelta(rb, delta)

				//M.SetPower(1.0/math.Pow(2.0+dbs.R.Pos.Distance(M.GetPos()),4 ) )
			}

		}
	}

}

func optimizePowerAllocationAgentRB(dbs *DBS) {

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		M := c.E

		if c.Status == 0 && !M.IsSetARB(0) { // if master connection	

			for rb := 1; rb < NCh; rb++ {

				if M.IsSetARB(rb) {

					var b, need, delta float64

					b = -M.SNRrb[rb] / M.Req()
					b = b * PowerAgentFact
					//PowerAgentAlpha = 1.0
					need = 2.0*math.Exp(-b)*(b+1.0) - 1.0

					//need = .5 - math.Atan(5*(b-1))/math.Pi

					delta = math.Pow(geom.Abs(need), 1) *
						math.Pow(M.GetPower(rb), 1) *
						geom.Sign(need-M.GetPower(rb)) * PowerAgentAlpha *
						math.Pow(geom.Abs(need-M.GetPower(rb)), 1.5)

					if math.IsNaN(delta) {
						fmt.Println("delta NAN", need, M.GetPower(rb), b, M.SBERrb[rb])
						delta = -1
					}

					if delta > 0 {
						v := (1.0 - M.GetPower(rb)) / 2.0
						if delta > v {
							delta = v
						}
					} else {
						v := -M.GetPower(rb) / 2.0
						if delta < v {
							delta = v
						}
					}

					if delta > 1 || delta < -1 {
						fmt.Println("Power Error ", delta)
					}

					M.PowerDelta(rb, delta)

				} else {
					M.SetPowerRB(rb,1);
					

				}
			}

		}
	}

}

func optimizePowerAllocationSimple(dbs *DBS) {

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		E := c.E

		if c.Status == 0 {
			if !c.E.IsSetARB(0) { // if master connection and transmitting data

				L := float64(1500)
				d := (L - dbs.Pos.Distance(E.Pos)) / L
				var p float64
				if d > 0 {
					//p=1-math.Pow(d,1)
					p = 1 - d*d*d
				} else {
					p = 1
				}

				//p:=0.001*(math.Pow(dbs.R.GetPos().Distance(M.GetPos()),4)/100000)

				E.SetPower(p)

			} else {
				E.SetPower(1)
			}

		}

	}

}

func PowerICIM(dbs *DBS) {
	//power allocation comes after RB setting

	r := dbs.Pos
	jMin := NChRes + ((1+dbs.Color)%3)*33 //for non priviledge band
	jMax := jMin + 33

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		M := c.E

		if c.Status == 0 {

			distRatio := r.DistanceSquare(M.GetPos()) / (IntereNodeBDist * IntereNodeBDist)

			// if edge region set power to one
			if distRatio > ICIMdistRatio {
				M.SetPower(1)

			} else { //if center region


				for rb := NChRes; rb < NCh; rb++ {

					if M.IsSetARB(rb) {
						// set power to low if outside previliedge range
						if rb < jMin || rb >= jMax {
							M.SetPowerRB(rb, 0.1)
						} else { //set power to high inside privilede range
							M.SetPowerRB(rb, 1.0)
						}

					}

				}
			}

		}

	}

}

//   Reformatted by   lerouxp    Tue Nov 1 13:12:39 CET 2011

