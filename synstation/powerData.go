package synstation


type PowerData struct {
	pr [M]float64
}


func (p *PowerData) CalculatePr(rx *PhysReceiver, FF FadingData ) {

	for i := 0; i < M; i++ {

		E:=&Mobiles[i]
		gain := rx.GainBeam(E, E.GetCh())
		fading, _ := rx.Fading(E)
		p.pr[i] = fading * gain * E.GetPower()

	}
}

func (p *PowerData) GetPr(m int) float64 {
	return p.pr[m]
}

