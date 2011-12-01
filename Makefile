


include $(GOROOT)/src/Make.inc

DEPS=geom synstation gocairo draw
TARG=simu
GOFILES=main.go save.go runtimeOutput.go


include $(GOROOT)/src/Make.cmd


cleanall :	
	$(MAKE) clean
	-for d in $(DEPS); do (cd $$d; $(MAKE) clean ); done
