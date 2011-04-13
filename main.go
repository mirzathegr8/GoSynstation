package main


import "fmt"
import s "synstation"
import "runtime"
//import "draw"
//import "math"
//import "geom"
import "time"
//import "bytes"


// Data to save for output during simulation
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


func main() {

	runtime.GOMAXPROCS(14)

	var outD outputData  // one to print
	var outDs outputData // one to sum and take mean over simulation

	// initialize var
	outChannel = make(chan outputData, 8000)

	var n = 0

	s.Init()

	//draw.Init(s.M, s.D*s.NConnec) // init drawing system

	fmt.Println("Init done")

	//draw.DrawReceptionField(Synstations[:], "receptionLevel.png")

	go printData() //launch thread to print output

	var ttps int64

	fmt.Println("Start Simulation")

	//fadingF, err := os.Open("fading.mat", os.O_WRONLY|os.O_CREATE, 0666)
	//fmt.Println(err)
	//fadingF.WriteString(fmt.Sprintln("# name: fading\n# type: matrix\n# rows: ", s.Duration, "\n# columns: ", s.NCh))


	s.ChannelHop()

	// precondition

	for k := -1000; k < 0; k++ {

		tps := time.Nanoseconds()

		for i := range s.Synstations {
			go s.Synstations[i].RunPhys()
		}

		//	for i := range s.Mobiles {
		//		go s.Mobiles[i].RunPhys()
		//	}

		//synchronise here
		//	s.Sync(s.D + s.M)
		s.Sync(s.D)

		ttps += (time.Nanoseconds() - tps)

		// physics is done, now launch Mobiles data work
		for i := range s.Mobiles {
			go s.Mobiles[i].FetchData()
		}

		//here we synchronise threads and fetch data for ouput at the same time
		outD.connected, outD.BER1, outD.BER2, outD.BER3 = 0, 0, 0, 0

		n = 0
		for v := range s.SyncChannel {
			n++

			//fmt.Print(s," ")	
			switch {
			case v < -3:
				outD.BER1++
				fallthrough
			case v < -2:
				outD.BER2++
				fallthrough
			case v < -1:
				outD.BER3++
				fallthrough
			case v < -0.01:
				outD.connected++
			}
			if n >= s.M {
				break
			}

		}

		//geting a bit more data
		outD.d_connec = float64(s.GetConnect())
		outD.d_discon = float64(s.GetDisConnect())
		outD.Diversity = float64(s.GetDiversity()) / float64(s.M)
		outD.lost = float64(s.SystemChan[0].GetAdded())
		outD.d_lost = float64(s.GetLostConnect())
		outD.HopCount = float64(s.GetHopCount())

		outD.k = float64(k)
		if k%10 == 0 {
			outChannel <- outD //sent data to print to  stdout			
		}

		//Run DBS Agent, and sync
		for i := range s.Synstations {
			go s.Synstations[i].RunAgent()
		}

		//sync
		s.Sync(s.D)
		s.ChannelHop()

	}

	// simu
	for k := 0; k < s.Duration; k++ {

		//	fmt.Println(" Channel 0 ", s.SystemChan[0].Emitters.Len())
		tps := time.Nanoseconds()

		for i := range s.Synstations {
			go s.Synstations[i].RunPhys()
		}

		//	for i := range s.Mobiles {
		//		go s.Mobiles[i].RunPhys()
		//	}

		//synchronise here
		//	s.Sync(s.D + s.M)
		s.Sync(s.D)

		ttps += (time.Nanoseconds() - tps)

		// physics is done, now launch Mobiles data work
		for i := range s.Mobiles {
			go s.Mobiles[i].FetchData()
		}

		//here we synchronise threads and fetch data for ouput at the same time
		outD.connected, outD.BER1, outD.BER2, outD.BER3 = 0, 0, 0, 0

		n = 0
		for v := range s.SyncChannel {
			n++

			//fmt.Print(s," ")	
			switch {
			case v < -3:
				outD.BER1++
				fallthrough
			case v < -2:
				outD.BER2++
				fallthrough
			case v < -1:
				outD.BER3++
				fallthrough
			case v < -0.01:
				outD.connected++
			}
			if n >= s.M {
				break
			}

		}

		//geting a bit more data
		outD.d_connec = float64(s.GetConnect())
		outD.d_discon = float64(s.GetDisConnect())
		outD.Diversity = float64(s.GetDiversity()) / float64(s.M)
		outD.lost = float64(s.SystemChan[0].GetAdded())
		outD.d_lost = float64(s.GetLostConnect())
		outD.HopCount = float64(s.GetHopCount())

		outDs.Add(&outD)

		outD.k = float64(k)
		if k%10 == 0 {
			outChannel <- outD //sent data to print to  stdout			
		}

		if k%200 == 0 {
			runtime.GC()
		}

		t := s.CreateTrace(s.Mobiles[:], s.Synstations[:], k)
		//draw.Draw(t)
		sendTrace(t)

		/*if s.Mobiles[10].MasterConnection != nil {
			ffR := s.Mobiles[10].MasterConnection.GetffR()
			buffer := bytes.NewBufferString("")
			for _, a := range ffR {
				fmt.Fprint(buffer, a)
				fmt.Fprint(buffer, " ")
			}
			fadingF.WriteString(string(buffer.Bytes()))

		} else {

			buffer := bytes.NewBufferString("")
			for i := 0; i < s.NCh; i++ {
				fmt.Fprint(buffer, float64(0.0))
				fmt.Fprint(buffer, " ")
			}
			fadingF.WriteString(string(buffer.Bytes()))
		}

		fadingF.WriteString("\n")*/

		//Run DBS Agent, and sync
		for i := range s.Synstations {
			go s.Synstations[i].RunAgent()
		}

		//sync
		s.Sync(s.D)

		s.ChannelHop()
		//s.PowerC(Synstations[:])

	}

	// Print some status data

	outDs.Div(float64(s.Duration))
	fmt.Println("Mean", outDs.String())

	for i := range s.Synstations {
		fmt.Print(" ", s.Synstations[i].Connec.Len())
	}
	fmt.Println()

	for i := range s.SystemChan {
		fmt.Print(" ", s.SystemChan[i].Emitters.Len())
	}
	fmt.Println()

	SaveToFile(s.Mobiles[:])

	//And finaly close channels and background processes

	close(outChannel)
	for ch := range s.SystemChan {
		close(s.SystemChan[ch].Change)
	}
	close(s.SyncChannel)

	StopSave()
	fmt.Println(ttps)

	//draw.Close()

}

