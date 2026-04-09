<?php


/*
 * Created on 07/11/2005
 * Luis M. Melgratti
 * Perform Actions in Data Containers
 * kind of action bootstrap
 */

require ("./autoload.php");
require ("../funciones/utiles.php");

ob_start();

$ArchivoConexion = "../funciones/conexion.php";
if (is_readable($ArchivoConexion))
    include_once($ArchivoConexion);

include ("./sessionCheck.php");

set_time_limit(200);

////////////////////////////////////////////////////////////////////////////
// get $_GET and $_POST variables
////////////////////////////////////////////////////////////////////////////

//$xmldatos    = $_GET["xmldatos"]; // mandatory - replaced by instance


//loger('proceso'.$xmldatos, 'process.log');


// process event flag
$processEvent      = (isset($_GET["_pe"])) ?       $_GET["_pe"]     : 'true';

$accion      = (isset($_GET["accion"])) ?       $_GET["accion"]     : '';
$orden       = (isset($_GET["orden"])) ?        $_GET["orden"]      : '';
$contenedor  = (isset($_GET["contenedor"])) ?   $_GET["contenedor"] : '';

$cadena      = (isset($_REQUEST["cadena"])) ?      $_REQUEST["cadena"]    : '';

$esficha     = (isset($_GET["Ficha"])) ?        $_GET["Ficha"]      : '';
$nosel       = (isset($_GET["nosel"])) ?        $_GET["nosel"]      : '';

$rowaborrar  = (isset($_POST['rowaborrar'])) ? $_POST["rowaborrar"] : '';

// Valores para el filtro
$filtro      = (isset($_GET["filtro"])) ? trim($_GET["filtro"])     : '';
$delfiltros  = (isset($_GET["delfiltros"])) ? $_GET["delfiltros"]   : '';  // si borra TODAS las condiciones
$del_filtro  = (isset($_GET["del_filtro"])) ? $_GET["del_filtro"]   : ''; // si borra 1 CONDICION
//
//

// Autofiltros
$addfiltro      = (isset($_GET["addfiltro"])) ?      $_GET["addfiltro"] : '';
$autocampo      = (isset($_GET["autocampo"])) ?      $_GET["autocampo"] : '';
$autooperador   = (isset($_GET["autooperador"])) ?   $_GET["autooperador"] : '';

$removerfiltro  = (isset($_GET["removerfiltro"])) ?  $_GET["removerfiltro"] : '';
$uidfiltro      = (isset($_GET["uidfiltro"])) ?      $_GET["uidfiltro"] : '';

$valor          = (isset($_GET["valor"])) ?          $_GET["valor"] : '';
$operador       = (isset($_GET["operador"])) ?       $_GET["operador"] : '=';
$reemplazo      = (isset($_GET["reemplazo"])) ?      $_GET["reemplazo"] : 'reemplazo';

if ($operador == '=')
    $reemplazo = 'reemplazo';

$ejecutar       = (isset($_GET["ejecutar"])) ?       $_GET["ejecutar"] : '';

$forzar         = (isset($_GET["forzar"])) ?         $_GET["forzar"] : '';

$nrofila        = (isset($_POST['Nro_Fila'])) ?      $_POST['Nro_Fila'] : '';

$_fila          = (isset($_GET['fila'])) ?          $_GET['fila']  : '';
$_campo         = (isset($_GET['campo'])) ?         $_GET['campo'] : '';
$valor          = (isset($_GET['valor'])) ?         $_GET['valor'] : '';
$export         = (isset($_GET['export'])) ?        $_GET['export'] : '';

$xmlOrig        = (isset($_REQUEST['xmlOrig'])) ?   $_REQUEST['xmlOrig'] : '';


if ($xmlOrig == 'undefined')
    $xmlOrig = ''; // force blank context where js culdnt find it


if (substr($xmlOrig, 0, 4) == 'Form')
    $xmlOrig = substr($xmlOrig, 4);
if (substr($xmlOrig, 0, 5) == 'FForm')
    $xmlOrig = substr($xmlOrig, 5);


///////////////////////////
// Create Data Container
//////////////////////////

$MisDatos = new ContDatos("");
$MisDatos = Histrix_XmlReader::unserializeContainer(null, $_REQUEST['instance']);

if (!is_object($MisDatos)){
    loger($_REQUEST, 'cont_vacio.log');

    header('HTTP/1.1 400 Bad Request');
    die('Process Error');
}

$xmldatos = $MisDatos->xml;    


///////////////////////////
// Export
/////////////////////////
if ($export != ''){
    $exportFormat = 'Export_'.$export;
    $DataExport   = new $exportFormat($MisDatos, $_GET['titulo']);
    $DataExport->export();
    die();
}




//loger($MisDatos, $MisDatos->xml);
/*
$Tablatemp = $MisDatos->TablaTemporal->datos();
loger($Tablatemp, $_REQUEST['instance'] );
loger($_REQUEST, $_REQUEST['instance'] );
*/
$tipo = $MisDatos->tipoAbm;

$salida = '';


///////////////////////////
// liveGrid processing
///////////////////////////

if (isset($_POST['oldValues'])) {
    $MisDatos->checkDupOnUpdate = true;
    parse_str($_POST['oldValues'], $oldValues);
    foreach ($oldValues as $fieldName => $value) {

        $MisDatos->setFieldValue($fieldName, $value);
    }
}

if (isset($_POST['newValues'])) {
    $MisDatos->checkDupOnUpdate = true;
    parse_str($_POST['newValues'], $newValues);
    foreach ($newValues as $fieldName => $value) {
        $MisDatos->setNuevoValorCampo($fieldName, $value);
    }
}


/////////////////////////
// Add new ROW
/////////////////////////

if (isset($_POST['addRow'])) {
    $data = Array();
    foreach ($MisDatos->tablas[$MisDatos->TablaBase]->campos as $MiNro => $ObjCampoF) {
        $data[$ObjCampoF->NombreCampo]  = $ObjCampoF->valor;
    } 

    $MisDatos->TablaTemporal->insert($data, $MisDatos->autoUpdateRow);
   // loger($data, 'updatessql.log');
}



/////////////////////////
// Autofilters
/////////////////////////
// add Autofilters
if ($addfiltro != '') {
//    $MisDatos->getCampo($autocampo)->oplogico = ' and ';
    $label = $MisDatos->getCampo($autocampo)->Etiqueta;
    $MisDatos->addFiltro($autocampo, $autooperador, $label, '', 'auto', null, null, null, null);
}
/////////////////////////
// Remove Autofilters
/////////////////////////
if ($removerfiltro != '') {
    $MisDatos->delCondiciones();
    foreach ($MisDatos->filtros as $nomfiltro => $objFiltro) {
        if ($objFiltro->uid == $uidfiltro)
            unset($MisDatos->filtros[$nomfiltro]);
    }
}


/////////////////////////
// For Help containers I use a CLoned object to not mess with original DataContainer
/////////////////////////
if ($accion == 'help' && $MisDatos != '') {
    $MisDatosHelp = clone $MisDatos;
    $MisDatosHelp->xmlOrig = $xmlOrig;
}

if ($xmlOrig != '')
    $MisDatos->xmlOrig = $xmlOrig;

// Force a particular Type for the xml
if ($forzar != '') {
    $MisDatos->forzado = true;
    $MisDatos->tipoAbmOrig = $MisDatos->tipoAbm;
    $MisDatos->tipoAbm = $forzar;
} else {
    // Restore original Type
    if (isset($MisDatos->tipoAbmOrig)) {
        $MisDatos->tipoAbm = $MisDatos->tipoAbmOrig;
    }
}



if ($accion == 'insupdate') {
    if ($MisDatos->modificar == 'true')
        $accion = 'update';
    else
        $accion = 'insert';
}
// restauro la bandera
$MisDatos->modificar = 'false';


if (isset($_GET['autocomplete']) && $_GET['autocomplete'] == 'true') {
    $autocomplete = true;
    $cadena = $_REQUEST['q'];
    
}
else
    $autocomplete = false;

// destroy object

$destroy = false;





/* lleno los datos del Objeto */
// si es por get solo updateo la tabla temporal
if ($_fila != '') {

    if ($valor == 'true')
        $valor = 1;
    if ($valor == 'false')
        $valor = 0;

    $arraydato[$_campo] = $valor;
    $MisDatos->TablaTemporal->updateFila($_fila, $arraydato);
    
    //$MisDatos->calculointerno($_fila);
    $MisDatos->calculointernoFila($MisDatos->TablaTemporal->Tabla[$_fila], $_fila, 0);
    
    //
    //
    //$MisDatos->calculointerno($_fila); //no calculo todo??? ver
}

if ($delfiltros != '') {
    $MisDatos->delCondiciones();
}

if ($del_filtro != '') {
    unset($MisDatos->getCampo($del_filtro)->Condiciones);
}



/////////////////////////////////////////
// Search Strings in Help
/////////////////////////////////////////
$MisDatos->filterString($cadena);

/////////////////////////////
// Apply Filter
/////////////////////////////

if ($filtro != '') {

    $ObjCampoF = $MisDatos->getCampo($filtro);

    $tipoDato = Types::getTypeXSD($ObjCampoF->TipoDato, 'xsd:integer');

    if ($tipoDato == 'xsd:integer' || $tipoDato == 'xsd:decimal' ||
            strpos($ObjCampoF->TipoDato, 'integer') !== false) {
        $MisDatos->addCondicion($filtro, $operador, $valor, ' and ', $reemplazo);
    } else {
        $MisDatos->addCondicion($filtro, $operador, "'" . $valor . "'", ' and ', $reemplazo);
    }

    $MisDatos->setFieldValue($filtro, $valor, 'both');
//    $MisDatos->setCampo($filtro, $valor);
//    $MisDatos->setNuevoValorCampo($filtro, $valor);
    // actualizo el objeto filtro
    if (isset($MisDatos->filtros)) {
        foreach ($MisDatos->filtros as $nomfiltro => $objFiltro) {
            if ($objFiltro->campo == $filtro && $objFiltro->operador == $operador)
                $objFiltro->valor = $valor;
        }
    }
}

        

////////////////////////////////////////////////
// Fill Container from POST data
///////////////////////////////////////////
//loger('obtengo post Data'.$xmldatos, 'process.log');

if (isset($MisDatos->CabeceraMov) && ($accion == 'procesar')){
    $Cabecera = current($MisDatos->CabeceraMov);
    $Cabecera = Histrix_XmlReader::unserializeContainer(null, $Cabecera->getInstance());
}


$Update = false;
$ObjCampo = '';
foreach ($_POST as $dato => $lista) {

    // skip local variables
    if (substr($dato, 0, 2) == '__')
        continue;
    unset($ObjCampo);
    $ObjCampo = $MisDatos->getCampo($dato);

    // Inner Tables
    if (is_array($lista)) {

//        loger($xmldatos.'inner Table '.$dato, 'process.log');

        $innerTable = new TablaDatos($dato);
        $innerTable->insert($lista);
        $ObjCampo->innerTable = $innerTable;
    }

    $val = stripslashes(urldecode($_POST[$dato]));
    if (is_object($ObjCampo)) {

        $xsdType = Types::getTypeXSD($ObjCampo->TipoDato);
        if (isset($ObjCampo->checkType) && $ObjCampo->checkType != 'false')
            switch ($xsdType) {
                case "xsd:decimal" :
                    $val = mifloat($val);
                    break;
                case "xsd:integer" :
                    $val = intval($val);
                    break;
                case "xsd:date" :
                    if (strlen($val) == 8) {
                        $date = strtotime(substr($val, 6, 4) . '-'
                                        . substr($val, 3, 2) . '-'
                                        . substr($val, 0, 2));
                        $val = date('d/m/Y', $date);
                    }
                    break;
            }
    }
    $MisDatos->setFieldValue($dato, $val, 'new');

    if ($MisDatos->tipoAbm == 'fichaing') {
        $MisDatos->setFieldValue($dato, $val, 'both');
        //$MisDatos->setCampo($dato, $val);
    }

    if ($accion == 'procesar') {
        if (isset($Cabecera)) {
            //foreach ($MisDatos->CabeceraMov as $NCab => $Cabecera) {
                $ObjCampo2 = $Cabecera->getCampo($dato);

                if ($ObjCampo2->TipoDato == 'date') {
                    if (strlen($val) == 8) {
                        $date = strtotime(substr($val, 6, 4) . '-'
                                        . substr($val, 3, 2) . '-'
                                        . substr($val, 0, 2));
                        $val = date('d/m/Y', $date);
                    }
                }
              //  echo $dato.' = '.$val.'|';
                $Cabecera->setFieldValue($dato, $val, 'both');
           // }
        }
    }                                 
    $Update = true;
}

if(isset($Cabecera) && ($accion == 'procesar')){
    Histrix_XmlReader::serializeContainer($Cabecera);
    $MisDatos->addCabecera($Cabecera);
}

//loger('fin obtengo post Data '.$xmldatos, 'process.log');

if ($MisDatos->tipoAbm == 'fichaing' ) {
    //loger('ini calculo interno'.$xmldatos, 'process.log');//    
    $MisDatos->calculointerno();
    $MisDatos->CargoTablaTemporalDesdeCampos(true);
    //$Update = false;
    //loger('fin calculo interno'.$xmldatos, 'process.log');//
}

// by default we redraw the current XML
$redrawXml = true;

if ($ejecutar != 'no') {

    switch ($accion) {

        case "refresh":
                $MisDatos->modificar = 'true'; // makes ficha update button stays as update
                $MisDatos->restaurarValores();

        break;
        case "update" :
            if ($Update) {

                $process = $MisDatos->Update($nrofila);
                if ($process === -1) {
          //          header('HTTP/1.1 400 Bad Request');
                    die();
                }
                if ($MisDatos->_menuId != '') {
                    $_SESSION['_menuId'] = $MisDatos->_menuId;
                    $MisDatos->Notify('Modificó ');
                }

                /* restaura los valores al Select Original que se forma con el AbmGenerico
                 * Para que el select posterior sea igual (se utiliza en los Arboles)
                 * */
                //loger('hago el update '.$MisDatos->tipoAbm );
                // if ($nrofila != '' || $MisDatos->tipoAbm == 'arbol') {

                $MisDatos->modificar = 'true'; // makes ficha update button stays as update
                $MisDatos->restaurarValores();
                
                // }

		if ($processEvent == 'true')
                  $reload = getActionEvents($accion, $MisDatos, $nrofila + 1);
                
                
            }
            break;
        case "delete" :
            $MisDatos->Delete($rowaborrar);

            if ($MisDatos->_menuId != '') {
                $_SESSION['_menuId'] = $MisDatos->_menuId;
                $MisDatos->Notify('Borrado ');
            }
            //  if ($nrofila != '' || $MisDatos->tipoAbm == 'arbol') {
            $MisDatos->restaurarValores();
            //    }
            
            
            if ($processEvent == 'true')
          	      	$reload = getActionEvents($accion, $MisDatos, $rowaborrar);
                    

          
            break;
        case "insert" :

            if ($Update){
                $autoinc = $MisDatos->Insert();
            }

            if ($autoinc === -1) {
               // header('HTTP/1.1 400 Bad Request');
                die();
            }
            // Insert Data a into inner TablaData cell
            $MisDatos->modificar = 'true';
            // Si hay un autonumerico busco el registro creado y lo obtengo

            if ($autoinc != 0 && ( $tipo == 'ficha' || $tipo == 'liveGrid')) {
                foreach ($MisDatos->tablas[$MisDatos->TablaBase]->campos as $MiNro => $ObjCampoAuto) {
                    $nomcamp = $ObjCampoAuto->NombreCampo;

                    if (isset($ObjCampoAuto->autoinc) && $ObjCampoAuto->autoinc == 'true') {

                        $reemplazo = 'reemplazo';
                        $operador = '=';
                        $valor = $autoinc;
                                          loger($nomcamp.'=', 'auto');
                        $MisDatos->setFieldValue($nomcamp, $autoinc, 'both');
                        $MisDatos->addCondicion($nomcamp, $operador, $valor, ' and ', $reemplazo);
                        if ($tipo == 'liveGrid') {

                            $MisDatos->Select();
                        }
                    }
                }
            }

            if (isset($MisDatos->_menuId) && $MisDatos->_menuId != '') {
                $_SESSION['_menuId'] = $MisDatos->_menuId;
                $MisDatos->Notify('Insertó ');
            }


            if ($MisDatos->graboasiento ) {
                $MisDatos->calculointerno();
            }               


            $MisDatos->restaurarValores();

            if ($processEvent == 'true')
                $reload = getActionEvents($accion, $MisDatos, $autoinc );
                     

            break;
        case "procesar" :
        case "process" :        
	       $accion = 'procesar';

           if ($MisDatos->saveState == "true"){
                $MisDatos->llenoTemporal = 'false';
	            $MisDatos->__savedState  = true;
                $MisDatos->saveState();
                $redrawXml = false;

           } else {
//               loger('ini Grabar '.$xmldatos, 'process.log');



                if ($MisDatos->tipoAbm == 'ing') {
                    $MisDatos->calculointerno();
                }               


                $process = $MisDatos->GrabarRegistros();

                unset($MisDatos->cadenasSQL);
                if ($process === -1) {
                  //  header('HTTP/1.1 400 Bad Request');
                    //ECHO '<div class="error boton" style="text-align:center;font-size:20px;" onclick="$(this).parent(\'div\').remove();$(\'.modalWindow\').remove();">ERROR EN LA GRABACION</div>';
                    die();
                }

                loger('fin Grabacion'.$xmldatos, 'process.log');//

                if ($MisDatos->_menuId != '') {
                    $MisDatos->Notify('Procesar ');
                }

           }
        

           if ($MisDatos->llenoreferente == 'true'){
                $redrawXml = false;
           }

            $redrawXml = false;

         //   $destroy= true;

	    if ($processEvent == 'true'){
              $reload = getActionEvents($accion, $MisDatos);
	    }


            if ($MisDatos->eventosXML[$accion] != '') {

                foreach ($MisDatos->eventosXML[$accion] as $nevent => $event) {
                    if ($nevent == 'close') {
                        $reload[] = "Histrix.closeTab('DIV" . $MisDatos->idxml . "', '{$MisDatos->xmlOrig}' );";
                        
                   //     $reload[] = "cerrarVent('PRNDIV" . $MisDatos->idxml . $MisDatos->xmlOrig . "');";

                        $redrawXml = false;
                    }

                }

            }
            

            break;
    }


    if (($orden))
        $MisDatos->setOrden($orden);

    //remove this for posible code duplication

    if ($accion == 'help') {
        $MisDatos->tipo = 'ayuda'; // change type
        $MisDatosHelp->tipo = 'ayuda'; // change type    

      //  Histrix_XmlReader::serializeContainer($MisDatosHelp, $xmlOrig);
    } else {
       // Histrix_XmlReader::serializeContainer($MisDatos, $MisDatos->xmlOrig);
    }




    // create apropiate graficalInterface
    // if ($MisDatos->tipo== '')$MisDatos->tipo = 'consulta';

    $UI = trim('UI_' . str_replace('-', '', $MisDatos->tipo));
    //loger($reload);
    try {
        $datos = new $UI($MisDatos);
    } catch (Exception $ex) {
        loger($ex, 'excep');
    }

    $tipo =  $MisDatos->tipoAbm;
    $datos->tipo = $tipo;
    $datos->setTitulo($MisDatos->tituloAbm);

    $datos->nosel = $nosel;  // no hace select

    // do not redraw if window will close
    // 
    // 

    if ($redrawXml) {

        if ($esficha == true || $tipo == 'ficha') {
            if ($accion == 'delete') {
                if ($MisDatos->nosql != 'true' /* && $accion != 'insert' */)

                    $MisDatos->Select();
            }
            else {
                $MisDatos->CargoTablaTemporalDesdeCampos(true);
                $MisDatos->CargoCamposDesdeTablaTemporal();
                unset($MisDatos->resultSet); // borro el resultSet
                $datos->preFetch = true;
            }

            $salida = $datos->showAbmInt('valores', 'INT' . $MisDatos->xml);
        } else {

            /////////////////////////
            //// CALENDAR Events
            /////////////////////////
            if (isset($_GET['start']) && isset($_GET['end']) && $tipo == 'calendar' || $tipo == 'map'){
		header('Content-Type: application/json; charset=utf-8');
                echo $datos->createEvents($_GET['start'], $_GET['end']);
                die();

            }
            else
            if ($tipo == 'arbol') {

                /* Al hacer el UPDATE de un arbol, debo setear el valor inicial para el nuevo Select
                 *  y volver a la Raiz inicial
                 */
                /*
                  $MisDatos->addCondicion($filtro, $operador, "'".$valor."'", ' and ', $reemplazo);
                  $MisDatos->addCondicion($MisDatos->, '=', "'".$valor."'", ' and ', $reemplazo);
                 */
                echo $datos->showTablaInt();
            } else {
                $opt = '';
                $act = '';
                if ($tipo == 'ing' || $tipo == 'grid' || $tipo == 'fichaing') {
                    $opt = 'NoSql';
                    $act = '2davez';
                    $form = 'Form' . $MisDatos->idxml;
                    $reload[] = "foco('$form');";

                }

                if ($accion == 'help') {
                    $datos->esAyuda = true;

                    $datos->campoRetorno = $_GET['idinput'];
                    $MisDatos->campoRetorno = $_GET['idinput'];

                    if ($autocomplete) {
                        $datos->autocomplete = true;
                        $salida .= $datos->showTablaInt($opt, '', $act, null, null, null, null, $MisDatos);
                    }
                    else{
               	
                        $salida .= $datos->show('', $_GET['divcont'], $opt);
                    }
                } else {

                    if ($filtro != '') {
                        $act = '';
                        $opt = '';
                    }

                    if ($accion == 'autoupdate') {
                        $MisDatos->Select('nolog');
                        $datos->cantCampos = _num_fields($MisDatos->resultSet);
                        $MisDatos->CargoTablaTemporal();
                        $xml = (isset($datos->xml)) ? $datos->xml : '';
                        $salida .= $datos->showDatos($xml, '');
                        if ($MisDatos->xml != '')
                            $reload[] = 'Histrix.registroEventos(\'' . $MisDatos->xml . '\')';
                    }
                    else {


                        if ($addfiltro != '')
                            echo $datos->showFiltrosXML('nodiv');
                        else {
                            if (isset($_GET['getFila']) && $_GET['getFila'] == 'true') {
                                // solo refresco 1 fila
                                if ($_GET['accion'] == 'insert') {
                                    // Just retrive row without TR information
                                    $datos->showTablaInt();
                                    $options = 'noRowInformation';
                                    $salida = $datos->showDatos(null, $options, $_fila);
                                } else {

                                    $salida = $datos->showDatos(null, null, $_fila);
                                }
                                if ($tipo != 'ing' || $MisDatos->updateTotals == 'true') {
                                    $upTotals = $datos->updateTotals();
                                    $salida .= $upTotals;
                                }
                                unset($reload);
                            } else {

                                if ($MisDatos->forceReload == 'true') {
                                    $MisDatos->Select('nolog');
                                    $datos->cantCampos = _num_fields($MisDatos->resultSet);
                                    $MisDatos->CargoTablaTemporal();
                                }
                                $salida .= $datos->showTablaInt($opt, $MisDatos->idxml, $act, null, null, null, null, $MisDatos);

                            }
                        }
                    }
                }
            }
        }
    }
    else {

        $salida .= '<div>Process..</div>';

    }
}




///////////////////////////////////////
// Asientos
///////////////////////////////////////

if (isset($MisDatos->graboasiento) && $MisDatos->graboasiento == "true" && $MisDatos->muestraasiento != 'false') {
    $id = uniqid('asiento' . $MisDatos->idxml);
    echo '<script type="text/javascript">cerrarVentclase(\'asiento\',\'' . $id . '\' );</script>';
    $salida .= '<div class="asiento" id="' . $id . '">';
//    $salida .= UI::barraDrag2($id, 'Minuta Contable');
    $salida .= '<div class="contewin">';
    $salida .= $MisDatos->Minuta->show();
    $salida .= '</div></div>';
    $drag[] = "$('#$id').draggable({
            handle: '#dragbar$id'});";

    $salida .= Html::scriptTag($drag);
}

///////////////////////////////////////
// Reload Graphics
///////////////////////////////////////
 
if (isset($MisDatos->grafico) && $MisDatos->grafico != '') {
    foreach ($MisDatos->grafico as $id_grafico => $grafico) {
        $reload[] = "reloadImg('$id_grafico');";
    }
}

////////////////////////////////////////
// Close Process
////////////////////////////////////////

if (isset($MisDatos->cerrar_proceso) && $MisDatos->cerrar_proceso == 'true' && $accion == 'procesar') {
    $titulo = 'procesando';
    $subdir = '&dir=' . $MisDatos->subdir;
    unset($reload);
    // commented out on 2011-03-10  seems redundant.
    //  $reload[]= "xmlLoader('DIV$MisDatos->xmlOrig', '&xml=$MisDatos->xmlOrig$subdir', {title:'$titulo', reload:true}); ";
}
else
    echo $salida;

/////////////////////////////
// free memory
/////////////////////////////
unset($salida);


if ($destroy){
    $MisDatos->destroy();

} else {


    //////////////////////////////////////
    // Serialize Containers
    /////////////////////////////////////

    //loger('ini serialize', 'process.log');//

    if ($accion == 'help') {
    // Si es una Ayuda serializo la copia que guarde antes
        $MisDatos->tipo = 'ayuda'; // change type    
        $MisDatosHelp->tipo = 'ayuda'; // change type    
        Histrix_XmlReader::serializeContainer($MisDatosHelp);
    } else {

        Histrix_XmlReader::serializeContainer($MisDatos);
    }

    //loger('fin serialize', 'process.log');//

}

////////////////////////////////////////
// Reload Events
////////////////////////////////////////

if (isset($reload)) {
/*
loger('processEvent'.$processEvent, '__pe');
loger($reload, '__pe');
loger('fin'.$processEvent, '__pe');
*/
    echo Html::scriptTag(array_unique($reload));
}


// Get action events
function getActionEvents($eventName, &$MisDatos, $rowNumber=null){
    if (isset($MisDatos->eventosXML[$eventName])){
        foreach ($MisDatos->eventosXML[$eventName] as $nevent => $event) {
            switch ($nevent) {
                case 'close':
                    $reload[] = "cerrarVent('PRN" . $MisDatos->idxml . "');";
                    $reload[] = "cerrarVent('PRNDIV" . $MisDatos->idxml . "');";
                    $reload[] = "cerrarVent('PRNDIV" . $MisDatos->idxml . $MisDatos->xmlOrig . "');";

                    break;
                case 'refresh':
                    $reload[] = $MisDatos->refreshParentScript(null, false);
                break;
                case 'reload':
                    $reload[] = $MisDatos->refreshParentScript(null, true);
                break;
                default:
                    $parsedEvent = str_replace('_ORDER', $rowNumber - 1, $event);
                    $reload[] = $parsedEvent;
                    break;
            }
        }
       /*$debug = true;
        if($debug){
            var_dump($reload);
            die();
        }
        */
        return $reload;
    }
}

die();
?>