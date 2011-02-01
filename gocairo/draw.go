

package gocairo



// #include "cairolib.h"
import "C"



type Canvas struct{
	Cv *C.MyCanvas

}

//var I sync.Mutex

func Cairotest(){
	C.cairolibtest()
}

func (Cv *Canvas) Create(){
	//I.Lock()
	Cv.Cv = C.create()
	C.clear(Cv.Cv)
	//I.Unlock()
}

func (Cv *Canvas) Clear() {
	//I.Lock()
	C.clear(Cv.Cv)
	//I.Unlock()
}

func (Cv *Canvas) DrawLine(x1 float, y1 float,x2 float,y2 float ){
	C.drawConnection(Cv.Cv, _Ctype_float(x1), 
				_Ctype_float(y1), 
				_Ctype_float(x2),
				_Ctype_float(y2) )
}


func (Cv *Canvas) DrawCircle(x float, y float,r float){
	C.drawCircle(Cv.Cv, _Ctype_float(x),
			_Ctype_float(y),
			_Ctype_float(r)	)
}

func (Cv *Canvas) SetColor(r float, g float,b float, a float){
	C.setColor(Cv.Cv, _Ctype_float(r),_Ctype_float(g),_Ctype_float(b), _Ctype_float(a))
}
	
func (Cv *Canvas)Save( name string, k int){
	//I.Lock()
	C.save(Cv.Cv, C.CString(name), C.int(k))
	//I.Unlock()
	
}

func (Cv *Canvas) Stroke (){
	//I.Lock()
	C.stroke(Cv.Cv)
	//I.Unlock()
}

func (Cv *Canvas) MoveTo (x,y float){
	C.move_to(Cv.Cv,_Ctype_float(x),_Ctype_float(y))
}

func (Cv *Canvas) LineTo (x,y float){
	C.line_to(Cv.Cv,_Ctype_float(x),_Ctype_float(y))
}

func (Cv *Canvas) ClosePath (){
	C.close_path(Cv.Cv)
}

func (Cv *Canvas) Fill(){
	//I.Lock()
	C.fill(Cv.Cv)
	//I.Unlock()
}

func (Cv *Canvas) FillPreserve(){
	//I.Lock()
	C.fill_preserve(Cv.Cv)
	//I.Unlock()
}


	
func (Cv *Canvas)Close(){	
	C.freeC(Cv.Cv)
}
