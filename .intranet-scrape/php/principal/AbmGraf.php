<?php

/*
 * Created on 07/11/2005
 *
 * To change the template for this generated file go to
 * Window - Preferences - PHPeclipse - PHP - Code Templates
*/
include("./autoload.php");

$ArchivoConexion = "../funciones/conexion.php";

if (is_readable($ArchivoConexion))
    include_once($ArchivoConexion);
include ("./sessionCheck.php");
set_time_limit ( 200);


$Contenedor 	= new ContDatos("");
$xmldatos 	= $_GET["xmldatos"];
$accion 	= $_GET["accion"];
$idgraf 	= $_GET["idgraf"];
$instance     = $_GET["instance"];
$Contenedor     = Histrix_XmlReader::unserializeContainer(null, $instance);


if ($accion == 'update') {
    $hay = false;
    foreach($_POST as $npost => $post) {
        //echo $npost.' = '.$post.'<br>';
        $hay = true;
        if (is_array($post)) {
            foreach($post as $nvar => $var) {
                $npost2= str_replace('_xml', '.xml', $npost);
                $Contenedor->grafico[$npost2][$nvar]= $var;

                if (is_array($var)) {
                    unset($Contenedor->grafico[$npost2]['series'][0] );

                    $Contenedor->grafico[$npost2][$nvar][current($var)]= current($var);

                }
            }
        }

    }


    
    // create apropiate graficalInterface
    $UI = 'UI_'.str_replace('-', '', $Contenedor->tipo);
    $datos = new $UI($Contenedor);


    $datos->showDatos();
    //$datos->showTablaInt($opt, '', $act);


    $reloadgraf .='<script type="text/javascript">';
    /* foreach ($Contenedor->grafico as $id_grafico => $grafico){
		    $reloadgraf .="reloadImg('$idgraf'); ";
	} */
    $reloadgraf .="reloadImg('$idgraf'); ";
    $reloadgraf .='</script>';
    echo $reloadgraf;

    //print_r($Contenedor->grafico);
    //die();
}

if ($accion == 'add' || $Contenedor->grafico=='') {
    $uid = uniqid();
    $idGraf = $xmldatos.$uid;
    $id_grafico = $idGraf;
    $graf['ancho'] = 500;
    $graf['tipo'] = 'C';
    $graf['titulo'] = 'Titulo';
    $graf['subtitulo'] = 'Subtitulo';
    $graf['series'][] = '';
    $graf['datos'] = '';

    $Contenedor->grafico[$idGraf] = $graf;
    $UI = 'UI_'.str_replace('-', '', $Contenedor->tipo);
    $datos = new $UI($Contenedor);

    $datos->showDatos();
    //$datos->showTablaInt($opt, '', $act);

    Histrix_XmlReader::serializeContainer($Contenedor);

    $reloadgraf .='<script type="text/javascript">';
    foreach ($Contenedor->grafico as $id_grafico => $grafico) {
        $reloadgraf .="reloadImg('$id_grafico'); ";
    }
    $reloadgraf .='</script>';
    echo $reloadgraf;
    $idgraf = $idGraf;
}


echo '<form id="'.$xmldatos.'_graf" action="">';

if ($Contenedor->grafico)
/*
	 * foreach($Contenedor->grafico as $ngraf => $grafico){
	if ($idgraf != ''){
		if ($idgraf != $ngraf) continue;
	}
*/
    $ngraf = $idgraf;
$grafico = $Contenedor->grafico[$ngraf];
echo '<fieldset>';
echo '<table>';
echo '<tr>';
echo '<td>';

echo '<table>';
$id=$ngraf.'[titulo]';
echo '<tr>';
echo '<td>';
echo 'Titulo: </td><td>'	.input($id, $grafico['titulo']).'<br/>';
echo '</td>';

echo '</tr>';
$id=$ngraf.'[subtitulo]';
echo '<tr>';
echo '<td>';
echo 'Subtitulo: </td><td>'	.input($id, $grafico['subtitulo']).'<br/>';
echo '</td>';
echo '</tr>';
$id=$ngraf.'[ancho]';
echo '<tr>';
echo '<td>';
echo 'Ancho: </td><td>'	.input($id, $grafico['ancho']).'<br/>';
echo '</td>';
echo '</tr>';

$id=$ngraf.'[tipo]';
echo '<tr>';
echo '<td>';
echo 'Tipo: </td><td>'		.tipo($id, $grafico['tipo']).'<br/>';
echo '</td>';
echo '</tr>';
$id=$ngraf.'[etiquetas]';
echo '<tr>';
echo '<td>';
echo 'Etiquetas: </td><td>'	.selCampos($Contenedor, $grafico['etiquetas'], $id).'<br/>';
echo '</td>';
echo '</tr>';

if ($grafico['datos']) {
    $id=$ngraf.'[datos]';
    echo '<tr>';
    echo '<td>';
    echo 'Datos: </td><td>'		.$grafico['datos'].'<br/>';
    echo '</td>';
    echo '</tr>';

}
if ($grafico['series']) {
    foreach($grafico['series'] as $nserie => $serie) {
        echo '<tr>';
        echo '<td>';
        $id=$ngraf.'[series]['.$nserie.']';
        echo 'Serie: </td><td>'.selCampos($Contenedor, $nserie, $id).'<br/>';
        echo '</td>';
        echo '</tr>';

    }
}

echo '</table>';
echo '</td>';
echo '<td style="background-color:white;">';
echo '<img width="'.$grafico['ancho'].'" id="IMG'.$ngraf.'" src="grafico.php?grafico='.$ngraf.'&amp;uid='.uniqid().'" alt="Grafico no disponible" tittle="'.$grafico['titulo'].'">';
echo '</td>';
echo '</tr>';
echo '</table>';
echo '</fieldset>';


//echo '<button type="button" onclick="addGraf(null, \''.$xmldatos.'\', null );" >Agregar Gr&aacute;fico</button>';
echo '<div  class="filabotones" style="text-align:center;" >';
$btnOK = new Html_button('Aceptar', "../img/filesave.png" ,"Crear" );
$btnOK->addEvent('onclick', 'setGrafico(\''.$xmldatos.'_graf'.'\',\''.$xmldatos.'\', \''.$idgraf.'\', \''.$Contenedor->getInstace().'\' ); ');
echo $btnOK->show();
$btnCancel = new Html_button('Cerrar&nbsp;&nbsp;', "../img/cancel.png" ,"cerrar" );
$btnCancel->addEvent('onclick', 'cerrarVent(\'GRAF'.$xmldatos.'\')');
echo $btnCancel->show();
echo '</div>';
echo '</form>';



function input($id, $val) {
    $salida = '<input size="35" name="'.$id.'" value="'.$val.'" type="text">';
    return $salida;
}

function selCampos($Contenedor, $val, $id) {

    $listaCampos = $Contenedor->camposaMostrar();

    $salida .= '<select  id="'.$id.'" name="'.$id.'" >';
    foreach($listaCampos as $Nnombre => $nombrelista) {

        $ObjCampo = $Contenedor->getCampo($nombrelista);
        if ($ObjCampo->Parametro['noshow'] == 'true') continue ;
        $nombre =  htmlentities(ucfirst($ObjCampo->Etiqueta), ENT_QUOTES, 'UTF-8');
        $sel = '';
        if ($nombrelista == $val)
            $sel = 'selected="selected"';

        $salida .= '<option value="'.$nombrelista.'" '.$sel.'>'.$nombre.'</option>';
    }

    $salida .= '</select>';
    return $salida;
}
function tipo( $id, $val) {

    $salida .= '<select  id="'.$id.'" name="'.$id.'" >';
    $sel = '';
    if ($val == 'C') $sel = 'selected="selected"';
    $salida .= '<option value="C" '.$sel.'>Columnas</option>';
    $sel = '';
    if ($val == 'P') $sel = 'selected="selected"';
    $salida .= '<option value="P" '.$sel.'>Sectores</option>';
    $sel = '';
    if ($val == 'L') $sel = 'selected="selected"';
    $salida .= '<option value="L" '.$sel.'>Lineas</option>';

    $salida .= '</select>';
    return $salida;
}


?>