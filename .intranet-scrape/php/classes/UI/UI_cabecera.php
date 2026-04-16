<?php
/* 
 * 2009-09-09
 * help popup class 
 */

class UI_cabecera extends UI_ficha {

/**
 * User Interfase constructor
 *
 */
    public function __construct(&$Datacontainer) {

        parent::__construct($Datacontainer);

        $this->hasFieldset = 'false';

        $this->defaultClass = 'cabecera';
    }

    // render de complete XML
    public function show($idFormulario = '', $divcont='', $opt='') {

        $clase = ($this->Datos->clase != '')? $this->Datos->clase:'';

        $this->Datos->CargoTablaTemporalDesdeCampos();
        $this->Datos->calculointerno();
        $this->Datos->CargoCamposDesdeTablaTemporal();

        $salidaAbm = $this->showAbm( null, $clase);

        $customjs    = 'Histrix.registerTableEvents(\'TablaInterna'.$this->Datos->idxml.'\');';

        // El Abm
        $salida = $salidaAbm;

        // create Javascript functions
        $script[]= $customjs;
        //$script[]= "$('#$id2').draggable({handle:'#dragbar$id2'});";
        $script[]= "Histrix.registroEventos('".$this->Datos->idxml."')";

        $salida .= Html::scriptTag($script);
        return $salida;

    }
    

    public function showAbm($modoAbm = '', $clase = '') {
        if ($clase != '')
            $class = $clase;
        else
             $class = 'class="cabecera"';

        $idContenedor = $this->Datos->idxml;
        $intClass= 'class="contewin"';

        if ($this->Datos->__inline == true) {
            $intClass= '';
        }

        $style = '';
        if ($this->Datos->col2 != '') $style .='width:'.($this->Datos->col2 - 0.5).'%;';
        $salida = '<div '.$class.' id="DIVFORM'.$idContenedor.'" style="'.$style.'">'.
                   '<div '.$intClass.' id="INT'.$idContenedor.'">';
        $salida .= $this->showAbmInt($modoAbm, 'INT'.$idContenedor);
        $salida .= '</div></div>';

        return $salida;
    }

}

?>