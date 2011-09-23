
package synstation

import "math"
import "geom"
import "fmt"

func optimizePowerNone(dbs *DBS){


}

func  optimizePowerAllocationAgent(dbs *DBS) {

	var meanPtot, meanPd, meanPtotPd, meanPePd, meanPr, meanPe float64
	//var meanBER float64

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		M := c.E
		meanPtotPd += M.BERT() / M.Req()
		meanPePd += c.BER / M.Req()
		meanPtot += M.BERT()
		meanPe += c.BER
		meanPd += M.Req()
		meanPr += c.Pr
	}

	nbconnec := float64(dbs.Connec.Len())

	meanPtot /= nbconnec
	meanPtotPd /= nbconnec
	meanPePd /= nbconnec
	meanPe /= nbconnec
	meanPd /= nbconnec
	meanPr /= nbconnec

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		M := c.E

		if c.Status == 0 && !c.E.IsSetARB(0) { // if master connection	

			var b, need, delta, alpha float64

			b = M.BERT() / M.Req() // meanPtotPd
			//b = M.BERT() / M.Req() / meanPtotPd
			b = b * 5
			alpha = 1.0
			need = 2.0*math.Exp(-b)*(b+1.0) - 1.0

			//need = .5 - math.Atan(5*(b-1))/math.Pi

			delta = math.Pow(geom.Abs(need), 1) *
				math.Pow(M.GetPower(), 1) *
				geom.Sign(need-M.GetPower()) * alpha *
				math.Pow(geom.Abs(need-M.GetPower()), 1.5)

			if math.IsNaN(delta) {
				fmt.Println("delta NAN", need, M.GetPower(), b, M.BERT())
				delta = -1
			}

			if delta > 0 {
				v := (1.0 - M.GetPower()) / 2.0
				if delta > v {
					delta = v
				}
			} else {
				v := -M.GetPower() / 2.0
				if delta < v {
					delta = v
				}
			}

			if delta > 1 || delta < -1 {
				fmt.Println("Power Error ", delta)
			}

			M.PowerDelta(delta)

			//M.SetPower(1.0/math.Pow(2.0+dbs.R.Pos.Distance(M.GetPos()),4 ) )
		}

	}

}


func  optimizePowerAllocationSimple(dbs *DBS) {

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		M := c.E

		if c.Status == 0 {
			if !c.E.IsSetARB(0) { // if master connection and transmitting data

				L := float64(1500)
				d := (L - dbs.R.GetPos().Distance(M.GetPos())) / L
				var p float64
				if d > 0 {
					//p=1-math.Pow(d,1)
					p = 1 - d*d*d
				} else {
					p = 1
				}

				//p:=0.001*(math.Pow(dbs.R.GetPos().Distance(M.GetPos()),4)/100000)

				M.SetPower(p)

			} else {
				M.SetPower(1)
			}

		}

	}

}



