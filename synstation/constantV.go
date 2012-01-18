package synstation
const BERThres= 0.15
const SNRThresConnec= 0
var initScheduler=initChHopping2
var estimateFactor=estimateFactor1
const conservationFactor=1.0
const DiversityType=MRC
const BeamAngle=0.4254
const mDf=10
const NetLayout= HONEYCOMB
const subsetSize= 15
const InterferenceCancel=SIZEESCANCELATION


//types of eNode interference cancelation
// either none, or only master signals are canceled from others received signals
// or all mobiles listened too are canceled from every listened signals
// or SizeES best received signals on each RB are canceld out
const (
	NOCANCEL = iota
	MASTERCANCELATION
	CONNECTEDCANCELATION
	SIZEESCANCELATION
)
