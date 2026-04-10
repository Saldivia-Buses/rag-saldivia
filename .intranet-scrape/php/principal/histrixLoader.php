<?php
/* HISTRIX PROGRAM LOADER
 * Created on 26/10/2005
 * Luis M. Melgratti
*/

include_once ("./autoload.php");                 // Check if a valid session exists
include_once ("../funciones/conexion.php");
include_once ("../funciones/utiles.php"); 	// Config
include      ("./sessionCheck.php");            // Check if a valid session exists

$xml         = $_GET["xml"];                            // xml to parse
$subdir      = isset($_GET["dir"])?$_GET["dir"]:'';     // Optional Subdir

// parse dir

$dirfile = dirfile($xml );
$helperDir = $dirfile['dir'];
$xml = $dirfile['file'];

if ($helperDir != '')
    $subdir= $helperDir;


$titulo_div  = isset($_REQUEST['titulo_div'])?$_REQUEST['titulo_div']:''; //
$helplink    = isset($_REQUEST['helplink'])?$_REQUEST['helplink']:''; // help file to link
$xmlOrig     = isset($_REQUEST['xmlOrig'])? $_REQUEST['xmlOrig']:''; // xml that make the call
//$__email     = isset($_REQUEST['__email'])? $_REQUEST['__email']:'';

/* Get configurationValues */

$registry = &Registry::getInstance();
$i18n   = $registry->get('i18n');
$dirXML = $registry->get('xmlPath');

/*
 * Load XML file and generate GUI
*/

// If we have an XML
if (isset($xml)) {

    if ( (isset($_GET['_xmlreferente']) && isset($_GET['_param_in']))
         || isset($_GET["cierraproceso"])) {

        $ContenedorReferente = new ContDatos("");
        $ContenedorReferente = Histrix_XmlReader::unserializeContainer(null, $_REQUEST['parentInstance']);
    }

    if (isset($_GET["cierraproceso"])) {
        // busco los campos que se incorporan
        // delete previous movements
        unset($ContenedorReferente->cadenasSQL);
                 
        $_GET['_param_in']=$ContenedorReferente->paring;
        
        
    }

    //  load Container

    $xmlReader = new Histrix_XmlReader($dirXML, $xml, null, $xmlOrig, $subdir);

    if (isset($ContenedorReferente) ) {
        $xmlReader->addReferentContainer($ContenedorReferente);

        if ($_REQUEST['__row'] != '' /*   && $Contenedor->resume == 'true'  */) {

            $oldInstance = $ContenedorReferente->childContainers[$xml];

            if ($_REQUEST['__col'] != '') {

                $savedContainer = new ContDatos("");
                $savedContainer = $ContenedorReferente->TablaTemporal->getMetadata($_REQUEST['__row'],  $_REQUEST['__col']); 

            }

        }
    }


    // resumen saved container
    if (isset($savedContainer)
    // && get_class($savedContainer) == 'ContDatos' 
        ) {

        $Contenedor = $savedContainer;
        $Contenedor->llenoTemporal = 'false';
        $Contenedor->__savedState  = true;

    } else {
        $xmlReader->addParameters($_GET);
        $xmlReader->addParameters($_POST);

        $Contenedor   = $xmlReader->getContainer(true);
        $Contenedor->__savedState  = false;

    }

    $Contenedor->__parent_row = $_REQUEST['__row'];
    $Contenedor->__parent_col = $_REQUEST['__col'];

    // get old Data from container

    /*    
    if ($Contenedor->resume == 'true' && $oldInstance != ''){

        $Contenedor   = $xmlReader->getContainer(true, $oldInstance);
        $Contenedor->llenoTemporal = 'false';

    }
     */


    //$serializados = $xmlReader->getSerializedContainers();
    $cont_filtro  = $xmlReader->getFilterContainer();
    $filtro       = $xmlReader->getFilterField();

    $Contenedor->_getParameters = $_GET;
    
    
    if (isset($_REQUEST['parentInstance']))
     $Contenedor->parentInstance = $_REQUEST['parentInstance'];

    if (isset($ContenedorReferente) && $ContenedorReferente->tabindex != '') {
        $Contenedor->tabindex =  $ContenedorReferente->tabindex; // Force tabindex into header From

        ///// propagate transaccional
	   if (isset($ContenedorReferente->transaccional)){
	       $Contenedor->transaccional=  $ContenedorReferente->transaccional; 
       }
        


        // IMPORT TEMPTABLES FROM CALLING XML
        // this allows temp tables been passed around between xmls

        if (isset($Contenedor->importTempTable)) {
            if (isset($ContenedorReferente->tempTables[$Contenedor->importTempTable])) {
                $Contenedor->tempTables[$Contenedor->importTempTable] = $ContenedorReferente->tempTables[$Contenedor->importTempTable];

                unset($Contenedor->tempTables[$Contenedor->importTempTable]->row);
                unset($Contenedor->tempTables[$Contenedor->importTempTable]->relationship);
                // propagate temptables

                //    print_r($Contenedor->xml);

                //    echo '<pre>';
                //    print_r($Contenedor->tempTables);
                //    echo '</pre>';

                //die();

                $fieldArray = $Contenedor->camposaMostrar();
    	        foreach ($fieldArray as $number => $id_field) {

                        $Field = $Contenedor->getCampo( $id_field);
                        if (isset($Field->contExterno)  && isset($Field->contExterno->importTempTable)) {

                            $Field->contExterno->tempTables[$Field->contExterno->importTempTable] = $ContenedorReferente->tempTables[$Field->contExterno->importTempTable];

                            Histrix_XmlReader::serializeContainer($Field->contExterno);
                        }
                }

            }
        }
    }

    if (isset ($_GET["xmlOrig"])) {
        $Contenedor->xmlOrig = $xmlOrig;
    }
    if (isset ($_GET["cierraproceso"])) {
        $Contenedor->cerrar_proceso = 'true';
        $Contenedor->xmlReferente = $Contenedor->xmlOrig;
    }

} else {

    // If we dont have an XML file but we have a Table name at least
    $tablaBase 	= $_GET["tabla"];
    if ($tablaBase != '') {
        $Contenedor = new ContDatos($tablaBase, $tablaBase, 'abm');
        $Contenedor->tipoAbm="abm";
        $Contenedor->xml = $tablaBase;

        foreach($Contenedor->tablas[$Contenedor->TablaBase]->campos as $fieldName => $field) {
            $Contenedor->addCampo($fieldName);
        }
    }
    else
        echo '<div class="error">'.$i18n['errorMisingTable'].'</div>';
}

//////////////////////////////////////
// add winid info to container
// this will be used to destroy session info if one window close
// ////////////////////////////////////
$Contenedor->__winid = $_REQUEST['__winid'];


// special type of container
if ($Contenedor->tipo=='dashboard') {
    echo $Contenedor->display();
    unset ($Contenedor);
}

if (is_object($Contenedor)) {

    if ($titulo_div != '')
        $Contenedor->titulo_div = $titulo_div;
    if ($helplink != '')
        $Contenedor->helplink = $helplink;

    $UI = 'UI_'.str_replace('-', '', $Contenedor->tipo);
    $datos = new $UI($Contenedor);

    // Default focus element
    $formFoco = 'Form'.$Contenedor->idxml;

    // show header FORM
    if (isset ($Contenedor->CabeceraMov)) {
        foreach ($Contenedor->CabeceraMov as $Ncab => $cabecera) {

            $UI = 'UI_'.str_replace('-', '', $cabecera->tipo);
            $datosCabecera = new $UI($cabecera);

            $html[]= $datosCabecera->Show("FormuCab");

            $formFoco = 'Form'.$cabecera->xml;            // Focus header Form
            $Contenedor->tabindex =  $cabecera->tabindex; // Force tabindex into header From
        }
    }

    $javascriptExec[] = " foco('".$formFoco."');";

    // TODO: remove and set it in UI class
    // Agrego Filtro
    if (isset($cont_filtro)) {
        $Contenedor->filtroPrincipal = $cont_filtro;
        $Contenedor->campofiltroPrincipal = $filtro;
        $rs = $datos->setFiltro($cont_filtro, $filtro);
        $Contenedor->Ocultar($filtro, true);
        // 	Tengo que buscar el primer dato y filtrar por eso
        $fila = _fetch_array($rs);
        $i = 0;
        if ($fila != '')
            foreach ($fila as $clave => $valor) {
                $i ++;
                if ($i == 1) {
                    $valorFiltro = $valor;
                    $claveFiltro = $clave;
                }
            }
        $campoFiltro = $Contenedor->getCampo($filtro);

        $tipo = Types::getTypeXSD($campoFiltro->TipoDato, 'xsd:integer');

        if ($tipo =='xsd:integer' || $tipo =='xsd:decimal' ) {
            $Contenedor->addCondicion($filtro, "=", 	$valorFiltro	, ' and ', 'reemplazo');
        } else {
            $Contenedor->addCondicion($filtro, "=", "'".$valorFiltro."'", ' and ', 'reemplazo');
        }

        $Contenedor->setCampo($filtro, $valorFiltro);
        $Contenedor->setNuevoValorCampo($filtro, $valorFiltro);
        $cont_filtro->Select();
    }

    if (isset($_GET['xmlsub'])) {
        $datos->detalle = true;
        $Contenedor->esDetalle = true;

        // Asume data already loaded in this kind of xml
        if ($Contenedor->tipo=='ficha') {
            $Contenedor->preFetch='true';
            $Contenedor->onDuplicateKey = "true";
        }
    }

    if (isset($_POST['__inlineid']) && $_POST['__inlineid']!='' ) {
        $Contenedor->__inline   = $_POST['__inline'];
        $Contenedor->__inlineid = $_POST['__inlineid'];
        
    }

    // add custom init javascript
    $javascriptExec[]= $datos->initialJavascript();
    
    
    // Render HML
    $html[]= $datos->show();

    // EXPORT CONTAINER DATA INTO A ZIP FILE

    if (isset($Contenedor->exportData) && $Contenedor->exportData == 'true') {
        echo ExportData($Contenedor);
    }

    // reserializo (ver de sacar esto)
    
    $serializedId = Histrix_XmlReader::serializeContainer($Contenedor);

    //$_SESSION['SERIALIZADOS'.$serializedId] = $serializados;

    if (isset($_GET['autoprint'])){
        $Contenedor->autoprint =$_GET['autoprint'];
    }
    if (isset($_GET['automail'])){
        $Contenedor->automail =$_GET['automail'];
    }
    

    if (((isset($_GET['autoprint']) && $_GET['autoprint']=='true') ||
        (isset($Contenedor->autoprint) && $Contenedor->autoprint == 'true') ||
        (isset($_GET['automail']) && $_GET['automail']=='true') ||
        (isset($Contenedor->automail) && $Contenedor->automail == 'true')) 

        )
        {
        
        unset($javascriptExec);


        if ($Contenedor->autoprint == 'true' || $Contenedor->automail == 'true') {

            $orientation = (isset($Contenedor->PDForientation))?$Contenedor->PDForientation:'P';
            $pdfnom = $Contenedor->xml;

            $title = urlencode(($Contenedor->titulo_div != '')? utf8_decode($Contenedor->titulo_div) : utf8_decode($Contenedor->getTitulo() ) );
            $instance = $Contenedor->getInstance();
	    $title = urlencode($title);

            $pdf = 'printpdf.php/'.$title.'.pdf?instance='.$instance.'&pdfnom='.$Contenedor->xml.'&__orientacion='.$orientation;
            $uid = UID::getUID($Contenedor->idxml, true);

            // EMBED METHOD?? IS PORTABLE? let's try
            // PREVENTS ANOTHER TRIP TO THE SERVER?
            $embed = '<object id="pdf" classid="clsid:CA8A9780-280D-11CF-A24D-444553540000" width="100%" height="100%">
            <param name="SRC" value="'.$pdf.'" />
            <embed src="'.$pdf.'" width="100%" height="100%" type="application/pdf">
            Loading...
            <noembed> Your browser does not support embedded PDF files. </noembed>
             </embed>
             </object>';

            // iframe method

	    // Test if there is internet connection
	    
	    $fp = Histrix_Functions::hasInternetConnection();

	    //if the socket failed it's offline...
	    
	    if (!$fp) {
                $btn = new Html_label('Sin conexion disponible, imposible enviar', '' ,'enviar');
	    } else {
                $btn = new Html_button('Enviar', "../img/mail_generic.png" ,'enviar');
	        $btn->addEvent('onclick', 'Histrix.imprimirpdf(\''.$Contenedor->xml.'\', \''.$Contenedor->getTitulo().'\', null,\'send\' , null, \''.$Contenedor->getInstance().'\')');

    	        if ($_SESSION['email'] == ''){
        	    $btn->addParameter('disabled', 'disabled');
                    $btn->addParameter('class', 'ui-state-disabled');
	            $btn->addParameter('title', $i18n['emaildisabled']);
                }
	    }
	    
	    
            if ($Contenedor->automail == 'true'){
        	// add send windows
                //include('javascript.php');
                /*echo '		<script type="text/javascript" src="../funciones/concat.php?type=javascript"></script>';
                echo '		<script type="text/javascript" src="../javascript/histrix.js"></script>'; */

                $javascriptExec[]= 'Histrix.imprimirpdf(\''.$Contenedor->xml.'\', \''.$Contenedor->getTitulo().'\', null,\'send\' , null, \''.$Contenedor->getInstance().'\')';
                
	    }
            if ($Contenedor->autoprint == 'true'){
                $iframe   = $btn->show();
	        $iframe .=  '<iframe class="pdfFrame" id="'.$uid.'" width="99.5%" height="460" src="'.$pdf.'" type="application/pdf" instance="'.$Contenedor->getInstance().'" />';
	        $javascriptExec[]= 'Histrix.resizeAll();';
	        
	    }
	    else {
//		IF THERE IS NO AUTOPRINT REMOVE FRAME
        	$javascriptExec[]= '$(\'#PRN'.$Contenedor->idxml.'\').remove();';
	    }
    	    if ($_GET['printer'] == '')	
                echo $iframe;
            
        }
        //else {
            if ($_GET['printer'] !='') {
                $instance = $Contenedor->getInstance();
                $pdfnom  = $Contenedor->xml;
                $destino = 'printer';
                $orientacion = 'P';
                //loger($_GET, 'print');
                include "printpdf.php";

            }
            else {
            /*
                $orientacion = 'P';

                $formPrint = 'FormPrint'.$xml;
                $titulo = $Contenedor->tituloAbm;
                $carga = 'Histrix.imprimirpdf(\''.$xml.'\', \''.$titulo.'\', \''.$formPrint.'\' );';
                $print = '<script type="text/javascript">';
                $print .= $carga;
                $print .= '</script>';
                
                echo $print;
                */
            }
        //}
    }
    else {
        //  print output to
        if (isset($html))
            foreach($html as $n => $code) {
                echo $code;
            }
    }
/*
    if ($__email != ''){
        $javascriptExec[]=  'window.opener.Histrix.imprimirpdf(\'' . $__email . '\', \'SEND\', null,\'send\' , null  , \''. $Contenedor->getInstance() .'\' );';
        $javascriptExec[]=  'Histrix.imprimirpdf(\'' . $__email . '\', \'SEND\', null,\'send\' , null  , \''. $Contenedor->getInstance() .'\' );';
    }
*/


// Generate javascript functions
    
    echo Html::scriptTag($javascriptExec);
}
function ExportData($Contenedor) {

    $mixml = new Cont2XML($Contenedor);
    $mixml->exportData();
    $mixml->out('F', $Contenedor->xml, $Contenedor->titulo);

    $filename = $Contenedor->idxml;
    $titulo = $Contenedor->titulo;
    //fileName
    if (isset($Contenedor->exportFileName)) {
        $filename = $Contenedor->getCampo($Contenedor->exportFileName)->getValor();
        $titulo = $filename;
    }

    // ZIP generated files
    $datapath    = $_SESSION["datapath"];
    $dir = '../database/'.$datapath.'/tmp/';
    $zipname = $dir.$filename.'.zip';
    if (is_file($zipname))
        unlink($zipname);
    $zip = new ZipArchive();
    $zip->open($zipname, ZIPARCHIVE::CREATE);
    // zip all xml generated in xml exportData Method
    if ($mixml->xmlfiles != '')
        foreach($mixml->xmlfiles as $n => $xmlfilename) {
            $zip->addFile($xmlfilename, $n);
            //  unlink($xmlfilename);
        }

    $zip->close();

    // GENERATE DOWNLOAD LINK
    $downloadLink = Archivo::downloadButton($zipname, $i18n['download'].' '.$titulo);
    return $downloadLink;
}

?>
