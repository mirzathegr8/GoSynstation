package synstation


const Field = 6000 //length in meters

const Duration = 1000 // in iterations 

const M = 800 //numbers of mobiles
const D = 128  // numbers of DBS

const L2 = 2 // modulation factor
const L1 = 1

const DivCh=1
const NCh = 100/DivCh + NChRes // number of channels
const EffectiveBW =  80  * DivCh

// Here we define the Coherence bandwith as a ratio of the total bandwith (20MHz)
const corrF = 0.2

// 10 0 11 .1 12 .2 19 .5 37 .75
const roverlap = 0.0 // ratio of overlaping of two adjacent channels

// thermal noise per RB 121.45dBm normalized per maximum terminal power output 21dBm and divided for one TTI
const WNoise = DivCh * 5.6885e-15 //7.1614e-16 // White noise at reciever //21.484e-16
const NChRes = 1              //numbers of reserved channels, not used yet, but chan 0 must be reserved
const NConnec = 25            // numbers of connections per dbs

const BERThres = 0.4
const SNRThresConnec = 15

const SNRThresChHop = 0

const MaxSpeed = 15

const EnodeBClock = 2



const NetLayout = RANDOM

const (
	HONEYCOMB = iota
	RANDOM
)

const SetReceiverType = BEAM

//type ReceiverType int

const (
	OMNI = iota
	BEAM
	SECTORED
)

const SetShadowMap = SHADOWMAP

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

const FastFading = MONTECARLO

const (
	MEANEVAL = iota
	MONTECARLO
)

const DiversityType = MRC

const (
	SELECTION = iota
	MRC
)

const PowerControl = NOPC

const (
	NOPC = iota
	AGENTPC
)


var ARBSchedulFunc = ARBScheduler3
//var ARBSchedulFunc = ARBScheduler
//var ARBSchedulFunc = ChHopping

var estimateFactor = estimateFactor0

const conservationFactor = 10 // 0.8 best for estimateFactor1 and ARBScheduler
//const conservationFactor = 10 // best for estimateFactor0 and ARBScheduler3


const popsize = 10
const generations = 10
const CAPAthres = 3000 // this value to define the relative min capacity compared to the maximum over ARB for one mobile, under this threshold RB will not be assigned

