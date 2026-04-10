<?php


/*
 * 2009-09-09
 * help popup class
 */

class UI_grid extends UI_ing {

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
        $this->procesa =false;        
        $this->hasForm = false;
        $this->dlbClickEditRow = false;
    }

    protected function addRowCell() {

    }

    protected function customTotalJavascript($ObjCampo, $cellValue='') {
        $script[]= $this->Datos->getCustomScript();

        if ($ObjCampo->uid2 != ''){

            $script[] = 'Histrix.calculoTotal($(\'#' . $ObjCampo->uid2 . '\')[0]);';
        }

        $salida = Html::scriptTag($script);

        return $salida;

    }

    public function updateTotals() {
        $campos = $this->Datos->camposaMostrar();
        foreach ($campos as $nom => $valor) {

            if ($this->Datos->seSuma($valor)) {
                $ObjCampo = $this->Datos->getCampo($valor);                
                $salida .= $this->customTotalJavascript($ObjCampo);
            }
        }
        return $salida;                
    }

    protected function inlineCrud($idTableForm, $form , $opt, $formini = '', $formfin = '' , $segundaVez = ''){
	$procesa = (isset($this->Datos->procesa))?$this->Datos->procesa:$this->procesa;

    $procesa = (isset($this->Datos->process))?$this->Datos->process:$procesa;

	if ($procesa == 'true'){
	   $this->Datos->grillasContenidas[$this->Datos->xml] = $this->Datos->getInstance();
	   $this->Datos->modificaABM = 'false';
           $output ='<table '.$idTableFrom.' width="100%"  class="form">';

           $output .= $this->showBtnIng($opt);
           $output .='</table>';
           return $output;
        }
           
    }

    public function showTabla($opt = '') {

        $idTabla = $this->Datos->idxml;

        if ($this->Datos->imprime != 'false')
            $bottom = 31;
        else
            $bottom = 0;

        //  $estiloPriv= 'position:absolute;top:0px;bottom:'.$bottom.'px; left:0px;right:0px; overflow:auto;';

        if ($this->contFiltro || $this->Datos->filtros || $this->Datos->autofiltro != 'false') {
            if ($this->Datos->autofiltro != 'false')
                $filtros = $this->autoFiltros();
            $filtros .= $this->showFiltrosXML();
        }
        
        $salidaTabla = $this->showTablaInt($opt, $idTabla);

        if ($this->Datos->__inline == 'true')
            return $salidaTabla;
  
        $tablaInt = Html::tag('div', $salidaTabla,
                        array('id' => $idTabla, 'class' => 'contTablaInt', 'instance' => $this->Datos->getInstance()));
        $propDiv = array('id' => 'IMP' . $idTabla, 'class' => 'TablaDatos',
            'cellpadding' => 0, 'cellspacing' => 0, 'style' => $estiloPriv);
        $salida = Html::tag('div', $filtros . $tablaInt, $propDiv);
//        $salida .= $this->importDataButton();
        $salida .= $this->botonera();

        return $salida;
    }

}

?>