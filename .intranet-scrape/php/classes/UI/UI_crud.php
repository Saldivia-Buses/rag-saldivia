<?php

/*
 * 2011-06-21
 * crud popup class
 */

class UI_crud extends UI_abm {

    /**
     * User Interfase constructor
     *
     */
    public function __construct(&$Datacontainer) {
        parent::__construct($Datacontainer);
        $this->hasForm = true;
        $this->formClass = 'singleForm';
        $this->muestraCant = true;

//        $this->updateButton = ($Datacontainer->modifica == 'false')?false:true;

        $this->showFormOptions = true;
        $this->defaultClass = 'consultaing2';
        $this->slider = false;
    }

    function showSlider($id, $retrac='') {
          if ($this->slider){
             return parent::showSlider($id, $retrac);
          }
         
    }

    public function showAbm($modoAbm = '', $clase = '') {
        if ($clase != '')
            $class = $clase;
        else {

            if (isset($this->defaultClass)) {
                $class = 'class="' . $this->defaultClass . '"';
            }
        }
        $idContenedor = $this->Datos->idxml;

        $style = '';
        $strStyle = '';
        if ($this->Datos->col2 != '')
            $style .='width:' . ($this->Datos->col2 - 0.5) . '%;';

        if ($style != '')
            $strStyle = ' style="' . $style . '" ';

        $salida = '<div style="display:none;" ' . $class . ' id="DIVFORM' . $idContenedor . '" ' . $strStyle . ' >' .
                '<div class="contewin" id="INT' . $idContenedor . '">';
        $salida .= $this->showAbmInt($modoAbm, 'INT' . $idContenedor);
        $salida .= '</div></div>';

        return $salida;
    }

    public function addButton($label=false, $image = "../img/add22.png"){
        if ($this->Datos->inserta != 'false' && $this->Datos->sololectura != 'true') {

            if ($label == false)
                $label = $this->i18n['new'];

            $btnImprimir = new Html_button($label, $image);
            $btnImprimir->addEvent('onclick', '$(\'#DIVFORM' . $this->Datos->idxml . '\').slideDown();');
            $btnImprimir->tabindex = $this->tabindex();


            $salida = $btnImprimir->show();
        }
        return $salida;
    }
    public function botonera($buttons='') {

        $addButton = $this->addButton();

        $salida = parent::botonera($addButton);
        return $salida;
    }

    protected function topButtons() {
        $btnCancel =  new Html_button('', "../img/remove.png", $this->i18n['cancel']);
        $btnCancel->addParameter('name', $this->i18n['cancel']);
        $btnCancel->addEvent('onclick', 'Histrix.clearForm(\'' . $this->Datos->xml . '\', true)');
        $btnCancel->addStyle('border-radius', '20px');
        $btnCancel->addStyle('background-color', 'transparent');
        $btnCancel->addStyle('width', '26px');
                $btnCancel->addStyle('height', 'auto');
       // $btnCancel->addStyle('background-image', "url('../principal/gradient.php?o=v&s=30&c1=ffffff&c2=ff0000')");
         $btnCancel->addStyle('background-image', "none");
        $htmlCancel .= $btnCancel->show();

      //  $htmlCancel = '<img src="../img/remover.png" onclick="'.'Histrix.clearForm(\'' . $this->Datos->xml . '\', true)'.'" />';
        $output = Html::tag('div',  $htmlCancel, array('class' => 'barraBusqueda'));
        return $output;
    }

}

?>