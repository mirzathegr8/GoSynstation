

CCGO=gccgo
CFLAGS = -o4 -mtune=corei7 

TARG=gocairo
GOFILES = draw.go

gocairo.o: cairolib.o
		$(CCGO) -c $(GOFILES) -o $(TARG).o
	
cairolib.o:cairolib.c
	gcc $(_CGO_CFLAGS_$(GOARCH)) -fPIC -O2 -o $@ -c $(CGO_CFLAGS) $<


