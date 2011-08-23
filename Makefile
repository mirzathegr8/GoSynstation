


include $(GOROOT)/src/Make.inc

DEPS=geom synstation gocairo  
TARG=simu
GOFILES=main.go save.go runtimeOutput.go


include $(GOROOT)/src/Make.cmd



