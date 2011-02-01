

#include "stdio.h"
#include "cairo/cairo.h"
#include <math.h>
#include <stdlib.h>
#include "cairolib.h"


inline float toField(float f){return  f/10.0;}

void cairolibtest(){

	printf("hello cairo lib");

        cairo_surface_t *surface =
            cairo_image_surface_create (CAIRO_FORMAT_ARGB32, 240, 80);
        cairo_t *cr =
            cairo_create (surface);

        cairo_select_font_face (cr, "serif", CAIRO_FONT_SLANT_NORMAL, CAIRO_FONT_WEIGHT_BOLD);
        cairo_set_font_size (cr, 32.0);
        cairo_set_source_rgb (cr, 0.0, 0.0, 1.0);
        cairo_move_to (cr, 10.0, 50.0);
        cairo_show_text (cr, "Hello, world");

        cairo_destroy (cr);
        cairo_surface_write_to_png (surface, "hello.png");
        cairo_surface_destroy (surface);
}

MyCanvas* create(){
	MyCanvas* cr= (MyCanvas *) malloc (sizeof(MyCanvas));
	cr->surface = cairo_image_surface_create (CAIRO_FORMAT_ARGB32, 600, 600);
	cr->cr = cairo_create (cr->surface);      

	//cairo_set_operator(cr->cr,CAIRO_OPERATOR_SCREEN);

	return cr;
}

void freeC(MyCanvas *cr){
	cairo_destroy(cr->cr);
	cairo_surface_destroy(cr->surface);
	free(cr);
}

void clear(MyCanvas *cr){
	cairo_fill(cr->cr);
	cairo_save(cr->cr); // save the state of the context
	cairo_set_source_rgb(cr->cr, 1, 1, 1);
	cairo_paint(cr->cr);    // fill image with the color
	cairo_restore(cr->cr);  // color is back to black now
	cairo_save(cr->cr); 
	cairo_set_source_rgba(cr->cr, 0.5, 0.5, 0.5, .5);
}

void save(MyCanvas *c, char* ck, int k){

	cairo_restore(c->cr);
	switch ((int)floor(log10((double)k)))
	{
		case 0:  sprintf(ck,"%s0000%d.png",ck,k);break;
		case 1:  sprintf(ck,"%s000%d.png",ck,k);break;
		case 2:  sprintf(ck,"%s00%d.png",ck,k);break;
		case 3:  sprintf(ck,"%s0%d.png",ck,k);break;
		default: sprintf(ck,"%s%d.png",ck,k);break;
	}
	cairo_surface_write_to_png(c->surface,ck);

}

void drawConnection (MyCanvas *cr, float x1, float y1, float x2, float y2){
		cairo_move_to (cr->cr, toField(x1), toField(y1));
		cairo_line_to (cr->cr, toField(x2), toField(y2));
	//	cairo_stroke (cr->cr);
}

void drawCircle(MyCanvas *cr, float x1, float x2, float r){

	cairo_arc(cr->cr,toField(x1),toField(x2),toField(r),0,2*M_PI);
//	cairo_stroke (cr->cr);

}

void setColor(MyCanvas *cr, float r, float g, float b, float a){
	cairo_set_source_rgba(cr->cr,r,g,b,a);
}

void stroke(MyCanvas *cr){
	cairo_stroke(cr->cr);
}

void move_to(MyCanvas *cr,float x, float y){
	cairo_move_to(cr->cr,x,y);
}

void line_to(MyCanvas *cr,float x, float y){
	cairo_line_to(cr->cr,x,y);
}

void close_path(MyCanvas *cr){
	cairo_close_path(cr->cr);
}


void fill(MyCanvas *cr){
	cairo_fill(cr->cr);
}


void fill_preserve(MyCanvas *cr){
	cairo_fill_preserve (cr->cr);
}

