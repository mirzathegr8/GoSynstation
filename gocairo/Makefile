

include $(GOROOT)/src/Make.inc

TARG=gocairo
CGOFILES=draw.go
CGO_OFILES=\
        cairolib.o
CGO_LDFLAGS=-lcairo -lm
CLEANFILES+=cairolib

include $(GOROOT)/src/Make.pkg


cairolib.o:cairolib.c
	gcc $(_CGO_CFLAGS_$(GOARCH)) -fPIC -O2 -o $@ -c $(CGO_CFLAGS) $<


