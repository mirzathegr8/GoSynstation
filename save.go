package main

import "synstation"
import "os"
import "fmt"
import "container/vector"
import "bytes"
import "encoding/binary"
import "unsafe"

var syncsavech chan int

var saveData vector.Vector

func init() {

	syncsavech = make(chan int)

	saveData.Push(CreateStart(MaxBER, synstation.M, "BERMax"))
	saveData.Push(CreateStart(InstMaxBER, synstation.M, "InstMatBER"))
	saveData.Push(CreateStart(BER, synstation.M, "BER"))
	saveData.Push(CreateStart(SNR, synstation.M, "SNR"))
	saveData.Push(CreateStart(CH, synstation.M, "CH"))
	saveData.Push(CreateStart(DIV, synstation.M, "DIV"))
	saveData.Push(CreateStart(Outage, synstation.M, "Outage"))
	saveData.Push(CreateStart(Ptxr, synstation.M, "Ptxr"))
	saveData.Push(CreateStart(PrMaster, synstation.M, "PrMaster"))

}

type saveTraceItem struct {
	save chan *synstation.Trace
}

func CreateStart(method func(t *synstation.Trace, i int) float64, m int, file string) saveTraceItem {
	var a saveTraceItem
	a.save = make(chan *synstation.Trace, 1000)
	//go WriteDataToFile(method, m, a.save, file)
	go save_binary_data(method, m, a.save, file)
	return a
}

func (s *saveTraceItem) Stop() {
	close(s.save)
	<-syncsavech
}

func MaxBER(t *synstation.Trace, i int) float64     { return t.Mobs[i].MaxBER }
func InstMaxBER(t *synstation.Trace, i int) float64 { return t.Mobs[i].InstMaxBER }
func BER(t *synstation.Trace, i int) float64        { return t.Mobs[i].BERtotal }
func SNR(t *synstation.Trace, i int) float64        { return t.Mobs[i].SNRb }
func CH(t *synstation.Trace, i int) float64         { return float64(t.Mobs[i].ARB[0]) }
func DIV(t *synstation.Trace, i int) float64        { return float64(t.Mobs[i].Diversity) }
func Outage(t *synstation.Trace, i int) float64     { return float64(t.Mobs[i].Outage) }
func Ptxr(t *synstation.Trace, i int) float64       { return float64(t.Mobs[i].Power) }
func PrMaster(t *synstation.Trace, i int) float64   { return float64(t.Mobs[i].PrMaster) }

func WriteDataToFile(method func(t *synstation.Trace, i int) float64, m int, channel chan *synstation.Trace, file string) {

	outF, err := os.Open(file+".mat", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer outF.Close()

	_, err = outF.WriteString(fmt.Sprintln("# name: ", file, "\n# type: matrix\n# rows: ", synstation.Duration, "\n# columns: ", m))
	if err != nil {
		return // f.Close() will automatically be called now 
	}

	for t := range channel {

		buffer := bytes.NewBufferString("")
		for i := 0; i < m; i++ {
			fmt.Fprint(buffer, method(t, i))
			fmt.Fprint(buffer, " ")
		}
		_, err = outF.WriteString(string(buffer.Bytes()))
		if err != nil {
			return // f.Close() will automatically be called now 
		}
	}

	syncsavech <- 1

}

func StopSave() {
	for i := 0; i < len(saveData); i++ {
		a := saveData.At(i).(saveTraceItem)
		a.Stop()
	}
}

func sendTrace(t *synstation.Trace) {

	for i := 0; i < len(saveData); i++ {
		a := saveData.At(i).(saveTraceItem)
		a.save <- t
	}
}


// Header (one per file):
// =====================
//
//   object               type            bytes
//   ------               ----            -----
//   magic number         string             10
//
//   float format         integer             1
//
//
// Data (one set for each item):
// ============================
//
//   object               type            bytes
//   ------               ----            -----
//   name_length          integer             4
//
//   name                 string    name_length
//
//   doc_length           integer             4
//
//   doc                  string     doc_length
//
//   global flag          integer             1
//
//   data type            char                1
//
// In general "data type" is 255, and in that case the next arguments
// in the data set are
//
//   object               type            bytes
//   ------               ----            -----
//   type_length          integer             4
//
//   type                 string    type_length
//
// The string "type" is then used with octave_value_typeinfo::lookup_type
// to create an octave_value of the correct type. The specific load/save
// function is then called.
//
// For backward compatiablity "data type" can also be a value between 1
// and 7, where this defines a hardcoded octave_value of the type
//
//   data type                  octave_value
//   ---------                  ------------
//   1                          scalar
//   2                          matrix
//   3                          complex scalar
//   4                          complex matrix
//   5                          string   (old style storage)
//   6                          range
//   7                          string
//
// Except for "data type" equal 5 that requires special treatment, these
// old style "data type" value also cause the specific load/save functions
// to be called. FILENAME is used for error messages.
func save_binary_data(method func(t *synstation.Trace, i int) float64, m int, channel chan *synstation.Trace, file string) {

	os, err := os.Open(file+".mat", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer os.Close()

	var bb []byte
	bb = make([]byte, 4)

	os.WriteString("Octave-1-L")
	var float_type [1]byte
	float_type[0] = 0
	os.Write(float_type[:])

	binary.LittleEndian.PutUint32(bb[:], uint32(len(file)))
	os.Write(bb)

	os.WriteString(file)

	block := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0x06, 0x00, 0x00, 0x00, 0x6d, 0x61, 0x74, 0x72, 0x69, 0x78, 0xfe, 0xff, 0xff, 0xff}

	os.Write(block)

	binary.LittleEndian.PutUint32(bb, uint32(m))
	os.Write(bb)

	binary.LittleEndian.PutUint32(bb, uint32(synstation.Duration))
	os.Write(bb)

	bb[0] = 6
	bb[1] = 0
	os.Write(bb[0:1])

	myfloat := float32(1)

	var bb2 [4]byte

	for t := range channel {

		buffer := bytes.NewBufferString("")

		for i := 0; i < m; i++ {
			myfloat = float32(method(t, i))
			bb2 = *(*[4]byte)(unsafe.Pointer(&myfloat))
			buffer.Write(bb2[:])
		}

		os.Write(buffer.Bytes())

	}

	os.Close()

	syncsavech <- 1

}

