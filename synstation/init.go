package synstation

import "math"
//import "fmt"
import rand "math/rand"
import "geom"
//import "container/list"


var Tti int //number of the current tti, to be incremented by the main loop

var SyncChannel chan float64


var Synstations [D]DBS
var Mobiles [M]Mob

var Agents [D + M]Agent

var CoherenceFilter FilterInt

func init() {

	SyncChannel = make(chan float64, 100000)
	Rgen = rand.New(rand.NewSource(123813541954235))
	Rgen2 = rand.New(rand.NewSource(12384235))

	for i := range Mobiles {
		Agents[i] = &Mobiles[i]
	}
	for i := range Synstations {
		Agents[i+M] = &Synstations[i]
	}

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

	
	for i := range Mobiles {
		Mobiles[i].Init(i)
	}

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

var IntereNodeBDist float64

func Init() {

	for i := range Synstations {
		go Synstations[i].Init()
	}

	//sync
	Sync(D)

	Dprim:=D
	if OneAgentPerBEAM {Dprim/=3}


	if NetLayout==HONEYCOMB{

		d:=0
		Wsd := math.Sqrt(float64(Dprim)*2*math.Sqrt(3) )	
		xD := int(Wsd/2.0)
		yD := int(Wsd/math.Sqrt(3))

		a:=(xD+1)*yD
		b:=(yD+1)*xD
		if a<=Dprim && b<=Dprim{
			if a>b { xD++} else {yD++}		
		} else if a<=Dprim{ xD++
		} else if b<=Dprim { yD++}
	

		DDx := Field / (float64(xD)+.8)
		DDy := DDx*math.Sqrt(3)/2

		IntereNodeBDist=DDx // set the interdistance for schedulers with antipasta

		deltaX:= (Field- (float64(xD)-0.5)*DDx) / 2.0
		deltaY:= (Field- (float64(yD)-1)*DDy) / 2.0
		for i:=0;i< xD ;i++{
			for j:=0;j < yD;j++{
				x:=deltaX + DDx*(float64(i)+ .5*float64(j%2) )
				y:=deltaY + DDy*float64(j)
				Synstations[d].R.SetPos(geom.Pos{x,y})
				Synstations[d].Color=( 2*(j%2) + i )%3
				d++
			}
		}
	}


	if OneAgentPerBEAM {
		for i:=0; i<Dprim;i++{		

			switch ICIMtype{
			case ICIMb:
			Color:=Synstations[i].Color	
			Synstations[i+Dprim].Color=Color		
			Synstations[i+2*Dprim].Color=Color
			case ICIMc:
			Synstations[i].Color=0
			Synstations[i+Dprim].Color=1
			Synstations[i+2*Dprim].Color=2

			}

			R:=Synstations[i].R.GetPos()
			Synstations[i+Dprim].R.(*PhysReceiver).SetPos(R)	
			Synstations[i+2*Dprim].R.(*PhysReceiver).SetPos(R)

			Synstations[i+Dprim].DBS2= &Synstations[i]
			Synstations[i+2*Dprim].DBS2= &Synstations[i]
			Synstations[i+2*Dprim].DBS3= &Synstations[i+Dprim]

		}
	
		if SetReceiverType==SECTORED2 {		
			for i:=0; i<Dprim;i++{
				R:=Synstations[i].R.(*PhysReceiver)			
				for i := 0; i < len(R.Orientation); i++ {
					R.Orientation[i] = 0
				}
				R=Synstations[i+Dprim].R.(*PhysReceiver)
				R.SetPos(Synstations[i].R.GetPos())
				for i := 0; i < len(R.Orientation); i++ {
					R.Orientation[i] = PI2/3.0
				}
				R = Synstations[i+2*Dprim].R.(*PhysReceiver)
				R.SetPos(Synstations[i].R.GetPos())
				for i := 0; i < len(R.Orientation); i++ {
					R.Orientation[i] = PI2*2.0/3.0
				}			
		}
	}
	}






	ChannelHop() //Set Mobile's initial channel

}

