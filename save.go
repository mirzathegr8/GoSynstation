package main

import "synstation"
import "os"
import "fmt"

func init() {

	saveBER = make(chan *synstation.Trace, 10)
	saveBERMax = make(chan *synstation.Trace, 10)

	go SaveMobBER()
	go SaveMobBERMax()

}

var saveBER chan *synstation.Trace
var saveBERMax chan *synstation.Trace

func SaveMobBER() {

	outF, err := os.Open("BER.m", os.O_WRONLY, 0666)
	fmt.Println(err)
	outF.WriteString(fmt.Sprintln("# name: BERMat\n# type: matrix\n# rows: ", synstation.Duration, "\n# columns: ", synstation.M))

	for t := range saveBER {
		for i := 0; i < synstation.M; i++ {
			outF.WriteString(fmt.Sprint(t.Mobs[i].BERtotal, " "))

		}
		outF.WriteString("\n")

	}

	outF.Close()
}


func SaveMobBERMax() {

	outF, err := os.Open("BERMax.m", os.O_WRONLY, 0666)
	fmt.Println(err)
	outF.WriteString(fmt.Sprintln("# name: BERMaxMat\n# type: matrix\n# rows: ", synstation.Duration, "\n# columns: ", synstation.M))

	for t := range saveBERMax {

		for i := 0; i < synstation.M; i++ {
			outF.WriteString(fmt.Sprint(t.Mobs[i].MaxBER, " "))

		}
		outF.WriteString("\n")

	}

	outF.Close()
}

