

package geom

import "math"

type Pos struct {
	X float64
	Y float64
}

func (p *Pos) GetX() float {
	return float(p.X)
}

func (p *Pos) GetY() float {
	return float(p.Y)
}

func (p *Pos) Distance(p2 *Pos) float64 {
	a := p.X - p2.X
	b := p.Y - p2.Y
	return math.Sqrt(a*a + b*b)
}

func (p Pos) Plus(p2 Pos) Pos{
	return Pos{p.X+p2.X, p.Y+p2.Y}
}

func (p Pos) Minus(p2 Pos) Pos{
	return Pos{p.X-p2.X, p.Y-p2.Y}
}

func (p Pos) Times(d float64) Pos{
	return Pos{p.X*d, p.Y*d}
}

func (p Pos) Scalar(p2 Pos) float64{
	return p.X*p2.X + p.Y*p2.Y
}


func (p Pos) Normalise() Pos{
	return p.Times(p.Len())
}


func (p Pos) Prodv(p2 Pos) float64{
	return p.X*p2.Y - p.Y*p2.X
}


func (p Pos) IsNull() bool{
	if p.X==0 && p.Y==0 {return true}
	return false
}

type GeomLine struct{
	Org Pos
	Vect Pos
}

func (l *GeomLine) Intersect( l2 *GeomLine) Pos{
	x1:=l.Org.X-10.0*l.Vect.X
	y1:=l.Org.Y-10.0*l.Vect.Y
	x2:=x1+10.0*l.Vect.X
	y2:=y1+10.0*l.Vect.Y
	x3:=l2.Org.X-10.0*l2.Vect.X
	y3:=l2.Org.Y-10.0*l2.Vect.Y
	x4:=x3+10.0*l2.Vect.X
	y4:=y3+10.0*l2.Vect.Y

//	det:= 20*l.Vect.Det(l2.Vect);
	det:= (x1-x2)*(y3-y4) - (y1-y2)*(x3-x4)

	var x,y float64
	if (Abs(det)>1e-9){
		x=(x1*y2-y1*x2)*(x3-x4) - (x1-x2)*(x3*y4-y3*x4)
		x/=det
		y=(x1*y2-y1*x2)*(y3-y4) - (y1-y2)*(x3*y4-y3*x4)
		y/=det
	}
	return Pos{x,y}
}

func (p Pos) Rot90() Pos{
	return Pos{-p.Y,p.X}
}

func (p Pos) DetV(p2 Pos) Pos{
	return Pos{p.X*p2.Y, p.Y*p2.X}
}

func (p Pos) Len() float64{
	return math.Sqrt(p.X*p.X + p.Y*p.Y)
}

func (p Pos) Det(p2 Pos) float64{
	return p.X*p2.Y - p.Y*p2.X
}

func (p Pos) ToFloat() (x ,y float){
	return float(p.X),float(p.Y)
}


func Abs(f float64) float64 {
	if f < 0.0 {
		return -f
	}
	return f
}

func Sign(f float64) float64 {
	switch {
	case f < 0:
		return -1.0
	case f > 0:
		return 1.0
	}
	return 0.0

}

