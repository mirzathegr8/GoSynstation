

#include "cairo/cairo.h"


typedef struct{
	cairo_surface_t *surface;
	cairo_t *cr;
} MyCanvas;

void cairolibtest();

MyCanvas* create();

void freeC(MyCanvas *cr);

void clear(MyCanvas *cr);

void save(MyCanvas *c, char* ck, int k);

void drawConnection (MyCanvas *cr, float x1, float y1, float x2, float y2);

void drawCircle(MyCanvas *cr, float x1, float x2, float r);

void setColor(MyCanvas *cr, float r, float g, float b, float a);

void stroke(MyCanvas *cr);

void move_to(MyCanvas *cr,float x, float y);

void line_to(MyCanvas *cr,float x, float y);

void close_path(MyCanvas *cr);

void fill(MyCanvas *cr);

void fill_preserve(MyCanvas *cr);
