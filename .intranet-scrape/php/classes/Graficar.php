<?php

/*
 * Created on 16/09/2007
 *
 * To change the template for this generated file go to
 * Window - Preferences - PHPeclipse - PHP - Code Templates
 */
require_once("../lib/barcode/barcode.inc.php");

class Graficar {

    var $clase;
    var $ancho;
    var $alto;
    var $archivo;
    var $tipo;
    var $imagen;

    public function __construct($clase='', $ancho=100, $alto=100, $file=null, $tipo= 'png', $trans=true) {
        $this->clase = $clase;
        $this->ancho = $ancho;
        $this->alto = $alto;
        $this->archivo = $file;
        $this->tipo = $tipo;
        $this->trans = $trans;
    }

    function crearImagen() {

        $this->imagen = imagecreatetruecolor($this->ancho, $this->alto);
        $trans_colour = imagecolorallocate($this->imagen, 255, 255, 255);

        // pg transparente
        if ($this->tipo == 'png' && $this->trans == true) {
            imagesavealpha($this->imagen, true);
            $trans_colour = imagecolorallocatealpha($this->imagen, 0, 0, 0, 127);
            //  imagefill($this->imagen, 0, 0, $trans_colour);
        }
        imagefill($this->imagen, 0, 0, $trans_colour);
    }

    function barcode($valor, $height=25, $scale=1.5, $imagen=null, $encode='I25') {

    	if ($encode == 'QR') {
    		require_once('../lib/phpqrcode/qrlib.php');
    		$return = QRcode::jpg($valor, $imagen.'.jpg', $scale, $height);    	
//    		print_r($return);	
    	}
    	else {
    	
	        $bar = new BARCODE();
	
	        if ($bar == false)
	            die($bar->error());
	        // OR $bar= new BARCODE("I2O5");
	
	        $barnumber = $valor;
	
	        $bar->setSymblogy($encode);
	        $bar->setHeight($height);
	
	        $bar->setFont("../lib/barcode/FreeSans.ttf");
	        $bar->setScale($scale);
	        $bar->setHexColor("#000000", "#FFFFFF");
	
	        $return = $bar->genBarCode($barnumber, 'jpg', $imagen);

    	}
        
        return $return;
    }

    function gradient($orientation, $size, $color1, $color2) {

        if ($color1 == null)
            $color1 = array('red' => 255, 'green' => 255, 'blue' => 255);   //white
        if ($color2 == null)
            $color2 = array('red' => 0, 'green' => 0, 'blue' => 0);         //black
            // Incremento en los canales
        $incr_r = ($color2['red'] - $color1['red']) / $size;
        $incr_g = ($color2['green'] - $color1['green']) / $size;
        $incr_b = ($color2['blue'] - $color1['blue']) / $size;

        // Valores correspondientes al primer color
        $r = $color1['red'];
        $g = $color1['green'];
        $b = $color1['blue'];


        for ($i = 0; $i < $size; $i++) {
            // Definimos el color a utilizar
            $color = imagecolorallocate($this->imagen, $r, $g, $b);
            // Dibujamos una linea con el color
            if ($orientation == 'h')
                imageline($this->imagen, $i, 0, $i, $size, $color);
            else
                imageline($this->imagen, 0, $i, $size, $i, $color);

            // Incrementamos los valores de $r, $g y $b
            $r += $incr_r;
            $g += $incr_g;
            $b += $incr_b;
        }
    }

    function grHoriz($valor, $min, $max, $color1=null, $color2= null, $showval=0) {

        //if ($valor == $min) return;
        $imageX = imagesx($this->imagen);
        $imageY = imagesy($this->imagen);
        if ($max == 0)
            $max = 1;
        $paso = $imageX / $max;
        $valorC = $valor * $paso;

        if ($color1 == null)
            $color1 = array('red' => 0, 'green' => 255, 'blue' => 0); //verde
        if ($color2 == null)
            $color2 = array('red' => 255, 'green' => 0, 'blue' => 0); //rojo
            // Incremento en los canales
        $incr_r = ($color2['red'] - $color1['red']) / $imageX;
        $incr_g = ($color2['green'] - $color1['green']) / $imageX;
        $incr_b = ($color2['blue'] - $color1['blue']) / $imageX;

        // Valores correspondientes al primer color
        $r = $color1['red'];
        $g = $color1['green'];
        $b = $color1['blue'];
        for ($i = 0; $i < $valorC; $i++) {
            // Definimos el color a utilizar
            $color = imagecolorallocate($this->imagen, $r, $g, $b);
            // Dibujamos una linea con el color
            imageline($this->imagen, $i, 0, $i, $imageY, $color);

            // Incrementamos los valores de $r, $g y $b
            $r += $incr_r;
            $g += $incr_g;
            $b += $incr_b;
        }

        if ($showval != 0){
		if ($valor) {
                    $font = 4;
                    $txtColor1 = ImageColorAllocate($this->imagen, 0 , 0 ,0);
                    $txtColor2 = ImageColorAllocate($this->imagen, 200 , 200 ,200);

                    $txtX = ($imageX / 2 ) - ((strlen($valor)) * imagefontwidth($font) / 2);
                    $txtY = $imageY - 1 * imagefontheight($font);


                    ImageString($this->imagen, $font,  $txtX + 1, $txtY + 1,  $valor, $txtColor2 );

                    ImageString($this->imagen, $font,  $txtX , $txtY,  $valor, $txtColor1 );

		}
        }
    }

    function show($file=null) {
        //	 header ("Content-type: image/jpeg");
        switch ($this->tipo) {
            case 'jpg':
                header("Content-type: image/jpeg"); # We will create an *.jpg
                imagejpeg($this->imagen, $file);

                break;
            case 'png':
                header("Content-type: image/jpeg"); # We will create an *.png
                imagepng($this->imagen, $file);
                break;
            case 'gif':
                header("Content-type: image/gif"); # We will create an *.gif
                imagegif($this->imagen, $file);

                break;
        }
    }

}

?>