


include $(GOROOT)/src/Make.inc

DEPS=geom synstation gocairo draw 
TARG=simu
GOFILES=main.go save.go


include $(GOROOT)/src/Make.cmd



