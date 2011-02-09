
CCGO=gccgo
CFLAGS = -o4 -mtune=corei7 

TARG=synstation
GOFILES=constants.go\
init.go\
emitter.go\
channels.go\
connection.go\
mobile.go\
synstation.go\
physReceiverSectored.go\
powerC.go\
trace.go\
physReceiver.go


synstation.o : 
	$(CCGO) -c $(GOFILES)  -o $(TARG).o




