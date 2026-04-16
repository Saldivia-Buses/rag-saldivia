<?php
/*
 * Created on 21/11/2008
 *
 * To change the template for this generated file go to
 * Window - Preferences - PHPeclipse - PHP - Code Templates
 */


$file	= $_GET['file'];
$subdir = $_GET["dir"];

include ("./autoload.php");

$path='../';
include ("./sessionCheck.php");



//$xmlPath=$datosbase->xmlPath;
//$dirXML = '../database/'.$xmlPath.'/xml/';
$dirXML = $_SESSION['dirXML'];

if ($subdir != '')
	$dirXML2 = $dirXML.$subdir.'/';
else
	$dirXML2 = $dirXML;


//$fileContent = file_get_contents($dirXML2.$file);


$content = stripslashes($_POST['contenido']);

if (is_file($dirXML2.$file)){
  try{
	$error = @file_put_contents($dirXML2.$file, $content);
	if ($error === false) throw new Exception('Error');
        echo 'Archivo Guardado: '.$dirXML2.$file;

	}
	catch (Exception $e){
            echo 'Error al guardar: '.$dirXML2.$file;
	
	}
}
else echo 'ERROR: '.$dirXML2.$file;


?>

