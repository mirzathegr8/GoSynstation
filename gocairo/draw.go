package gocairo

type MyCanvas struct { }
func create() *MyCanvas __asm__ ("create");
func clear(cv *MyCanvas)  __asm__ ("clear");
func drawConnection(cv *MyCanvas, x1 float, y1 float,x2 float,y2 float)  __asm__ ("drawConnection");
func drawCircle(cv *MyCanvas,x float,y float,r float)  __asm__ ("drawCircle");
func setColor(cv *MyCanvas,r float,g float,b float,a float)  __asm__ ("setColor");
func save(cv *MyCanvas, name *byte, k int )  __asm__ ("save");
func stroke(cv *MyCanvas)  __asm__ ("stroke");
func move_to(cv *MyCanvas, x float, y float)  __asm__ ("move_to");
func line_to(cv *MyCanvas, x float, y float)  __asm__ ("line_to");
func close_path(cv *MyCanvas)  __asm__ ("close_path");
func fill(cv *MyCanvas)  __asm__ ("fill");
func fill_preserve(cv *MyCanvas)  __asm__ ("fill_preserve");
func freeC(cv *MyCanvas)  __asm__ ("freeC");

type Canvas struct {
	Cv *MyCanvas
}


func (Cv *Canvas) Create() {
	Cv.Cv = create()
	clear(Cv.Cv)
}

func (Cv *Canvas) Clear() {
	clear(Cv.Cv)
}

func (Cv *Canvas) DrawLine(x1 float, y1 float, x2 float, y2 float) {
	drawConnection(Cv.Cv, x1, y1, x2, y2)
}


func (Cv *Canvas) DrawCircle(x float, y float, r float) {
	drawCircle(Cv.Cv, x, y, r)
}

func (Cv *Canvas) SetColor(r float, g float, b float, a float) {
	setColor(Cv.Cv, r, g, b, a)
}

func (Cv *Canvas) Save(nkame string, k int) {
var name = [4]byte{'f', 'o', 'o', 0};
	save(Cv.Cv, &name[0], k)
}

func (Cv *Canvas) Stroke() {
	stroke(Cv.Cv)
}

func (Cv *Canvas) MoveTo(x, y float) {
	move_to(Cv.Cv, x, y)
}

func (Cv *Canvas) LineTo(x, y float) {
	line_to(Cv.Cv, x, y)
}

func (Cv *Canvas) ClosePath() {
	close_path(Cv.Cv)
}

func (Cv *Canvas) Fill() {
	fill(Cv.Cv)
}

func (Cv *Canvas) FillPreserve() {
	fill_preserve(Cv.Cv)
}


func (Cv *Canvas) Close() {
	freeC(Cv.Cv)
}

