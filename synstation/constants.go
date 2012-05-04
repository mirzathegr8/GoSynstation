package synstation

import "math"

const Field = 1500 //length in meters

const Duration =500// in iterations 

const M = 500 //numbers of mobiles
const D =  50// numbers of DBS

//for M-QAM, km*km=M
const km = 4.0
const L2 = 2.0/3.0/km*(km*km-1.0) // modulation factor
const L1 = 2.0/km*(1.0-1.0/km)

const beta=1 // SNR Gap

const NRB=100
const NTDMA=1
const DivCh=1
const NCh = NRB*NTDMA/DivCh + NChRes // number of channels
const EffectiveBW =  90  * DivCh

// Here we define the Coherence bandwith as a ratio of the total bandwith (20MHz)
const corrF = 0.3

// 10 0 11 .1 12 .2 19 .5 37 .75
const roverlap = 0.0 // ratio of overlaping of two adjacent channels

// thermal noise per RB 121.45dBm normalized per maximum terminal power output 21dBm and divided for one TTI
const WNoise = DivCh * 5.6885e-15 //7.1614e-16 // White noise at reciever //21.484e-16
const NChRes = 1              //numbers of reserved channels, not used yet, but chan 0 must be reserved
const NConnec = 25        // numbers of connections per dbs


//const BERThres = 0.30 //0.16//0.4/log2*(16)
//const SNRThresConnec = 15
const SNRThresChHop = 0

// maximum speed of mobiles in m/s, speeds are drawn from a uniform [0, MaxSpeed] interval for each mobiles
// at the bigining of the simulation
const MaxSpeed = 15


// Defines the layout, either a treillis of honeycomb placed eNodes
// or randomly set positions
//const NetLayout = RANDOM
const (
	HONEYCOMB = iota
	RANDOM
)

//const BeamAngle = 1.1345 // for 120degre lobe (half lob size + 10% =65deg in rad)


// 
const (
	OMNI = iota
	BEAM
	SECTORED
	SECTORED2
)


// Shadow map can be set or not, direct evaluation can also be used by modifying one line in physReceiver.go
// direct evaluation generates is slower as it computes the shading value each time.
const SetShadowMap = NOSHADOW
const (
	NOSHADOW = iota
	SHADOWMAP
)

//// Shadow map parameters
//	variance used for the generation of the coeficients 
const shadow_deviance = 10
//	pixel frequency
const corr_res = 50
//	sampling of the  field : where the freq is defined in number of cycles/Field length
//					1Hz= 1Cylce in Field lenght
//				only Field/shadow_sampling will be used to generate the map 
//				Hence, a corr_res imples selectin frequencies below 
//				 Field/corr_res cycles
const shadow_sampling = 15
//  this is threshold to cancal frequencies which have too low power, which amplitude is a fraction compared to the mean amplitude
const mval = 0.1
// These two values sets the size of the generated map image for non direct evaluation (precalculation of map)
// and the length that the map covers
const maplength = 1500 // in m
const mapsize = 600 // in pixels


//const DiversityType = SELECTION
const (
	SELECTION = iota
	MRC
)

//Sets the power allocation methode
//var PowerAllocation = optimizePowerAllocationAgent
//var PowerAllocation = optimizePowerNone
//var PowerAllocation = optimizePowerAllocationSimple

//Sets the Scheduling methode
//var ARBSchedulFunc = ARBScheduler3
//var ARBSchedulFunc = ARBScheduler
//var ARBSchedulFunc = ChHopping
//var ARBSchedulFunc = ChHopping2



//var estimateFactor = estimateFactor1

//const conservationFactor = 0.8 // 0.8 best for estimateFactor1 and ARBScheduler
//const conservationFactor = 10 // best for estimateFactor0 and ARBScheduler3

// These parameter or set for the Genetic search SC-FDMA algorithms 
// functions ARBScheduler 3 4 and 4
const popsize = 10
const generations = 10
const CAPAthres = 3000 // this value to define the relative min capacity compared to the maximum over ARB for one mobile, under this threshold RB will not be assigned


// Function to evaluate the Noise, where Pint, is the total interference plus signal
// Pr is the Signal power (if it is already summed in Pint otherwise 0)
// This function allows to observe the effect of no interference, or mean interference power.
func GetNoisePInterference(Pint,Pr float64) float64{
	return Pint-Pr + WNoise
	//return 1.00e-14
	//return WNoise*5000
}


// Defines how interference is evaluated
//	Normal implies Signal* fastfading / (all interferers plus noise)
//	Fading implies Signal* fastfading / (all interferers with fading on strongest interferer plus noise)
//	Cancel implies Signal*fastfading / (all interferers except strongest interferer plus noise)
const FadingOnPint1 = Normal
const (
	Normal=iota
	Fading
	Cancel
)

// Defines which type of icim is used : 
//	b) 3 colors per 3 eNodeBs
//	c) 3 colors per 3 sectors of one eNodeB
const (
	ICIMb=iota
	ICIMc
)

// we can use three different ways to evaluate the capacity
// the direct normal one, will be
//	Sum_{RBs} effectiveBW * log2(1 + Sum_{b in eNodeBs} SINR_MRC(b, rb) )  where SINR(b,rb) the signal/(interference+noise) of the RB at TTI t for enodeb b
// the EffectiveSINR methode for OFDM evaluate first an equivalent SINR over all sinr of all rbs
// the capacity is
//	effectiveBW * Card({RBs}) * log2 (1 + SINR_eff)
// where SINR_eff is (the sum for each enodeb for MRC)  of - beta* ln( 1/Card({RBs}) Sum_{RBs} exp(- SINR(rb)/beta))
// Finaly the MCSJoint method takes the potential spectral efficiency, rounds (floor) it at the nearest modulation/coding scheme and uses this MCS for the next TTI, if the SINR is too low for it, the paquet is considered lost. SINR_eff is used in this case
const (
	NormalTR= iota
	EFFECTIVESINR
	MCSJOINT 
)




const PI2 = 2 * math.Pi
const PI = math.Pi



//const TransferRateTechnique=  EFFECTIVESINR//  NormalTR  //EFFECTIVESINR//, MCSJOINT //NormalTR   , MCSJOINT

const ICIMtype= ICIMb

// ICIM Theta is for the moment deprecated,
const ICIMTheta=false
// ICIMdistRatio sets the ratio of distance (emiter-enode/neighboring enode) for reusing all RB in the cell's center
const ICIMdistRatio=0.3


// These two parameters allows arranging eNodeBs to colocate 3 into one and allow RB reuse
// a hack is being used to prevent the 3 colocated enodes to connect more than once to one emitter
// this is a work in progress and the architecture needs to be rethought TODO
//const OneAgentPerBEAM =false // controls for beamforming or sectorisation the ability to reuse RB accross different beam






//var ICIMfunc=ICIMSplitEdgeCenter2
var ICIMfunc=ICIMNone
var PowerAllocation =  optimizePowerAllocationAgent // optimizePowerAllocationAgentRB//  optimizePowerAllocationAgent // optimizePowerNone // 
const PowerAgentFact = 0.8//0.8//0.2 
const PowerAgentAlpha = 1 //0.8//0.2 


const uARBcost = 0.00000 //meanMeanCapa / 5 //0.5 // math.Log2(1 + meanMeanCapa)

const TRATETECH = NORMAL
const (
	OFDM = iota
	OFDM2	
	SCFDM
	NORMAL
)

