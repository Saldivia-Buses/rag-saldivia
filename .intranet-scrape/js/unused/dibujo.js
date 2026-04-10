function draw_thumb() {
	var canvas = document.getElementById("canvas");
	var thumb = document.getElementById("thumb");
	var ctx = thumb.getContext("2d");
	ctx.drawImage(canvas, 0, 0, 100, 100);
}
function transform_event_coord(e) {
	return {x: e.clientX - 10, y: e.clientY - 10};
}
var drawing = false;
var lastpos = {x:-1,y:-1};

function on_mousedown(e) {
	drawing = true;
	lastpos = transform_event_coord(e);
}
function on_mousemove(e) {
	 if (!drawing)
		 return;

	var pos = transform_event_coord(e);

	var canvas = document.getElementById("canvas");
	var ctx = canvas.getContext("2d");

	ctx.strokeStyle = "rgba(200,0,0,0.6)";
	ctx.lineWidth = 4.0;
	//ctx.lineJoin = "round";
	ctx.beginPath();
	ctx.moveTo(lastpos.x, lastpos.y);
	ctx.lineTo(pos.x, pos.y);
	ctx.closePath();
	ctx.stroke();

	lastpos = pos;
}

function on_mouseup(e) {
	drawing = false;
	draw_thumb();
}
function init() {
	var ie = document.getElementById("ie");
	var canvas = document.getElementById("canvas");
	var ctx = canvas.getContext("2d");
	ctx.drawImage(ie, 120, 120);
	
	addEventListener("mousedown", on_mousedown, false);
	addEventListener("mousemove", on_mousemove, false);
	addEventListener("mouseup", on_mouseup, false);
	draw_thumb();
}