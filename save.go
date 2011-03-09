package main

import "synstation"
import "os"
import "fmt"
import "container/vector"

var syncsavech chan int

var saveData vector.Vector

func init() {

	syncsavech= make(chan int)

	//saveData=new(vector.Vector)//make([]saveTraceItem,5)

	saveData.Push(CreateStart(MaxBER, synstation.M, "BERMax"))
	saveData.Push(CreateStart(InstMaxBER, synstation.M, "InstMatBER"))
	saveData.Push(CreateStart(BER, synstation.M, "BER"))
	saveData.Push(CreateStart(SNR, synstation.M, "SNR"))
	saveData.Push(CreateStart(CH, synstation.M, "CH"))
	saveData.Push(CreateStart(DIV, synstation.M, "DIV"))
	
}

type saveTraceItem struct{
	save chan *synstation.Trace
}

func CreateStart(method func (t *synstation.Trace, i int) float64 , m int , file string ) saveTraceItem{
	var a saveTraceItem
	a.save = make(chan *synstation.Trace, 1000)
	go WriteDataToFile(method, m, a.save, file)
	return a
}

func (s *saveTraceItem) Stop(){
	close(s.save)
	<- syncsavech
}

func MaxBER( t *synstation.Trace,  i int) float64 { return t.Mobs[i].MaxBER}  
func InstMaxBER( t *synstation.Trace,  i int) float64 { return t.Mobs[i].InstMaxBER} 
func BER( t *synstation.Trace,  i int) float64 { return t.Mobs[i].BERtotal} 
func SNR( t *synstation.Trace,  i int) float64 { return t.Mobs[i].SNRb} 
func CH( t *synstation.Trace,  i int) float64 { return float64(t.Mobs[i].Ch)} 
func DIV( t *synstation.Trace,  i int) float64 { return float64(t.Mobs[i].Diversity)} 

func WriteDataToFile( method func (t *synstation.Trace, i int) float64 , m int, channel chan *synstation.Trace, file string )  { 

	outF, err := os.Open(file+".mat", os.O_WRONLY, 0666)
	fmt.Println(err)
	outF.WriteString(fmt.Sprintln("# name: ",file,"\n# type: matrix\n# rows: ", synstation.Duration, "\n# columns: ", m))

//	dataM:= make([]float64,m)

	for t := range channel {
		var s string
		for i := 0; i < m; i++ {	
		s += fmt.Sprintf("%1.2f ",method(t,i))
		}

		outF.WriteString(s +"\n")

	}

	
	

	outF.Close()

	syncsavech <-1

} 

func StopSave(){
	for i:= 0 ; i< len(saveData); i++ {
		a:=saveData.At(i).(saveTraceItem)
		a.Stop()
	}
}

func sendTrace(t *synstation.Trace){

	for i:= 0 ; i< len(saveData); i++ {
		a:=saveData.At(i).(saveTraceItem)
		a.save <- t
	}
}

