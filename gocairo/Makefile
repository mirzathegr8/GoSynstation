

include $(GOROOT)/src/Make.inc

TARG=gocairo
CGOFILES=draw.go
CGO_OFILES=\
        cairolib.o
CGO_LDFLAGS=-lm -lcairo -L/lib64
CLEANFILES+=cairolib

include $(GOROOT)/src/Make.pkg


cairolib.o:cairolib.c
	gcc $(_CGO_CFLAGS_$(GOARCH))  -fPIC -O2 -o $@ -c $(CGO_CFLAGS) $<


