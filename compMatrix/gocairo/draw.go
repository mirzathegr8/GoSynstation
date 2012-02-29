
package gocairo


// #include "cairolib.h"
import "C"

import "synstation"


type Canvas struct{
	Cv *C.MyCanvas

}

//var I sync.Mutex

func ToField(x float32) float32{ return x/synstation.Field *600.0}

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

func (Cv *Canvas) DrawLine(x1 float32, y1 float32,x2 float32,y2 float32 ){
	C.drawConnection(Cv.Cv, _Ctype_float(ToField(x1)), 
				_Ctype_float(ToField(y1)), 
				_Ctype_float(ToField(x2)),
				_Ctype_float(ToField(y2)) )
}


func (Cv *Canvas) DrawCircle(x float32, y float32,r float32){
	C.drawCircle(Cv.Cv, _Ctype_float(ToField(x)),
			_Ctype_float(ToField(y)),
			_Ctype_float(ToField(r))	)
}

func (Cv *Canvas) SetColor(r float32, g float32,b float32, a float32){
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

func (Cv *Canvas) MoveTo (x,y float32){
	C.move_to(Cv.Cv,_Ctype_float(ToField(x)),_Ctype_float(ToField(y)))
}

func (Cv *Canvas) LineTo (x,y float32){
	C.line_to(Cv.Cv,_Ctype_float(ToField(x)),_Ctype_float(ToField(y)))
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
