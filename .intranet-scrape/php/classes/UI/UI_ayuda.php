<?php
/* 
 * 2009-09-09
 * help popup class 
 */

class UI_ayuda extends UI_consulta {

/**
 * User Interfase constructor
 *
 */
    public function __construct($Datacontainer) {
        parent::__construct($Datacontainer);
        
        $this->disabledCheckDefault = true;
        $this->disabledCellId = true;
        $this->disableToolbar = true;
        $this->muestraCant = false;
        
        $this->defaultClass = 'consultahelp';

    }

    // render de complete XML
    public function show($idFormulario = '', $divcont='', $opt='') {

        $id = 'Show'.$this->Datos->idxml;
        
        // for unnamed help windows
	if ($this->Datos->idxml == '')
	    $this->Datos->idxml	= substr($divcont, 3);
        // id del contenedor (creo)
        $id2= str_replace('.', '_',($divcont != '')?$divcont:$id );

        $style = $this->Datos->style;

        // Columns
        $ancho = (isset($this->Datos->ancho))?$this->Datos->ancho : '';
        $width = (isset($this->Datos->width))?$this->Datos->width : $ancho;

        if ($width != '') {
            $this->Datos->col1=$width;
            $this->Datos->col2=100 - $width;
            $style.='width:'.$this->Datos->col1.'%;';
        }


        $clase  = $this->defaultClass;

        // Si se define explicitamente una clase en el xml
        if ($this->Datos->clase != '') {
            $clase = $this->Datos->clase;
        }

        // display Table
        $salidaDatos = $this->showTabla($opt);

        // create Utility dragBar
        if ($this->Datos->barraDrag != 'false') {

            $paramsDrag = $this->dragBarParameters();
            $paramsDrag['filter']='filter';
            $salidaDrag = $this->barraDrag2($id2,null, $paramsDrag ,true, null);
        }

        if ($this->Datos->campoRetorno != '') {
            $uidRetorno = $this->Datos->getCampo($this->Datos->campoRetorno)->uid;
            $retorno = ' origen="'.$uidRetorno.'" ';
        }

        $salida .=  '<div  class="'.$clase.'" id="'.$id.'" style="'.$style.'" '.$retorno.'>';
        $salida .= $salidaDrag;
        $salida .= '<div class="contewin" >';
        $salida .= $salidaDatos;
        $salida .= '</div>';
        $salida .= '</div>';

        // create Javascript functions
        $script[]= 'Histrix.registerTableEvents(\'TablaInterna'.$this->Datos->idxml.'\');';
        $script[]= "$('#$id2').draggable({handle:'#dragbar$id2'});";
        $script[]= "Histrix.registroEventos('".$this->Datos->idxml."')";

        $salida .= Html::scriptTag($script);

        return $salida;

    }


    public function showTabla($opt = '') {

        $idTabla = $this->Datos->idxml;
        $estiloPriv= 'position:absolute;top:0px;bottom:0px; left:0px;right:0px; ';

        $salidaTabla = $this->showTablaInt($opt, $idTabla);
        if ($this->Datos->__inline=='true') return $salidaTabla;

        $tablaInt = Html::tag('div', $salidaTabla, array('id' => $idTabla, 'class' => 'contTablaInt'));
        $propDiv  = array('id'=>'IMP'.$idTabla, 'class'=>'TablaDatos', 'cellpadding'=>0, 'cellspacing'=>0,'style'=>$estiloPriv );
        $salida   = Html::tag('div', $filtros.$tablaInt, $propDiv );

        return $salida;
    }


}

?>