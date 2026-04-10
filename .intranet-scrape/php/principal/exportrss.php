<?php
set_time_limit(300);
$DirectAccess = true;

include ("./autoload.php");
include_once('./inicio_sesion_remota.php');
include ("./sessionCheck.php");
include("../lib/feedCreator/feedcreator.class.php");

$titulo=$_GET['titulo'];

$xmldatos=$_GET['xmldatos'];
$format  = $_GET['format'];
$xmlOrig=$_REQUEST['xmlOrig'];
$subdir = $_GET["dir"];
$dirXML = '../xml/';

$xmlPath=$datosbase->xmlPath;
$dirXML = '../database/'.$xmlPath.'xml/';


//$MisDatos = leoXML($dirXML, $xmldatos, null, null, $subdir);

$xmlReader = new Histrix_XmlReader($dirXML, $xmldatos, null, null, $subdir);

if ($ContenedorReferente != '')
    $xmlReader->addReferentContainer($ContenedorReferente);

$xmlReader->addParameters($_GET);
$xmlReader->addParameters($_POST);

$MisDatos   = $xmlReader->getContainer();

$UI = 'UI_'.str_replace('-', '', $MisDatos->tipo);
$datos = new $UI($MisDatos);

$datos->setTitulo($MisDatos->tituloAbm);
$noprint= $datos->show();

$mixml = new Cont2XML($MisDatos, $xmldatos, null, null, true, true, null);

switch ($format) 
{
	case 'html':
		header ("Content-type: text/html");
		echo $noprint;
		break;

	case 'json':
		header ("Content-type: text/json");
  		$mixml->buildDataArray();
		echo json_encode($mixml->show(), JSON_PRETTY_PRINT);
		break;

	case 'xml':
  		header ("Content-type: text/xml");
  		$mixml->exportData();
		echo $mixml->show();
		break;

	
	default:
		$mixml->generateFeed();
		echo $mixml->show();
		break;

}
