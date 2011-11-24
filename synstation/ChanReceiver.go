package synstation

import "fmt"




// number of signal id saved in the list of the ChanReceiver
const SizeES = 2

// Structure to hold interference level for a RB, and multilevel interference with overlaping channels calculation
// as well as a list of ordered strongest signal received 
type ChanReceiver struct {
	Pint     float64 // to store total received power level including interference;
	Pint1lvl float64 // to store received power_levels of emitters in channel without co-interference 

	Signal [SizeES]int // used to store ordered list of ids of most important emitters interfering in this RB

	pr [M]float64 //stores received power with shadowing and power-level and distance atenuation
}

//This function is called to reset the list of received signals
//and the sum of powerlevel received on the RB
func (chR *ChanReceiver) Clear() {
	for i := range chR.Signal {
		chR.Signal[i] = -1
	}
	chR.Pint = 0
	chR.Pint1lvl = 0
}

//This function is called by the interference evaluation while summing the received power,
// in order to save the ordered list of the strongest received powers
func (chR *ChanReceiver) Push(S int, P float64) {
	var i = 0
	var j = 0
	for i = 0; i < SizeES; i++ {
		if chR.Signal[i] < 0 {
			chR.Signal[i] = S
			return
		}
		if chR.pr[chR.Signal[i]] < P {
			break
		}
	}
	if i < SizeES {
		for j = i + 1; j < SizeES-1; j++ {
			if chR.Signal[j] < 0 {
				break
			}
			chR.Signal[j] = chR.Signal[j-1]
		}
		chR.Signal[i] = S
	}

}

// Simple output to print the value of interference on the RB
func (chR ChanReceiver) String() string { return fmt.Sprintf("{%f }", chR.Pint*1e15) }


type RBsReceiver struct {
	Channels [NCh]ChanReceiver
}


func (rbs * RBsReceiver) SumInterference(){

	for  i:= range rbs.Channels {
		rbs.Channels[i].Clear()
	}

	for i := range rbs.Channels {
		chR := &rbs.Channels[i]
		for e := SystemChan[i].Emitters.Front(); e != nil; e = e.Next() {
			c := e.Value.(*Emitter)
			m := c.GetId()
			//chR.Pint1lvl += chR.pr[m]
			chR.Pint += chR.pr[m]
			chR.Push(m, chR.pr[m])
		}
	}


	//For now, no cochannel
	/*for i := 0; i < NCh; i++ {
		co.Channels[i].Pint = co.Channels[i].Pint1lvl
		for _, coc := range SystemChan[i].coIntC {
			rx.Channels[i].Pint += coc.factor * rx.Channels[coc.c].Pint1lvl
		}
	}*/




}

