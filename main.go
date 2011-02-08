package main


import "fmt"
import "synstation"
import "runtime"
import "draw"
import "os"
//import "math"
//import "geom"


// Data to save for output during simulation
type outputData struct {
	k, connected, BER1, BER2, BER3 float
	d_connec, d_discon, d_lost     float
	lost                           float
	Diversity                      float
	HopCount                       float
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


func (o *outputData) Div(k float) {
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

	var synstations [synstation.D]synstation.DBS
	var mobiles [synstation.M]synstation.Mob

	// initialize var
	outChannel = make(chan outputData, 8000)

	for i := range synstations {
		go synstations[i].Init()
	}
	//sync
	n := 0
	for _ = range synstation.SyncChannel {
		n++
		if n >= synstation.D {
			break
		}
	}

	/*d:=0
	nD := int(math.Sqrt(synstation.D))
	for i:=0;i< nD ;i++{
		for j:=0;j < nD;j++{
			x:=synstation.Field/float64(nD)*(float64(i)+ .5*float64(j%2) )
			y:=synstation.Field/float64(nD)*(float64(j)+.5)
			synstations[d].R.SetPos(geom.Pos{x,y})
			d++
		}
	}*/

	for i := range mobiles {
		mobiles[i].Init()
	}

	synstation.SystemChan[0].GetAdded() //reset adder 

	draw.Init(synstation.M, synstation.D*synstation.NConnec) // init drawing system

	fmt.Println("Init done")

	draw.DrawReceptionField(synstations[:], "receptionLevel.png")

	go printData() //launch thread to print output

	//heating up simulation
	for k := -200; k < 0; k++ {

		for i := range synstations {
			go synstations[i].RunPhys()
		}

		for i := range mobiles {
			go mobiles[i].RunPhys()
		}

		//synchronise here
		n = 0
		for _ = range synstation.SyncChannel {
			n++
			if n >= synstation.D+synstation.M {
				break
			}
		}

		// physics is done, now launch mobiles data work
		for i := range mobiles {
			go mobiles[i].FetchData()
		}

		//here we synchronise threads and fetch data for ouput at the same time
		outD.connected, outD.BER1, outD.BER2, outD.BER3 = 0, 0, 0, 0

		n = 0
		for s := range synstation.SyncChannel {
			n++

			//fmt.Print(s," ")	
			switch {
			case s < -3:
				outD.BER1++
				fallthrough
			case s < -2:
				outD.BER2++
				fallthrough
			case s < -1:
				outD.BER3++
				fallthrough
			case s < -0.01:
				outD.connected++
			}
			if n >= synstation.M {
				break
			}

		}

		//geting a bit more data
		outD.d_connec = float(synstation.GetConnect())
		outD.d_discon = float(synstation.GetDisConnect())
		outD.Diversity = float(synstation.GetDiversity()) / float(synstation.M)
		outD.lost = float(synstation.SystemChan[0].GetAdded())
		outD.d_lost = float(synstation.GetLostConnect())
		outD.HopCount = float(synstation.GetHopCount())

		if k%10 == 0 {
			outD.k = float(k)
			outChannel <- outD //sent data to print to  stdout		
		}

		//Run DBS Agent, and sync
		for i := range synstations {
			go synstations[i].RunAgent()
		}

		//sync
		n = 0
		for _ = range synstation.SyncChannel {
			n++
			if n >= synstation.D {
				break
			}
		}

		//synstation.PowerC(synstations[:])

	}

	//end of preset

	for k := 0; k < synstation.Duration; k++ {

		for i := range synstations {
			go synstations[i].RunPhys()
		}

		for i := range mobiles {
			go mobiles[i].RunPhys()
		}

		//synchronise here
		n := 0
		for _ = range synstation.SyncChannel {
			n++
			if n >= synstation.D+synstation.M {
				break
			}
		}

		// physics is done, now launch mobiles data work
		for i := range mobiles {
			go mobiles[i].FetchData()
		}

		//here we synchronise threads and fetch data for ouput at the same time
		outD.connected, outD.BER1, outD.BER2, outD.BER3 = 0, 0, 0, 0

		n = 0
		for s := range synstation.SyncChannel {
			n++

			//fmt.Print(s," ")	
			switch {
			case s < -3:
				outD.BER1++
				fallthrough
			case s < -2:
				outD.BER2++
				fallthrough
			case s < -1:
				outD.BER3++
				fallthrough
			case s < -0.01:
				outD.connected++
			}
			if n >= synstation.M {
				break
			}

		}

		//geting a bit more data
		outD.d_connec = float(synstation.GetConnect())
		outD.d_discon = float(synstation.GetDisConnect())
		outD.Diversity = float(synstation.GetDiversity()) / float(synstation.M)
		outD.lost = float(synstation.SystemChan[0].GetAdded())
		outD.d_lost = float(synstation.GetLostConnect())
		outD.HopCount = float(synstation.GetHopCount())

		outDs.Add(&outD)

		if k%10 == 0 {
			outD.k = float(k)
			outChannel <- outD //sent data to print to  stdout

			//	t := synstation.CreateTrace(mobiles[0:synstation.M], synstations[0:synstation.D], k)
			//	draw.Draw(t)

		}

		//Run DBS Agent, and sync
		for i := range synstations {
			go synstations[i].RunAgent()
		}

		//sync
		n = 0
		for _ = range synstation.SyncChannel {
			n++
			if n >= synstation.D {
				break
			}
		}

		//synstation.PowerC(synstations[:])

	}

	// Print some status data

	outDs.Div(float(synstation.Duration))
	fmt.Println("Mean", outDs.String())

	for i := range synstations {
		fmt.Print(" ", synstations[i].Connec.Len())
	}
	fmt.Println()

	for i := range synstation.SystemChan {
		fmt.Print(" ", synstation.SystemChan[i].Emitters.Len())
	}
	fmt.Println()
	/*
		for i:= range mobiles{
			fmt.Print(" ", mobiles[i].Power);
		}
		fmt.Println()
	*/
	fmt.Println(" hops  ", synstation.Hopcount)

	/*fmt.Printf(" BER=[")
	for i:= range mobiles{
		fmt.Printf("  %f ", mobiles[i].BERtotal);
	}
	fmt.Println("];")*/

	SaveToFile(mobiles[:])

	/*fmt.Printf(" Div=[")
	for i:= range mobiles{
		fmt.Printf(" ", mobiles[i].Diversity);
	}
	fmt.Println("];")*/

	//outChannel.close()

	//And finaly close channels and background processes

	close(outChannel)
	for ch := range synstation.SystemChan {
		close(synstation.SystemChan[ch].Change)
	}
	close(synstation.SyncChannel)

	draw.Close()

}

func SaveToFile(mobiles []synstation.Mob) {

	outF, err := os.Open("out.m", os.O_WRONLY, 0666)

	fmt.Println(err)

	outF.WriteString(fmt.Sprintln("# name: Ptxr\n# type: matrix\n# rows: ", synstation.M, "\n# columns: ", 1))
	for i := 0; i < synstation.M; i++ {
		outF.WriteString(fmt.Sprintln(mobiles[i].Power, " "))

	}
	outF.WriteString("\n")

	outF.WriteString(fmt.Sprintln("# name: Pr\n# type: matrix\n# rows: ", synstation.M, "\n# columns: ", 1))
	for i := 0; i < synstation.M; i++ {
		if mobiles[i].GetMasterConnec() != nil {
			outF.WriteString(fmt.Sprintln(mobiles[i].GetMasterConnec().Pr, " "))
		} else {
			outF.WriteString(fmt.Sprintln(-1, " "))
		}

	}
	outF.WriteString("\n")

	outF.WriteString(fmt.Sprintln("# name: Div\n# type: matrix\n# rows: ", synstation.M, "\n# columns: ", 1))
	for i := 0; i < synstation.M; i++ {
		outF.WriteString(fmt.Sprintln(mobiles[i].Diversity, " "))

	}
	outF.WriteString("\n")

	outF.WriteString(fmt.Sprintln("# name: BERt\n# type: matrix\n# rows: ", synstation.M, "\n# columns: ", 1))
	for i := 0; i < synstation.M; i++ {
		outF.WriteString(fmt.Sprintln(mobiles[i].BERtotal, " "))

	}
	outF.WriteString("\n")

	outF.WriteString(fmt.Sprintln("# name: MaxSNR\n# type: matrix\n# rows: ", synstation.M, "\n# columns: ", 1))
	for i := 0; i < synstation.M; i++ {
		outF.WriteString(fmt.Sprintln(mobiles[i].SNRb, " "))

	}
	outF.WriteString("\n")

	outF.WriteString(fmt.Sprintln("# name: MaxBER\n# type: matrix\n# rows: ", synstation.M, "\n# columns: ", 1))
	for i := 0; i < synstation.M; i++ {
		outF.WriteString(fmt.Sprintln(mobiles[i].MaxBER, " "))

	}
	outF.WriteString("\n")

	outF.WriteString(fmt.Sprintln("# name: Ch\n# type: matrix\n# rows: ", synstation.M, "\n# columns: ", 1))
	for i := 0; i < synstation.M; i++ {
		outF.WriteString(fmt.Sprintln(mobiles[i].GetCh(), " "))

	}
	outF.WriteString("\n")

	outF.WriteString(fmt.Sprintln("# name: XX\n# type: matrix\n# rows: ", synstation.M, "\n# columns: ", 1))
	for i := 0; i < synstation.M; i++ {
		outF.WriteString(fmt.Sprintln(mobiles[i].Pos.X, " "))

	}
	outF.WriteString("\n")

	outF.WriteString(fmt.Sprintln("# name: YY\n# type: matrix\n# rows: ", synstation.M, "\n# columns: ", 1))
	for i := 0; i < synstation.M; i++ {
		outF.WriteString(fmt.Sprintln(mobiles[i].Pos.Y, " "))

	}
	outF.WriteString("\n")

	outF.Close()

}

