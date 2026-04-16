<?php
/*
 * Created on 16/09/2007
 * Luis M. Melgratti
 */
 $DirectAccess = true;
 include ("./autoload.php");
 include ("./sessionCheck.php");
 include_once('../funciones/utiles.php'); 
 
 $ancho = $_GET['ancho'];
 $alto  = $_GET['alto'];
 $max   = $_GET['max'];
 $min   = $_GET['min'];
 $valor = $_GET['valor'];
 $clase = $_GET['clase'];
 $c1 = (isset($_GET['c1']))?hex_to_rgb($_GET['c1']):array('red' => 0, 'green' => 255, 'blue' => 0); //verde;
 $c2 = (isset($_GET['c2']))?hex_to_rgb($_GET['c2']):array('red' => 255, 'green' => 0, 'blue' => 0); //rojo;

 $showval = $_REQUEST['showval'];

$tmphash = $ancho.$alto.$max.$min.$valor.$clase.$_GET['c1'].$_GET['c2'].$showval;

$dataPath = $_SESSION['datapath'];

if ($dataPath != '') {
    $tmpbase= '../database/'.$dataPath;
}
$tmpFile = $tmpbase.'/tmp/'.$tmphash.'.png';

if (!is_file($tmpFile) ) {
    $img = new Graficar($clase, $ancho, $alto);
    $img->crearImagen();
    switch ($clase){
    case "grHoriz":
    	$img->grHoriz($valor, $min, $max, $c1, $c2, $showval);
		break;
	}
	$img->show($tmpFile);

}


$image = new Imagick($tmpFile);
header('Content-type: image/png');

$image->setImageFormat( "png" );
echo $image;






?>