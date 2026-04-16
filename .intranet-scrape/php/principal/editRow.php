<?php
/*
 * Created on 07/11/2005
 *
 * To change the template for this generated file go to
 * Window - Preferences - PHPeclipse - PHP - Code Templates
 */

include ("./autoload.php");

$ArchivoConexion = "../funciones/conexion.php";
if (is_readable($ArchivoConexion))
    include($ArchivoConexion);

include ("./sessionCheck.php");

$MisDatos = new ContDatos("");
$MisDatos = Histrix_XmlReader::unserializeContainer(null, $instance);


$rowNumber      = $_POST['editrow'];

// get correct UI
$UI = 'UI_'.str_replace('-', '', $MisDatos->tipo);
$datos = new $UI($MisDatos);

$salida .= $datos->editRow($rowNumber);

echo $salida;

?>