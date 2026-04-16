<?php
/**
 * Search File form
 * by: Luis M. Melgratti
 * Revised: 2010-03-15
 */

include('./autoload.php');

$ArchivoConexion = "../funciones/conexion.php";
if (is_readable($ArchivoConexion))
    include($ArchivoConexion);
include ("./sessionCheck.php");
header('Content-Type: text/xml');


//$xmldatos = $_GET["xmldatos"];

$instance = $_REQUEST["instance"];

$MisDatos = new ContDatos("");
$MisDatos = Histrix_XmlReader::unserializeContainer(null, $instance);
$xmldatos = $MisDatos->xml;

if (!is_object($MisDatos)){
    $xmlFileName   = isset($_GET["xmldatos"])?$_GET["xmldatos"]:'';
    header('HTTP/1.1 400 Error: Container '.$xmlFileName.' not Found');
    echo $xmlFileName.' '.$instance;
    die();
}

$cant=0;

$xmlerror= '';
if (isset($_REQUEST['__help'])){
    $helpField  = $MisDatos->getCampo($_REQUEST['__help']);

    
    $claveAyuda = ($helpField->ClaveAyuda != '')? $helpField->ClaveAyuda:$_REQUEST['__help'];

    $MisDatos = $helpField->getContenedorAyuda();    
    $MisDatos->parentInstance = $instance;
    $xmlerror = ' error="'.$_REQUEST['__help'].'" ';

}

foreach($_POST as $postCampo => $postValor) {

    if (substr($postCampo, 0, 2) == '__') continue;
    
    if (isset($_REQUEST['__help'])){
        $postCampo = $claveAyuda;
    }

    $postValor = urldecode($postValor);
    $objCampo = $MisDatos->getCampo($postCampo);

    if ($objCampo!='') {
        $valorant= $objCampo->getValor();

        $objCampo->setValor($postValor);

        $MisDatos->getCampo($postCampo)->delCondiciones();
        if ($postValor != '' && ($valorant != $postValor || isset($_REQUEST['__help']))) {
            // si no tiene opcion seleccionada
            
            if	(isset($objCampo->opcion) && $objCampo->search != 'true') continue;

            //if ($objCampo->opcion[$valorant]=='') continue;
            if 	($objCampo->TipoDato =='check') continue;
            if	($objCampo->TipoDato =='date' ) continue;

            if( ($objCampo->TipoDato == 'varchar' ||
                $objCampo->TipoDato == 'longtext')  
                && !isset($_REQUEST['__help'])          // in help fields allways use = (forms)
                && !isset($_REQUEST['__searchField'])   // in forms fields allways use =
            ) {
                $valor = "'%".$postValor."%'";

                $op = 'like';

            }  else {
                $valor = "'".$postValor."'";
                $op = '=';

            }

            if (isset($_REQUEST['__help'])){
                $op = '=';
            }            

            $MisDatos->addCondicion($postCampo, $op, $valor, 'and','reemplazo', 'false');

            // fix to filter unions
            if ($MisDatos->unionContainers != ''){

                foreach($MisDatos->unionContainers as $UnionCont){
                    $UnionCont->addCondicion($postCampo, $op, $valor, 'and', 'reemplazo', 'false');
                }
            }


    //        die();
            $cant++;
        }
    }
}
// Solamente si hay algo
//if ($cant <> 0) 
//{

$rsPrivado =$MisDatos->Select();
$numfilas  = _num_rows($rsPrivado);
$modo = false;

/* Si se devuelve mas de un registro serializo el Objeto con prefijo _aux_ para mostrarlo en una consulta por
	 * pantalla posteriormente;
*/
if ($numfilas == 0) {
    /* si no hay coincidencias */
    if ($_GET['ayudaFicha']=='true') $ayudaFicha='ayudaficha="true" ';
    echo '<?xml version="1.0" encoding="utf-8" ?><resultado '.$ayudaFicha.$xmlerror.' vacio="true"  parentInstance="'.$instance.'" ></resultado>';
    return;
} else {
    if ($numfilas > 1  && !isset($_REQUEST['__help']) ) {

          
        $modo = true;
        $nomaux = '__aux_'.$xmldatos;
        $MisDatos->_instance = '_aux_'.$MisDatos->getInstance();
        Histrix_XmlReader::serializeContainer($MisDatos);

    } else {

        /* recorro los campos */
        if ($rsPrivado) {
            $row = _fetch_array($rsPrivado);

            if (($row)) {

                foreach ($row as $clarow => $valrow) {
                    if ($MisDatos->getCampo($clarow)) {
                        $MisDatos->getCampo($clarow)->setValor($valrow);
                    }
                }
                // Actualizo los contenedores Externos
                foreach ($row as $clarow => $valrow) {
                    $_Ref_Campo= $MisDatos->getCampo($clarow);
                    if ($_Ref_Campo) {
                        if(isset($_Ref_Campo->contExterno) && $_Ref_Campo->contExterno !='') {

                            if ($_Ref_Campo->paring !='') {
                                foreach ($_Ref_Campo->paring as $destinodelValor => $origendelValor) {
                              	    $field = $MisDatos->getCampo($origendelValor['valor']);
                              	    if (is_object($field)){
                                      $valorDelCampo = $field->getValor();
	                                  $tipodatocampo = $field->TipoDato;

  	                                  // Not_So_Magic_Quotes Value
    	                              $quotedValue = Types::getQuotedValue($valorDelCampo, $tipodatocampo, 'xsd:integer');
      	                              if ($quotedValue =='') {
        	                              $quotedValue = 0 ;
          	                              $valorDelCampo=0;
            	                      }
              	                      $_Ref_Campo->contExterno->addCondicion($destinodelValor, '=', $quotedValue, 'and', 'reemplazo', true);
                	                  $_Ref_Campo->contExterno->setCampo($destinodelValor, $valorDelCampo);
                  	                  $_Ref_Campo->contExterno->setNuevoValorCampo($destinodelValor, $valorDelCampo);
                  	                }
                                }
                            }
                            //	loger($_Ref_Campo->contExterno->getSelect());
                            Histrix_XmlReader::serializeContainer($_Ref_Campo->contExterno);
                            
                        }

                    }
                }

            }
        }
    }

    $MisDatos->CargoTablaTemporalDesdeCampos();
    $MisDatos->CargoCamposDesdeTablaTemporal();


    if (isset($_REQUEST['__help'])) {

        $mixml = new Cont2XML($MisDatos, $xmldatos, null, false, true, true, $_REQUEST['__help']);

    } else {
        
        $mixml = new Cont2XML($MisDatos, $xmldatos, true, $modo, true, false);

    }
    
    echo  $mixml->show();

    if ($numfilas < 1) {

        //16-07-2008
        // Borro las Condiciones de Busqueda de los campos
        // Para evitar que se acumulen
        foreach($_POST as $postCampo => $postValor) {
            $objCampo = $MisDatos->getCampo($postCampo);

            if ($objCampo!='') {
                $valorant= $objCampo->getValor();
                $MisDatos->getCampo($postCampo)->delCondiciones();
            }
        }

        Histrix_XmlReader::serializeContainer($MisDatos);
    }

}
//}
?>
