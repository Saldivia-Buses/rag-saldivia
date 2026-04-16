<?php
include_once ("./autoload.php");

include ("../funciones/conexion.php");
include ('./sessionCheck.php');


$MisDatos  = new ContDatos("");
$instance = $_REQUEST["instance"];

$swap       = $_REQUEST['swap'];
$rowIndex  = $_REQUEST['_rowIndex'];

$NfilaOrig = $_REQUEST['Nro_Fila'];

if ($swap=='up' ) $NfilaDest = $NfilaOrig - 1;
else $NfilaDest = $NfilaOrig + 1;

$MisDatos = Histrix_XmlReader::unserializeContainer(null, $instance);
$xmldatos = $MisDatos->xml;

$Tablatemp = $MisDatos->TablaTemporal;


/** Drag And Drop method for reordering table */
if ($_REQUEST['dndSource'] != '') {

    $NfilaIni = min($_REQUEST['dndSource'], $_REQUEST['dndTarget']);
    $NfilaFin = max($_REQUEST['dndSource'], $_REQUEST['dndTarget']);

    $n = 0;
    for($i = $NfilaIni; $i <= $NfilaFin; $i++ ) {
        $slice[$n] = $Tablatemp->Tabla[$i];
        $n++;
    }

    if ($_REQUEST['dndSource'] > $_REQUEST['dndTarget'])
        $slice = array_reverse($slice);

    $cant = count($slice);
    for($i = 0; $i < $cant - 1; $i++ ) {
 
        $filaOrig = $slice[$i];
        $filaDest = $slice[$i + 1];

        foreach($filaOrig as $nom => $origValue) {
            $ObjCampo = $MisDatos->getCampo($nom);
            if (($ObjCampo->Parametro['esclave'])) {
                $keyField[$nom] = $nom;
            }
            else {
                $tmpVal = $filaOrig[$nom];
                $filaOrig[$nom] = $filaDest[$nom];
                $filaDest[$nom] = $tmpVal;
            }
        }
        $slice[$i]     = $filaOrig;
        $slice[$i + 1] = $filaDest;
    }


    _begin_transaction();
    for($i = 0; $i < $cant; $i++ ) {
        $MisDatos1 = new ContDatos("");
        $MisDatos1 = clone $MisDatos;
        $row = $slice[$i];
        foreach($row as $nom => $origValue) {
            
            // Erase previous Data
            $MisDatos1->setCampo($nom,null);
            $ObjCampo = $MisDatos1->getCampo($nom);

            if (($ObjCampo->Parametro['esclave'])) {
                $MisDatos1->setCampo($nom, $origValue);
                $MisDatos1->setNuevoValorCampo($nom, $origValue);
            }
            else {
                $MisDatos1->setNuevoValorCampo($nom, $origValue);
            }
        }
        $str1 = $MisDatos1->getUpdate(true, true);
        updateSQL($str1);
        unset($MisDatos1);
    }
    _end_transaction();
}
else {

    $filaOrig = $Tablatemp->Tabla[$NfilaOrig];
    $filaDest = $Tablatemp->Tabla[$NfilaDest];
    // swap records
    swap($MisDatos, $filaOrig, $filaDest);
}

$MisDatos->restaurarValores();
$UI = 'UI_'.str_replace('-', '', $MisDatos->tipo);
$datos = new $UI($MisDatos);

$datos->showTablaInt($opt, $xmldatos, $act);
//echo  $datos->showTablaInt($opt, '', $act, null, null, null, null, $MisDatos);

Histrix_XmlReader::serializeContainer($MisDatos);


function swap($MisDatos, $filaOrig, $filaDest) {
    $MisDatos1 = new ContDatos("");
    $MisDatos1 = clone $MisDatos;

    foreach($filaOrig as $nom => $origValue) {
        $destValue = $filaDest[$nom];

        // Erase previous Data
        $MisDatos1->setCampo($nom,null);

        $ObjCampo = $MisDatos1->getCampo($nom);

        if (($ObjCampo->Parametro['esclave'])) {
            $MisDatos1->setCampo($nom, $origValue);
            $MisDatos1->setNuevoValorCampo($nom, $origValue);
        }
        else {
            $MisDatos1->setNuevoValorCampo($nom, $destValue);
        }
    }
    
    $str1 = $MisDatos1->getUpdate(true, true);

    $MisDatos2 = new ContDatos("");
    $MisDatos2 = clone $MisDatos;

    foreach($filaOrig as $nom => $origValue) {
        $destValue = $filaDest[$nom];


        $ObjCampo = $MisDatos2->getCampo($nom);

        if (($ObjCampo->Parametro['esClave'])) {

            $MisDatos2->setCampo($nom, $destValue);
            $MisDatos2->setNuevoValorCampo($nom, $destValue);
        }
        else {
            // Erase previous Data
            $MisDatos2->setCampo($nom,null);
            $MisDatos2->setNuevoValorCampo($nom, $origValue);
        }
    }

    $str2 = $MisDatos2->getUpdate(true, true);
    
    if (trim($str1) != '' && trim($str2) != '' ) {
        updateSQL($str1);
        updateSQL($str2);
    }
}


?>