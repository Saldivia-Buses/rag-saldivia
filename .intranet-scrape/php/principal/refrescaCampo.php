<?php
/**
 * Refresh field
 * Created on 07/11/2005
 * @package Histrix
 * @author Luis M. Melgratti <luis@estudiogenus.com>
 * @link   http://www.estudiogenus.com
 * @date   2005-11-07
 */

include ("./autoload.php");

$ArchivoConexion = "../funciones/conexion.php";
if (is_readable($ArchivoConexion))
    include($ArchivoConexion);

include ("./sessionCheck.php");


$instance      = (isset($_REQUEST['instance']) && $_REQUEST['instance'] != 'undefined')?$_REQUEST['instance']:'';

$dataContainer = new ContDatos("");
$dataContainer = Histrix_XmlReader::unserializeContainer(null, $instance);


if (!is_object($dataContainer)) {
    $xmlFileName   = isset($_GET["xmldatos"])?$_GET["xmldatos"]:$instance;
    header('HTTP/1.1 400 Error: Container '.$xmlFileName.' not Found');
    echo $xmlFileName.' '.$instance;
    die();
}



if (isset($_REQUEST['campo'])){

    // refresh just one field in a form

    $fieldName 	   = $_POST['campo'];
    $destinationField = $_POST['destino'];
    $value  	= $_POST['valor'];

    $opcion 	= isset($_POST['opcion'])?$_POST['opcion']:'';
    $editrow 	= isset($_POST['editrow'])?$_POST['editrow']:'';

    $arraydato[$fieldName] = $value;

    $operador 	= '=';
    $reemplazo 	= 'reemplazo';
    $ObjCampoF 	= $dataContainer->getCampo($fieldName);

    if (isset($ObjCampoF->contAyuda) && $ObjCampoF->contAyuda !=''){
        $contExterno = 	$ObjCampoF->contAyuda;
    } else {
        if(isset($ObjCampoF->contExterno))
            $contExterno = 	$ObjCampoF->contExterno;
    }

    if ($ObjCampoF->ContParametro !='')
        $contExterno = 	$ObjCampoF->ContParametro;

    if ($contExterno !='' && $destinationField != '') {

        $ObjCampoDes = $contExterno->getCampo($destinationField);

        $tipo = Types::getTypeXSD($ObjCampoDes->TipoDato, 'xsd:integer');

        $quotes ="";

        if ($tipo !='xsd:integer' && $tipo !='xsd:decimal' ) {
            $quotes ="'";
        }

        $contExterno->addCondicion($destinationField, $operador, $quotes.$value.$quotes, ' and ', $reemplazo);

        if ($opcion == 'conservarValor' || (isset($ObjCampoDes->local) && $ObjCampoDes->local == 'true') ) {
            $contExterno->setFieldValue($destinationField, $value, 'both');        
            $value =  isset($ObjCampoF->valor)?$ObjCampoF->valor:'';
        }

        
        unset ($dataContainer->getCampo($fieldName)->opcion);

        if (isset($ObjCampoF->contExterno) && $ObjCampoF->contExterno != '')
            $dataContainer->getCampo($fieldName)->addContenedor($contExterno);
            /*
        $newValue = isset($ObjCampoF->valor)?$ObjCampoF->valor:'';

        if ($fieldName != $destinationField) {
            $newValue = '';
        } else $newValue = $value;
          */
        // parameter
        $value    = '';
        $newValue = '';
        if (isset($ObjCampoF->ContParametro) && $ObjCampoF->ContParametro !='') {
            $ObjCampoF->ContParametro = $contExterno;
            $dataContainer->getParametros($ObjCampoF);
            $value =  $ObjCampoF->valor;
            $newValue = $ObjCampoF->valor;
        }

    } else {

        $newValue = $value;
        $tipo = Types::getTypeXSD($ObjCampoF->TipoDato, 'xsd:integer');
        if ($tipo =='xsd:integer' || $tipo =='xsd:decimal' ) {

            $dataContainer->addCondicion($fieldName, $operador, $value, ' and ', $reemplazo);
        } else {
            $dataContainer->addCondicion($fieldName, $operador, "'".$value."'", ' and ', $reemplazo);
        }

        $dataContainer->setFieldValue($fieldName, $value, 'both');        
    }

    // Muestro el input
    $UI = 'UI_'.str_replace('-', '', $dataContainer->tipo);
    $datos = new $UI($dataContainer);


    if (isset($dataContainer->filtros) && $dataContainer->filtros != '') {
        foreach ($dataContainer->filtros as $nomfiltro => $objFiltro) {
            if ($objFiltro->campo != $fieldName) continue;
            $fieldNameFiltro = $dataContainer->getCampo($objFiltro->campo);
            if ($fieldNameFiltro == null)	continue;
            $operadores = $fieldNameFiltro->getOperadores();
            $atributos['operador']  = htmlentities($objFiltro->operador);
        }
    }

    if (!isset($newValue) || $editrow != '') {
        $newValue = $value;
    }

    $atributos['editrow'] = true;


    $input ='';
    if (isset($ObjCampoF->contExterno) && isset($ObjCampoF->esTabla) && $ObjCampoF->esTabla) {
        // WILL IT BROKE EVERYTHING???
        //  $ObjCampoF->refreshInnerDataContainer($dataContainer);

        $ObjCampoF->contExterno->tabindex = $dataContainer->tabindex +10;
        $ObjCampoF->contExterno->esInterno = true;
        
        $UI = 'UI_'.str_replace('-', '', $ObjCampoF->contExterno->tipo);
        $abmDatosDet = new $UI($ObjCampoF->contExterno);
        
        $ObjCampoF->contExterno->xmlpadre = $dataContainer->xml;
        //$ObjCampoF->contExterno->xmlOrig = $dataContainer->xmlOrig;
        
        $ObjCampoF->contExterno->xmlOrig = $dataContainer->xml;
        $ObjCampoF->contExterno->getSelect();

        $input  = '<div id="'.$ObjCampoF->NombreCampo.'">';
        $input .= $abmDatosDet->showTablaInt(null, $ObjCampoF->contExterno->idxml, '', 'false', true, 'noform', null, $ObjCampoF);
        $input .= '</div>';

        // Increase Tabindex
        $dataContainer->tabindex += $ObjCampoF->contExterno->tabindex;
        Histrix_XmlReader::serializeContainer($ObjCampoF->contExterno);

    }

    if ($input == '') {

        // si el destino del valor es diferente al del campo
       /*
        if ($destino != $ObjCampoF->NombreCampo && $destino != '')
            $newValue = $ObjCampoF->getValor();
    */

    	if ($newValue=='') $newValue = $ObjCampoF->getValor();
        

        if (isset($ObjCampoF->onRefresh) && $ObjCampoF->onRefresh == 'focus') {
            $atributos['class'][]='refresh';
        }
        
        if (is_object($ObjCampoF)) {
            $input = $ObjCampoF->renderInput($datos, 'Form'.$dataContainer->idxml, '', $newValue, null,  null, $atributos);
        } else {
            //  die($fieldName);
        }


    }

    echo $input;

} else {
    // refresh whole container (paginate an refresh table)

    $dataContainer->paginaActual= isset($_POST["pagina"])?$_POST["pagina"]:'' ;
    $dataContainer->llenoTemporal= false ;

    $UI = 'UI_'.str_replace('-', '', $dataContainer->tipo);
    $datos = new $UI($dataContainer);

    $nocant   = isset($_GET["nocant"])?$_GET["nocant"]:''; // do not show totals
    $select   = isset($_GET["select"])?$_GET["select"]:''; // do not perform another select

    if ($select == 'false'){
        $datos->nosel = 'true';
    }

    $datos->esInterno = true;
    if ($xmlpadre != '') $form  = 'Form'.$xmlpadre;
    echo  $datos->showTablaInt('', $dataContainer->idxml, '', $nocant, null, $form);

}

Histrix_XmlReader::serializeContainer($dataContainer);

?>