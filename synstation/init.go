package synstation

import "math"
//import "fmt"
import "rand"
//import "container/list"

var SyncChannel chan float64


var Synstations [D]DBS
var Mobiles [M]Mob


var CoherenceFilter FilterInt

func init() {

	SyncChannel = make(chan float64, 100000)
	Rgen = rand.New(rand.NewSource(123813541954235))
	Rgen2 = rand.New(rand.NewSource(12384235))

	// create channels
	SystemChan = make([]*channel, NCh)
	for i := range SystemChan {
		c := new(channel)
		c.Init(i)
		SystemChan[i] = c
	}

	// evaluate overlaping factors of channels	

	overlapN := int(math.Floor(1.0 / (1.0 - float64(roverlap))))
	if overlapN < 1 {
		overlapN = 1
	}
	if roverlap > 0.0 {

		for i := NChRes; i < NCh; i++ {

			SystemChan[i].coIntC = make([]coIntChan, overlapN*2)

			for k := 1; k <= overlapN; k++ {

				fac := 1.0 - float64(k)*(1.0-roverlap)

				SystemChan[i].coIntC[overlapN-k].c = i - k
				SystemChan[i].coIntC[overlapN-k].factor = float64(fac)

				if i+k < NCh {
					SystemChan[i].coIntC[overlapN+k-1].c = i + k
					SystemChan[i].coIntC[overlapN+k-1].factor = float64(fac)
				}

			}

		}
	}

	/*d:=0
	nD := int(math.Sqrt(synstation.D))
	for i:=0;i< nD ;i++{
		for j:=0;j < nD;j++{
			x:=synstation.Field/float64(nD)*(float64(i)+ .5*float64(j%2) )
			y:=synstation.Field/float64(nD)*(float64(j)+.5)
			Synstations[d].R.SetPos(geom.Pos{x,y})
			d++
		}
	}*/

	for i := range Mobiles {
		Mobiles[i].Init(i)
	}

	// Here we define the Coherence bandwith as a ratio of the total bandwith (20MHz)
	corrF := float64(.2)
	// we use that to create a filter used to generates inputs to the doppler filters for each signals	
	A := Butter(corrF)
	B := Cheby(10, corrF)
	CoherenceFilter = MultFilter(A, B)

}


// what a random variable doing in constant declaration file???
// ah, ok the random generator is constant,... :p
var Rgen *rand.Rand  //on used for position
var Rgen2 *rand.Rand //one used to init shadow maps
//different rndvar are used to ensure repeatability of position with or without shadow maps


func Sync(k int) {
	for n := 0; n < k; n++ {
		<-SyncChannel
	}
}

func Init() {

	for i := range Synstations {
		go Synstations[i].Init()
	}
	//sync
	Sync(D)

}

