package main


import "fmt"


func init() {

	// initialize var
	outChannel = make(chan outputData, 8000)
}

type outputData struct {
	k, listened, connected, BER1, BER2, BER3 float64
	d_connec, d_discon, d_lost               float64
	lost                                     float64
	Diversity                                float64
	HopCount                                 float64
	Fairness				float64
	SumTR					float64
	MUFactor				float64
}

func (o *outputData) Add(o2 *outputData) {
	o.connected += o2.connected
	o.listened += o2.listened
	o.BER1 += o2.BER1
	o.BER2 += o2.BER2
	o.BER3 += o2.BER3
	o.d_connec += o2.d_connec
	o.d_discon += o2.d_discon
	o.d_lost += o2.d_lost
	o.lost += o2.lost
	o.Diversity += o2.Diversity
	o.HopCount += o2.HopCount
	o.Fairness += o2.Fairness
	o.SumTR += o2.SumTR
	o.MUFactor += o2.MUFactor
}


func (o *outputData) Div(k float64) {
	o.listened /= k
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
	o.Fairness/=k
	o.SumTR /=k
	o.MUFactor /=k

}

func (o outputData) String() string {
//	return fmt.Sprint(o.listened, "	", o.connected, "	",
//		o.BER1, "	",		
//		o.BER3, "	",
//		o.d_connec, "	",
//		o.d_discon, "	",
//		o.lost, "	",
//		o.d_lost, "	",
//		o.Diversity, "	",
//		o.HopCount, "	",
//		o.Fairness, "	",
//		o.SumTR, "	",
//		o.MUFactor)

return fmt.Sprintf("%d	%d	%d	%d	%d	%d	%d	%d	%1.2f	%d	%1.2f	%1.2e	%4.2f",
		int(o.listened), 
		int(o.connected),
		int(o.BER1), 		
		int(o.BER3), 
		int(o.d_connec), 
		int(o.d_discon), 
		int(o.lost), 
		int(o.d_lost), 
		o.Diversity, 
		int(o.HopCount), 
		o.Fairness, 
		o.SumTR, 
		o.MUFactor)


}


var outChannel chan outputData


func printData() {

	for r := range outChannel {
		fmt.Println(r.k, "	", r.String())
	}

}

