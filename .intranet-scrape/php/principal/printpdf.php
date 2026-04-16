<?php

// Este script consume mucha memoria del servidor, necesito aumentarla para estos casos
ini_set('memory_limit', '512M');

// IMPRIMO EL CONTENEDOR EN FORMATO PDF
$DirectAccess=true;
include_once ("./autoload.php");
include_once ("../funciones/conexion.php");
include ("./sessionCheck.php");


//$xmldatos       = (isset($pdfnom)) ? $pdfnom : $_GET["pdfnom"];
$instance       = (isset($instance)) ? $instance : $_REQUEST["instance"];
$orientacion    = (isset($orientacion))? $orientacion : $_GET["__orientacion"];
$pagesize       = (isset($pagesize))? $pagesize : $_GET["__pagesize"];
$send           = (isset($_GET["send"]))?$_GET["send"]:'';
$destino = '';

if ($destino =='')      $destino = 'pdf';
if ($send    =='true') 	$destino = 'mail';
if (isset($printerName) && $printerName != '') $destino = 'printer';

if (isset($_GET['printer']) && $_GET['printer'] != '' && $_GET['printer'] != 'undefined') {
    $destino = 'printer';
    $printerName =$_GET['printer'];
}


//if ($orientacion != 'P' && $orientacion != 'L') $orientacion= 'P';



$MisDatos = new ContDatos("");
$MisDatos = Histrix_XmlReader::unserializeContainer(null, $instance);
$xmldatos = $MisDatos->xml;

// Reset Limit for Pagination cases
if (isset($MisDatos->paginar) && count($MisDatos->TablaTemporal->Tabla) >=  $MisDatos->paginar){
    // Redo the select statement
    unset($MisDatos->paginar);
    unset($MisDatos->limit);
    $MisDatos->Select();
    $MisDatos->CargoTablaTemporal();
}

/* NEW OO Printing Method */
$ContainerPrinter = new ContainerPrinter($MisDatos, $_GET);

$ContainerPrinter->target       = $destino; // pdf , mail or printer
$ContainerPrinter->copies       = (isset($_GET['copies']))?$_GET['copies']:1;
if (isset($printerName))
    $ContainerPrinter->printer      = $printerName;
if (isset($_GET['echo']))
    $ContainerPrinter->localecho    = $_GET['echo'];
if ($orientacion != '')
    $ContainerPrinter->orientation  = $orientacion;
if ($pagesize != '')
    $ContainerPrinter->pageSize  = $pagesize;

$ContainerPrinter->printContainer();

?>