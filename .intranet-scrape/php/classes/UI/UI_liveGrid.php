<?php
/* 
 * 2009-09-09
 * help popup class 
 */

class UI_liveGrid extends UI_grid {

/**
 * User Interfase constructor
 *
 */
    public function __construct($Datacontainer) {
        parent::__construct($Datacontainer);
        $this->rowDeleteButton = true;


        $this->disabledCheckDefault = false;
        $this->resizeTable = "false";
        $this->hasFieldNameReference = true;
        $this->enableCheckToggle = true;
        
    }

    
    protected  function addRowCell(){
       if ($this->Datos->inserta != 'false'){
         $campos = $this->cantCampos() - 1;
         $output = '</tr></tr><td  colspan="'.$campos.'" _class="sintotal" ><button style="width:auto;" class="addRow" title="'.$this->i18n['new'].'" ><img src="../img/list-add.png"/></button></td></tr>';
       }
       return $output;
    }

    public function showTabla($opt = '') {

        $idTabla = $this->Datos->idxml;

        if ($this->Datos->imprime != 'false') $bottom = 31;
        else $bottom = 0;

      //  $estiloPriv= 'position:absolute;top:0px;bottom:'.$bottom.'px; left:0px;right:0px; overflow:auto;';

        if ($this->contFiltro || $this->Datos->filtros || $this->Datos->autofiltro != 'false') {
            if ($this->Datos->autofiltro !='false')
                $filtros = $this->autoFiltros();
            $filtros .= $this->showFiltrosXML();
        }

        $salidaTabla = $this->showTablaInt($opt, $idTabla);
        if ($this->Datos->__inline=='true') return $salidaTabla;

        $tablaInt = Html::tag('div', $salidaTabla,
            array('id' => $idTabla, 'class' => 'contTablaInt'));
        $propDiv  = array('id'=>'IMP'.$idTabla, 'class'=>'TablaDatos',
            'cellpadding'=>0, 'cellspacing'=>0,'style'=>$estiloPriv );
        $salida = Html::tag('div', $filtros.$tablaInt, $propDiv );
        $salida .= $this->importDataButton();
        $salida .= $this->botonera();

        return $salida;
    }

    protected  function rowButtons($orden) {
        if ($this->Datos->deleteRow != 'false') {
            $TableData = '<td class="delrow" campo="Nro_Fila" valor="'.$orden.'"  ><div class="deleteImageButton" /></td>';
        }
        else {
         //   $TableData = '<td></td>';
        }
        return $TableData;
    }


}

?>