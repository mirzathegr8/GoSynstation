package main


import "fmt"
import s "synstation"
import "runtime"
//import "draw"
import "os"


// Data to save for output during simulation


var outD outputData  // one to print
var outDs outputData // one to sum and take mean over simulation

func main() {

	runtime.GOMAXPROCS(10)

	s.Init()

	//draw.Init(s.M, s.D*s.NConnec) // init drawing system
	//draw.DrawReceptionField(s.Synstations[:], "receptionLevel.png")

	go printData() //launch thread to print output

	fmt.Println("Init done")
	fmt.Println("Start Simulation")

	// precondition
	for s.Tti = -200; s.Tti < 0; s.Tti++ {
		/*fmt.Print("---- 1---- ", s.Mobiles[0].Diversity, " ")
		fmt.Print( &s.Mobiles[0].MasterConnection)
		fmt.Println()
		fmt.Println("ARB ",s.Mobiles[0].ARB)
		fmt.Println("fut ",s.Mobiles[0].ARBfutur)*/
		s.GoRunPhys()
		/*fmt.Print("---- 2---- ", s.Mobiles[0].Diversity, " ")
		fmt.Print( &s.Mobiles[0].MasterConnection)
		fmt.Println()
		fmt.Println("ARB ",s.Mobiles[0].ARB)
		fmt.Println("fut ",s.Mobiles[0].ARBfutur)*/

		s.GoFetchData()
	/*	fmt.Print("---- 3---- ", s.Mobiles[0].Diversity, " ")
		fmt.Print( &s.Mobiles[0].MasterConnection)
		fmt.Println()
		fmt.Println("ARB ",s.Mobiles[0].ARB)
		fmt.Println("fut ",s.Mobiles[0].ARBfutur)*/

		readDataAndPrintToStd(false)

		s.GoRunAgent()


	/*	fmt.Print("---- 4---- ", s.Mobiles[0].Diversity, " ")
		fmt.Print( &s.Mobiles[0].MasterConnection)
		fmt.Println()
		fmt.Println("ARB ",s.Mobiles[0].ARB)
		fmt.Println("fut ",s.Mobiles[0].ARBfutur)*/

		s.ChannelHop()
		/*fmt.Print("---- 5---- ", s.Mobiles[0].Diversity, " ")
		fmt.Print( &s.Mobiles[0].MasterConnection)
		fmt.Println()
		fmt.Println("ARB ",s.Mobiles[0].ARB)
		fmt.Println("fut ",s.Mobiles[0].ARBfutur)*/


		//s.PowerC(s.Synstations[:]) //centralized PowerAllocation
	}

	// simu
	for s.Tti = 0; s.Tti < s.Duration; s.Tti++ {
		s.GoRunPhys()
		s.GoFetchData()
		readDataAndPrintToStd(true)
		s.GoRunAgent()
		s.ChannelHop()
		//s.PowerC(s.Synstations[:]) // centralized PowerAllocation
	}

	// Print some status data
	outDs.Div(float64(s.Duration))
	fmt.Println("Mean", outDs.String())
	
	os.Remove("Mean.mat")
	outF, err := os.OpenFile("Mean.mat", os.O_WRONLY|os.O_CREATE, 0666)

	fmt.Println(err)

	outF.WriteString(fmt.Sprintln("# name: Mean\n# type: matrix\n# rows: ", 1, "\n# columns: ", 11))
	outF.WriteString(fmt.Sprintln(outDs.String(), " "))	
	outF.WriteString("\n")
	outF.Close()

	for i := range s.Synstations {
		fmt.Print(" ", s.Synstations[i].Connec.Len())
	}
	fmt.Println()
	for i := range s.SystemChan {
		fmt.Print(" ", s.SystemChan[i].Emitters.Len())
	}
	fmt.Println()

	SaveToFile(s.Mobiles[:],s.Synstations[:])

	//And finaly close channels and background processes

	close(outChannel)
	close(s.SyncChannel)

	StopSave()

	//draw.Close()

}


func readDataAndPrintToStd(save bool) {

	outD.listened, outD.connected, outD.BER1, outD.BER2, outD.BER3 = 0, 0, 0, 0, 0

	n := 0
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
		case v < -0.00001:
			outD.connected++
			fallthrough
		case v != 0:
			outD.listened++
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

	outD.k = float64(s.Tti)
	if s.Tti%10 == 0 {
		outChannel <- outD //sent data to print to  stdout			
	}

	if s.Tti%200 == 0 {
		runtime.GC()
	}

	if save {
		outDs.Add(&outD)
		t := s.CreateTrace(s.Mobiles[:], s.Synstations[:], s.Tti)
		//draw.Draw(t)
		sendTrace(t)
	}

}

