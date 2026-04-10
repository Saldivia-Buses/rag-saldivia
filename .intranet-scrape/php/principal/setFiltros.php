<?php
/*
 * Created on 30/12/2005
 *
*/
include ("./autoload.php");
include_once ("../funciones/conexion.php");
include ("./sessionCheck.php");

class ObjetoInput {
    var $NombreCampo;
    var $Operador;
    var $Valor;

    public function ObjetoInput($NomCampo='', $operador='' ,$Val='') {
        $this->NombreCampo = $NomCampo;
        $this->Operador    = $operador;
        $this->Valor		  = $Val;

    }

}

$mijson = $_POST['mijson'];

$contents = utf8_encode($mijson); 
$arrayJSON= json_decode($contents, true);

if ($arrayJSON == ''){
    // fix for broken json
    $mijson = stripslashes($mijson);
    $contents = utf8_encode($mijson);
    $arrayJSON= json_decode($contents, true);

}
    /*
    switch(json_last_error())
    {
        case JSON_ERROR_DEPTH:
            echo ' - Excedido tamaño máximo de la pila';
        break;
        case JSON_ERROR_CTRL_CHAR:
            echo ' - Encontrado carácter de control no esperado';
        break;
        case JSON_ERROR_SYNTAX:
            echo ' - Error de sintaxis, JSON mal formado';

        break;
        case JSON_ERROR_NONE:
        echo ' - Sin errores';
        break;
    }
     * 
     */

/////////////////////////////////////////////////////
// ATENTION
// WARNING!!!!!!!!!!!!!
// watchout for magic quotes in php, they must be off <--
/////////////////////////////////////////////////////

$i = 0;
if ($arrayJSON !='')
    foreach ($arrayJSON as $clave => $val) {
        $i++;

        if (!is_array($val[2])) {
            $utfValue = utf8_decode($val[2]);
        }
        else {
            $utfValue = $val[2];
        }
        

        $arrayCondiciones[$i] = new ObjetoInput($val[0], $val[1], $utfValue);
    }


if ($_POST['variable'] !='') {

    $arrayCondicionesOrig[0]='';
    $arrayCondicionesOrig = unserialize(urldecode(stripslashes($_POST['variable'])));
    $i = 0;
    if ($arrayCondicionesOrig !='')
        foreach ($arrayCondicionesOrig as $clave => $val) {
            $i++;
            $utfValue =  utf8_encode($val[3]);
            //$utfValue =  $val[3];
            
//                                loger($utfValue, 'utf');
            $arrayCondiciones[$i] = new ObjetoInput($val[1], $val[2], $utfValue);
        }
}



$MisDatos = new ContDatos("");
$MisDatos = Histrix_XmlReader::unserializeContainer(null, $_REQUEST['instance']);

unset($MisDatos->hasValue);
    

// Borro todas las condiciones del select;
$MisDatos->delCondiciones();

unset($MisDatos->grupoJoin);

if ($arrayCondiciones !='')
    foreach ($arrayCondiciones as $clave => $val) {
        $valor = '';
        $nom 		=  $val->NombreCampo;
        $operador 	=  $val->Operador;
        $valor		=  $val->Valor;
	
		$int = false;
		if (is_array($valor)){
//		  $valor = implode(',',$valor);
		  $int = true;
		}
  
        $ObjCampo   =  $MisDatos->getCampo($nom);
        if ($ObjCampo) {
            $MisDatos->grupoJoin[$ObjCampo->grupo]   = 'true';

            if ($operador == '') $operador ='=';
            if ($operador == '=') $reemplazo = 'reemplazo';
            $oplogico = ' and ';
            if ($ObjCampo->oplogico != '') $oplogico = ' '.$ObjCampo->oplogico.' ';

            if ($ObjCampo->oplogico == 'or') $reemplazo = '';
            $tipo = Types::getTypeXSD($ObjCampo->TipoDato, 'xsd:integer');
            if ($valor != '') {
                
                if ($tipo =='xsd:integer' || $tipo =='xsd:decimal' || $int == true) {
                    $filterError[$nom] = true;
                    $filterError[$nom] = $MisDatos->addCondicion($nom, $operador, $valor, $oplogico, $reemplazo, 'true');
                    $errorValues[$nom] = $valor;
                }
                else {
                    $comillasini= '';
                    $comillasfin= '';

                    if ($tipo == 'xsd:string') {
                        $comillasini = "'";
                        $comillasfin = "'";
                    }
                    if ($operador == 'like') {
                        $comillasini = "'%";
                        $comillasfin = "%'";

                    }
                    //   echo $nom, $oplogico, $reemplazo, $valor;

                    $MisDatos->addCondicion($nom, $operador, $comillasini.$valor.$comillasfin, $oplogico, $reemplazo, 'false');
        	    if ($MisDatos->unionContainers != ''){
        		foreach($MisDatos->unionContainers as $UnionCont){
        		    $UnionCont->addCondicion($nom, $operador, $comillasini.$valor.$comillasfin, $oplogico, $reemplazo, 'false');
        		}
        	    }
        	}
            }


            // actualizo el objeto filtro
            if ($MisDatos->filtros)
                foreach ($MisDatos->filtros as $nomfiltro => $objFiltro) {
                    if ($objFiltro->campo == $nom && $objFiltro->operador == $operador)
                        $objFiltro->valor = $valor;
                }

            //$MisDatos->getCampo($nom)->valor=$valor;

            $MisDatos->setCampo($nom, $valor);
            $MisDatos->setNuevoValorCampo($nom, $valor);

        }
    }

$MisDatos->llenoTemporal ='true';
$UI = 'UI_'.str_replace('-', '', $MisDatos->tipo);
$datos = new $UI($MisDatos);

$die = false;
if ($filterError != '')
    foreach ($filterError as $fieldName => $value) {
        if ($value == false) {
            $die = 'noSelect';
            $ObjCampo   =  $MisDatos->getCampo($fieldName);
            echo '<div class="error">';
            echo $datos->i18n['error'].': "'.$ObjCampo->Etiqueta.'" '.$datos->i18n['wrongValue'].': "'.$errorValues[$fieldName].'" <br>';
            echo '</div>';
        }
    }

//echo $MisDatos->getSelect();
$tablaInterna = $datos->showTablaInt(null, $MisDatos->idxml, $die);

if($datos->tipo != 'chart')
    echo $tablaInterna;
if ($MisDatos->grafico != '') {
    $reloadgraf .='<script type="text/javascript">';
    foreach ($MisDatos->grafico as $id_grafico => $grafico) {
        $reloadgraf .="reloadImg('$id_grafico');";
        $reloadgraf .="\n";
    }
    $reloadgraf .='</script>';
    echo $reloadgraf;
}
Histrix_XmlReader::serializeContainer($MisDatos);



?>