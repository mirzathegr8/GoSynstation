


CCGO=gccgo
CFLAGS = -o4 -mtune=corei7 

DEPS=geom synstation gocairo draw
TARG=simu
GOFILES=main.go


main.o : 
	$(CCGO) -c $(GOFILES)  -o $(TARG).o






