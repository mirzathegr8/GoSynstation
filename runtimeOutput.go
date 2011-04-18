package main


import "fmt"


type outputData struct {
	k, connected, BER1, BER2, BER3 float64
	d_connec, d_discon, d_lost     float64
	lost                           float64
	Diversity                      float64
	HopCount                       float64
}

func (o *outputData) Add(o2 *outputData) {
	o.connected += o2.connected
	o.BER1 += o2.BER1
	o.BER2 += o2.BER2
	o.BER3 += o2.BER3
	o.d_connec += o2.d_connec
	o.d_discon += o2.d_discon
	o.d_lost += o2.d_lost
	o.lost += o2.lost
	o.Diversity += o2.Diversity
	o.HopCount += o2.HopCount
}


func (o *outputData) Div(k float64) {
	o.connected /= k
	o.BER1 /= k
	o.BER2 /= k
	o.BER3 /= k
	o.d_connec /= k
	o.d_discon /= k
	o.lost /= k
	o.Diversity /= k
	o.d_lost /= k
	o.HopCount /= k
}

func (o outputData) String() string {
	return fmt.Sprint(o.connected, "	",
		o.BER1, "	",
		o.BER2, "	",
		o.BER3, "	",
		o.d_connec, "	",
		o.d_discon, "	",
		o.lost, "	",
		o.d_lost, "	",
		o.Diversity, "	",
		o.HopCount)

}


var outChannel chan outputData


func printData() {

	for r := range outChannel {
		fmt.Println(r.k, "	", r.String())
	}

}

