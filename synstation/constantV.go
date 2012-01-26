package synstation
const BERThres= 0.15
const SNRThresConnec= 0
var initScheduler=initChHopping2  //ARBScheduler4//SDMA // 
//var estimateFactor=estimateFactor0
const conservationFactor=1
const DiversityType=MRC
const BeamAngle=    1.1345//0.4254//
const mDf=1
const NetLayout= HONEYCOMB
const subsetSize= 3
const InterferenceCancel= SIZEESCANCELATION //NOCANCEL //
const SetReceiverType = SECTORED

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

// number of signal id saved in the list of the ChanReceiver
const SizeES = 1
