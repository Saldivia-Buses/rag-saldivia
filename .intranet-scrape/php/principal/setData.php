<?php
/*
 * Created on 07/11/2005
 *
 * To change the template for this generated file go to
 * Window - Preferences - PHPeclipse - PHP - Code Templates
*/
if ($_REQUEST['instance'] == '' || $_REQUEST['instance'] == 'undefined') 
    die('SetData empty instance');


include ("./autoload.php");
require ("../funciones/utiles.php");


$ArchivoConexion = "../funciones/conexion.php";
if (is_readable($ArchivoConexion))
    include($ArchivoConexion);
  
include ("./sessionCheck.php");


//$tiempo = processing_time();

header('Content-Type: text/xml');


/* Main Container */
$MisDatos = new ContDatos("");
$MisDatos = Histrix_XmlReader::unserializeContainer(null, $_REQUEST['instance']);
$xmldatos = $MisDatos->xml;

if (!is_object($MisDatos)) 
    die('no object found:'.$_REQUEST['instance']);

$xmldatossub = (isset($_GET["xmldatossub"]))?$_GET["xmldatossub"]:'';

if ($xmldatossub != '' && $xmldatossub != $xmldatos) {
    if (isset ($MisDatos->CabeceraMov)) {
        foreach ($MisDatos->CabeceraMov as $NCabecera => $ContCab) {
            $cabInstance = $ContCab->getInstance;
        }
        
        $xmldatossub = $ContCab->xml;
        $MisDatosCab = Histrix_XmlReader::unserializeContainer($ContCab);
    }
}


$show=true;

if(isset($_GET["_show"]) && ($_GET["_show"]=='false' || $_GET["_show"]=='null') )
    $show=false;

$mijson = (isset($_POST['mijson'])) ? $_POST["mijson"] : '';
if ($mijson !='') {

    $arrayJSON= json_decode($mijson, true);
    if ($arrayJSON == NULL)
        $arrayJSON= json_decode(stripslashes($mijson), true);
        
    if ($arrayJSON !='') {
//        loger($arrayJSON, 'json2');
        $nombresCampos 	= $arrayJSON[0];
        $tablaJSON 		= $arrayJSON[1];
        // Actualizo TablaDatos
        foreach($tablaJSON as $fila => $filaJSON) {
            // reemplazo el index numerico por los nombres de los campos
            foreach ($filaJSON as $num => $campoJSON) {
                $filaJSON[$nombresCampos[$num]] = $campoJSON;
                unset($filaJSON[$num]);
            }
            $MisDatos->TablaTemporal->updateFila($fila, $filaJSON );
  //      loger($fila, 'json2');
            
        }
        $MisDatos->calculointerno();
        Histrix_XmlReader::serializeContainer($MisDatos);

    }
    else {
    switch (json_last_error()) {
        case JSON_ERROR_NONE:
            $jsonError = ' - No errors';
        break;
        case JSON_ERROR_DEPTH:
            $jsonError =  ' - Maximum stack depth exceeded';
        break;
        case JSON_ERROR_STATE_MISMATCH:
            $jsonError =  ' - Underflow or the modes mismatch';
        break;
        case JSON_ERROR_CTRL_CHAR:
            $jsonError =  ' - Unexpected control character found';
        break;
        case JSON_ERROR_SYNTAX:
            $jsonError =  ' - Syntax error, malformed JSON';
        break;
        case JSON_ERROR_UTF8:
            $jsonError =  ' - Malformed UTF-8 characters, possibly incorrectly encoded';
        break;
        default:
            $jsonError =  ' - Unknown error';
        break;
    }
	loger($jsonError, 'error_json_decode')           ;
	loger($mijson, 'error_json_decode');
    
    }
}


$fila 		= (isset($_GET['fila'])) ? $_GET['fila'] :'';
$campo 		= (isset($_GET['campo']))? $_GET['campo']:'';
$delfiltros = (isset($_GET['delfiltros']))? $_GET['delfiltros']:'';
$valor 		= (isset($_GET['valor']))? $_GET['valor']:'';

if ($valor == 'true') $valor = 1;
if ($valor == 'false') $valor = 0;

$modificar 	= (isset($_GET['modificar']))? $_GET['modificar']:'';
if ($modificar == 'true'){
    $MisDatos->modificar =  'true';
    $show = false;
  //  $MisDatos->setInstance(); // Search ficha brokens
}
else $MisDatos->modificar =  'false';

$arraydato[$campo] = $valor;


$MisDatos->_setData = true;

// Just order Temporal Table
if (isset($_POST['newOrder']) && $_POST['newOrder'] != ''  ) {
    if ($MisDatos->tipoAbm != 'ing')
        $MisDatos->TablaTemporal->newOrder($_POST['newOrder']);

    $show=false;
    
} else {
/* lleno los datos del Objeto */
// si es por get solo updateo la tabla temporal
    if ($fila!= '') {
        $MisDatos->TablaTemporal->updateFila($fila, $arraydato );
    }
    else {
        $field = '';
        foreach($_POST as $dato => $val) {
            $val = urldecode($val);
            // si tengo que actualizar un subcontenedor
            if ($xmldatossub != '') {
                if (isset($MisDatos->CabeceraMov) && $MisDatos->CabeceraMov != '')
                    foreach($MisDatos->CabeceraMov as $ncab => $cab) {
                        if ($cab->xml == $xmldatossub) {
                            $tipo = $cab->getCampo($dato)->TipoDato;
                            if ($tipo == 'numeric') $val = mifloat($val);
                            if ($tipo == 'date') {
                                if (strlen($val)==8) {
                                    $date= strtotime(substr($val, 6, 4).'-'.substr($val, 3, 2).'-'.substr($val, 0, 2));
                                    $val = date('d/m/Y', $date);

                                }
                            }
                            $MisDatos->CabeceraMov[$ncab]->setFieldValue($dato, $val, 'both');
                            $MisDatosCab->setFieldValue($dato, $val, 'both');

                        }
                    }
            }
            else {
                if ($MisDatos != '') {

                    $objCampo = $MisDatos->getCampo($dato);
                    if ($objCampo) {
                        if ( ($objCampo != false) && $objCampo->TipoDato == 'numeric') $val = mifloat($val);

                        if (isset($_GET['blanqueo']) && $_GET['blanqueo'] !='') {
                            $ref_C= &$MisDatos->getCampoRef($dato);
                            $ref_C->valor = $val;
                        }
                        $MisDatos->setFieldValue($dato, $val, 'both');
                    }
                }
            }
            $field .= $dato;

        }
        
        if ($MisDatos->tipoAbm == 'ficha' || $MisDatos->tipoAbm == 'fichaing'){
            $MisDatos->CargoTablaTemporalDesdeCampos();
        }
        

        // Actualizo los contenedores Externos
        if (is_array($MisDatos->tablas[$MisDatos->TablaBase]->campos)) {
            foreach ($MisDatos->tablas[$MisDatos->TablaBase]->campos as $clavecampo => $Campo) {

                if (isset($Campo->contExterno)) {
                    $_Ref_Campo= &$MisDatos->getCampoRef($clavecampo);
                    if ($_Ref_Campo) {
                        if ($delfiltros == 'true') {
                            //2008-01-24
                            if ($_Ref_Campo->contExterno->xml != '') {

                                $tempOBJ = Histrix_XmlReader::unserializeContainer($_Ref_Campo->contExterno);

                                if ($tempOBJ != '') $_Ref_Campo->contExterno = $tempOBJ;
                            }

                            $campoObj = $MisDatos->getCampo($clavecampo);
                            if ($campoObj)
                                $campoObj->delCondiciones();
                        }

                        if (isset($_Ref_Campo->paring)) {
                            
                            if (isset($_Ref_Campo->contExterno->xml) && $_Ref_Campo->contExterno->xml != '') {
                                $tempOBJ = Histrix_XmlReader::unserializeContainer($_Ref_Campo->contExterno);

                                if ($tempOBJ != '') $_Ref_Campo->contExterno = $tempOBJ;
                            }

                            foreach ($_Ref_Campo->paring as $destinodelValor => $origendelValor) {
                                $campo =$MisDatos->getCampo($origendelValor['valor']);

                                if ($campo != ''){
                                    
                                    $valorDelCampo = $campo->getValor();
                                    $tipodatocampo = $campo->TipoDato;
                                    $operador= ($origendelValor['operador'])?$origendelValor['operador']:'=';

                                    // Not_So_Magic_Quotes Value
                                    $quotedValue = Types::getQuotedValue($valorDelCampo, $tipodatocampo, 'xsd:integer');
                                    if ($quotedValue =='') {
                                        $quotedValue = 0 ;
                                        $valorDelCampo=0;
                                    }

                                    if ($operador != 'null')
                                        $_Ref_Campo->contExterno->addCondicion($destinodelValor, $operador, $quotedValue, 'and', 'reemplazo', true);

                                    $_Ref_Campo->contExterno->setFieldValue($destinodelValor, $valorDelCampo, 'both');
                                    $field = $_Ref_Campo->contExterno->getCampo($destinodelValor);
                                    if ($field)
                                        $field->setValorOriginal($valorDelCampo);
                                }

                            }
                        }
                        if (isset($tempOBJ) && $tempOBJ != '') {
                            Histrix_XmlReader::serializeContainer($_Ref_Campo->contExterno);
                        }
                    }
                }
            }
        }
    }

}


if ($modificar != 'true' 
    //&& $fila == ''
    // && $show == true
    ){

    $MisDatos->calculointerno();
}

$MisDatos->_setData = false;

/* actualizo el objeto */

// update referent container
if (isset($MisDatos->parentInstance)) {
    $ParentContainer = Histrix_XmlReader::unserializeContainer('', $MisDatos->parentInstance);
    if (is_array($ParentContainer->tablas[$ParentContainer->TablaBase]->campos)) {
        foreach ($ParentContainer->tablas[$ParentContainer->TablaBase]->campos as $clavecampo => $Campo) {
            if (isset($Campo->contExterno) && $Campo->contExterno->getInstance() == $MisDatos->getInstance() ) {

                $Campo->contExterno = $MisDatos;

                Histrix_XmlReader::serializeContainer($ParentContainer);
            }
        }   
    }
}
// end update referent container


Histrix_XmlReader::serializeContainer($MisDatos);


if (isset($MisDatosCab) && $MisDatosCab != '') {

    Histrix_XmlReader::serializeContainer($MisDatosCab);
}

if ($show) {
    $mixml = new Cont2XML($MisDatos, $xmldatos, true, false);
    echo $mixml->show();
}
else echo '<end/>';

?>