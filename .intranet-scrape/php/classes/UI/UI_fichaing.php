<?php
/* 
 * 2009-09-09
 * help popup class 
 */

class UI_fichaing extends UI_ficha {

/**
 * User Interfase constructor
 *
 */
    public function __construct(&$Datacontainer) {
        parent::__construct($Datacontainer);
       
    }

    public function showTablaInt($opt = '', $idTabla = '', $segundaVez = '', $nocant='', $div=false, $form=null, $pdf=null, &$parentObject=null){
	return $this->show();
    }

    protected function addFormButtons(){
            $output ='<table width="100%"  border="0"  cellspacing="0" cellpadding="0" class="form" >';
            $output .= $this->showBtnIng();
            $output .='</table>';
            return $output;
    }

}

?>