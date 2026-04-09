<?php

/**
* Drop files Html5 updload
*
*/


include ("./autoload.php");
include ("../funciones/utiles.php");
include ("./sessionCheck.php");

$db = $_SESSION["db"];
$datosbase = Cache::getCache('datosbase'.$db);
if ($datosbase === false){
    $config = new config('config.xml', '../database/', $db);
    $datosbase = $config->bases[$db];
    Cache::setCache('datosbase'.$db, $datosbase);
}

$xmlPath=$datosbase->xmlPath;

$basePath = dirname(__FILE__).'/../database/'.$xmlPath.'xml/';

// TODO get max image upload

$save_path = $_GET["dir"];
$maxwidth  = isset($_REQUEST["maxWidth"])?$_REQUEST["maxWidth"]:'';
$path = $basePath.$save_path;
$path = str_replace('//','/',$path);
$ft = new FileStreamer(); 

loger('subir archivo '.$ft->fileName().' a '.$path, 'drop.log');

$ft->setDestination($path);
$ft->receive();

if ($maxwidth != '' && $maxwidth > 0){
    try {
        $image = new Imagick($ft->filePath());
	if ($image->getImageWidth() > $maxwidth){
    //    $image->setResolution( 300, 300 );
	    $image->adaptiveResizeImage($maxwidth, 0);
            $image->writeImage($ft->filePath());

	}

    } catch(Exception $ex) {
	// not a valid Image
    }
}


//die();
?>