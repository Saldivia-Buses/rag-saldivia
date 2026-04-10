<?php
// Este script consume mucha memoria del servidor, necesito aumentarla para estos casos
ini_set('memory_limit', '512M');

// IMPRIMO EL CONTENEDOR EN FORMATO PDF
require_once ('../lib/fpdf/fpdf.php');

define('FPDF_FONTPATH','../lib/fpdf/font/');
require ('./htmlparser.inc.php');

$DirectAccess=true;
include('./autoload.php');
include_once ("../funciones/conexion.php");
include ("./sessionCheck.php");

// Funciones de lectura del XML
//include_once('./pdf_clase.php');

$xmlPath = $datosbase->xmlPath;
$dir = ($_GET['dir'] != '')?$_GET['dir'].'/':'';
$dirXML = '../database/'.$xmlPath.'/xml/'.$dir;

//$dirXML = '../xml/'.$xmlPath;




$instance       = (isset($instance)) ? $instance : $_REQUEST["instance"];

//$anchocel = $_GET["anchocel"];

$orientacion = $_GET["__orientacion"];

$MisDatos = new ContDatos("");
$MisDatos = Histrix_XmlReader::unserializeContainer(null, $instance);
$xmldatos = $MisDatos->xml;

$pageBreak = (isset($MisDatos->pageBreak)) ? $MisDatos->pageBreak : '';

// Inicio el PDF
$tamPagina = isset($_GET['__pagesize']) ? $_GET['__pagesize'] : 'A4';

$pdf = new Histrix_pdf($orientacion, 'mm', $tamPagina);
$pdf->sinNumero = 'true';

if (isset($MisDatos->PDFsincabecera ) && $MisDatos->PDFsincabecera == "true")
    $pdf->sincab = true;

$titulo = utf8_decode($MisDatos->getTitulo());


// Obtengo el Contenedor con la tabla a recorrer
$Tablatemp   = $MisDatos->TablaTemporal->datos();
$listaCampos = $MisDatos->camposaMostrar();
$y = 0;
$lineCounter = 0;

$campoCond = isset($MisDatos->CampoCond) ? $MisDatos->CampoCond : '';

foreach ($Tablatemp as $orden => $row) {
    set_time_limit  ( 60  );
    $y++;
    $lineCounter++;

    $filadetalle = '';
    
    if ( $campoCond != '' && is_array($row) )
    {
		if( isset($row[$campoCond]) )
	
		  if ($row[$campoCond] == 'false' || $row[$campoCond] == false || $row[$campoCond] == 0) 
			  continue;
    }

    
    if ($row !='')
        foreach($listaCampos as $Nnombre => $nombrelista) {

            $parametros  ='';
            $Valcampo = $row[$nombrelista];
            $ObjCampo = $MisDatos->getCampo($nombrelista);
            //		if ($ObjCampo->paring != ''){
            //			foreach($ObjCampo->paring as $destino => $ncampo){
            //				$parametros .= '&amp;'.$destino.'='.$row[$ncampo['valor']];
            //			}
            //		}

            if($ObjCampo != null) {
            //$detallecampo = $ObjCampo->getUrlVariableString($Valcampo);
            // simulo la vinculacion del detalle pero lleno la variable $_GET

                if (isset($ObjCampo->Detalle) && $ObjCampo->Detalle != '') {
                    foreach ($ObjCampo->Detalle as $ndet => $deta) {
                        $_GET[$deta]=$Valcampo;
                    }
                }
            }
        }


    if($pageBreak != 'false' || $y == 1) 
    {
  		// Imprimo Cada PDF
  		$pdf->maxY = 1;
        $pdf->titulo = $MisDatos->getTitulo();
        $pdf->AliasNbPages();
        $pdf->AddPage();
        if (isset($MisDatos->PDFsincabecera) && $MisDatos->PDFsincabecera == "true")
            $pdf->sincab = true;
    }


    $pdf->SetFont('helvetica', '', 9);

	$detalle   = $MisDatos->detalle;
	if (isset($row[$MisDatos->detalle])){
		$detalle = $row[$MisDatos->detalle];
	
	}
	$dirDet = (dirname($detalle) != '' && dirname($detalle) != '.')?dirname($detalle):$_GET['dir'];
	$detalle = basename($detalle);

    $xmlReader = new Histrix_XmlReader('../database/'.$xmlPath.'/xml/', $detalle, null, null, $dirDet);
    $xmlReader->serialize = false;
    $xmlReader->addParameters($_GET);
    $xmlReader->addParameters($_POST);

    $Contenedor = $xmlReader->getContainer();


 // remove unserialization and code execution
    if (isset($Contenedor->pdfCode)) {
        $code = (string) $Contenedor->pdfCode;
        if ($code != '') {
            eval($code);
        }
    }

    $UI = 'UI_'.str_replace('-', '', $Contenedor->tipo);
    $datos = new $UI($Contenedor);

    $datos->setTitulo($Contenedor->tituloAbm);

    if (isset ($Contenedor->CabeceraMov)) {

        foreach ($Contenedor->CabeceraMov as $Ncab => $cabecera) {

            $UI = 'UI_'.str_replace('-', '', $cabecera->tipo);
            $datosCabecera = new $UI($cabecera);
             $datosCabecera->Show("FormuCab");

            $titulo = utf8_decode($cabecera->getTitulo());

            $tablacab[$Ncab] = $pdf->impAbm($cabecera);

            $wcab = $pdf->setAnchoCol($tablacab[$Ncab]);

            $pdf->SetY($pdf->maxY);
            $pdf->SetY($pdf->GetY() + 4);
            $pdf->WriteTable($tablacab[$Ncab], $wcab);

        }
    }

    $datos->show();

    
    if (isset ($Contenedor->CabeceraMov)) {
        foreach ($Contenedor->CabeceraMov as $NCabecera => $ContCab) {
            $titulo = utf8_decode($ContCab->getTitulo());

        }
    }
    if (isset ($Contenedor->filtros) || isset ($Contenedor->filtroPrincipal)) {
        $pdf->SetY($pdf->GetY() + 2);
        $datos = $pdf->impFiltros($Contenedor);
        $w = $pdf->setAnchoCol($datos);
        $anchofiltros = array_sum($w);
        $pdf->WriteTable($datos, $w);
    }

    $pdf->SetY($pdf->GetY() + 2);

    switch ($Contenedor->tipoAbm) {
        case "ficha":
            $pdf->SetY(28);

            $tabla = $pdf->impAbm($Contenedor);

            $wcab = $pdf->setAnchoCol($tabla) ;

            $pdf->SetY($pdf->GetY() + 4);

            $pdf->WriteTable($tabla, $wcab);

            break;
        case "arbol":
            $arbol = $Contenedor->ARBOL;
            $startX         = 5;
            //		$nodeFormat     = '<%k>';
            //		$nodeFormat     = '';
            //		$childFormat    = '<%k> = [%v]';
            $childFormat    = '%v';
            $w              = 100;
            $h              = 3;
            $border         = 0;
            $fill           = 0;
            $align          = 'L';
            $indent         = 8;
            $vspacing       = 0;
            $pdf->SetY($pdf->GetY() + 4);
            $pdf->MakeTree($arbol,$startX,$nodeFormat,$childFormat,$w,$h,$border,$fill,$align,$indent,$vspacing);

            break;


        default:
        //$tablaHtml = $datos->showTablaInt('noecho', '', '');
            $pdf->impTabla($Contenedor);

            break;
    }


    //$pdf->WriteHTML($tablaHtml);
    $pdf->showGraficos($Contenedor);
// Termino 1 pdf
}
$pdf->Output($titulo . '.pdf', 'I');

?>
