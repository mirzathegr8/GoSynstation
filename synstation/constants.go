package synstation


const Field = 6000 //length in meters

const Duration = 1000// in iterations 

const M = 1000 //numbers of mobiles
const D = 100  // numbers of DBS

//for M-QAM, km*km=M
const km = 4.0
const L2 = 2.0/3.0/km*(km*km-1.0) // modulation factor
const L1 = 2.0/km*(1.0-1.0/km)

//const BERThres = 0.30 //0.16//0.4/log2*(16)

//const L1=1
//const L2=2
const beta=1 // SNR Gap


const DivCh=1
const NCh = 100/DivCh + NChRes // number of channels
const EffectiveBW =  90  * DivCh

// Here we define the Coherence bandwith as a ratio of the total bandwith (20MHz)
const corrF = 0.2

// 10 0 11 .1 12 .2 19 .5 37 .75
const roverlap = 0.0 // ratio of overlaping of two adjacent channels

// thermal noise per RB 121.45dBm normalized per maximum terminal power output 21dBm and divided for one TTI
const WNoise = DivCh * 5.6885e-15 //7.1614e-16 // White noise at reciever //21.484e-16
const NChRes = 1              //numbers of reserved channels, not used yet, but chan 0 must be reserved
const NConnec = 25           // numbers of connections per dbs


//const SNRThresConnec = 15

const SNRThresChHop = 0

const MaxSpeed = 15

const EnodeBClock = 4



//const NetLayout = RANDOM

const (
	HONEYCOMB = iota
	RANDOM
)

//const BeamAngle = 1.1345 // for 120degre lobe ==PI/2 (half lob size)
const BeamAngle = 1.1345

//const SetReceiverType = BEAM

//type ReceiverType int

const (
	OMNI = iota
	BEAM
	SECTORED
)

const SetShadowMap = NOSHADOW

//type ReceiverType int

const (
	NOSHADOW = iota
	SHADOWMAP
)


const shadow_deviance = 10
const corr_res = 50
const shadow_sampling = 15
const mval = 0.1
const maplength = 1500
const mapsize = 600


//const DiversityType = SELECTION

const (
	SELECTION = iota
	MRC
)

const PowerControl = NOPC

const (
	NOPC = iota
	AGENTPC
)


//var ARBSchedulFunc = ARBScheduler3
//var ARBSchedulFunc = ARBScheduler
//var ARBSchedulFunc = ChHopping
//var ARBSchedulFunc = ChHopping2
var subsetSize=5

//var estimateFactor = estimateFactor1

//const conservationFactor = 0.8 // 0.8 best for estimateFactor1 and ARBScheduler
//const conservationFactor = 10 // best for estimateFactor0 and ARBScheduler3


const popsize = 10
const generations = 10
const CAPAthres = 3000 // this value to define the relative min capacity compared to the maximum over ARB for one mobile, under this threshold RB will not be assigned


func GetNoisePInterference(Pint,Pr float64) float64{
	return Pint-Pr + WNoise
	//return WNoise*5000
}

const FadingOnPint1 = Normal
const (
	Normal=iota
	Fading
	Cancel
)

