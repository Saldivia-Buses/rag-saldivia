<?php
/* 
 * Dinamic Gradient generator
 *
 *
 */
$DirectAccess = true;
/*
include ("autoload.php");
include ("./sessionCheck.php");
//include_once('./img_clase.php');
include_once('../funciones/utiles.php');
*/
$orientation= $_GET['o'];
$size       = $_GET['s'];
$color1     = $_GET['c1'];
$color2     = $_GET['c2'];
$gw = 1;

$tmphash = 'tmp-gradient'.md5($orientation.$size.$color1.$color2);
$tmpFile = '/tmp/'.$tmphash.'.jpg';


if (!is_file($tmpFile)){


    if (@class_exists('Imagick')){

        /*** new imagick object ***/
        $im = new Imagick();

        if ($orientation == 'h'){
             /*** a new image with gradient ***/
            $im->newPseudoImage( $gw, $size, "gradient:#$color2-#$color1" );
            $im->rotateImage(new ImagickPixel(), 90);
        }else {
            $im->newPseudoImage( $gw, $size, "gradient:#$color1-#$color2" );
        }
        $im->setImageFormat('jpeg');
        $im->writeImage($tmpFile);

    }
    else {
        // GD Version
        if ($orientation=='h'){
            $width  = $size;
            $height = $gw;
        } else {
            $height = $size;
            $width  = $gw;
        }

        if($color1 != '')
         $rgb1 = hex_to_rgb($color1);

        if($color2 != '')
         $rgb2 = hex_to_rgb($color2);


        $img = new Graficar('gradient', $width, $height);
        $img->crearImagen();
        $img->gradient($orientation, $size, $rgb1, $rgb2);
        $img->show($tmpFile);
    }
    
}
    
header ("Content-type: image/jpeg"); # We will create an *.jpg
header("Expires: Thu, 15 Apr 2020 20:00:00 GMT"); // Date in the Future
readfile($tmpFile);
?>