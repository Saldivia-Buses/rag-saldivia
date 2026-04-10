<?php
/* 
 * 2009-09-09
 * help popup class 
 */

class UI_ficha extends UI_abm {

/**
 * User Interfase constructor
 *
 */
    public function __construct(&$Datacontainer) {
        parent::__construct($Datacontainer);

        $this->showFormOptions = true;
        $this->defaultClass = 'Fichagrande';
       
    }

               
    // pdf printing of data 
    public function pdf($pdf , $fontsize = '', $opImpresion = '', $anchoTabla = '', $posx = ''){
        $pdf->SetY(22);
        $tabla = $pdf->impAbm($this->Datos);
        $wcab = $pdf->setAnchoCol($tabla) ;
        if ($pdf->GetY() == '')
            $pdf->SetY(25);
        $pdf->SetY($pdf->GetY() + 4);
        $pdf->WriteTable($tabla, $wcab);
    }
    
    
    public function initialJavascript(){
        if ($this->Datos->sololectura != 'true' && $this->Datos->conBusqueda == 'true')
             $js = "Histrix.clearForm('".$this->Datos->xml."', true);";
        else $js = '';
        return $js;
    }

    // render de complete XMLf
    public function show($idFormulario = '', $divcont='', $opt='') {

//	    if ($this->Datos->importTempTable == ''){
	       $this->Datos->CargoTablaTemporalDesdeCampos();
            $this->Datos->calculointerno();
            $this->Datos->CargoCamposDesdeTablaTemporal();
//        }
         

        if (isset($this->Datos->preFetch) && $this->Datos->preFetch=='true'){
           $this->Datos->Select();
           
        }

        // El Abm
        $salida = $this->showAbm();

        // create Javascript functions
       
        $script[]= "Histrix.registroEventos('".$this->Datos->idxml."')";
		$script[]= $this->Datos->getCustomScript();
        $salida .= Html::scriptTag($script);
        return $salida;

    }

    public function showAbm($modoAbm = '', $clase = '') {
        if ($clase != '')
            $class = $clase;
        else
             $class = 'class="'.$this->defaultClass.'"';

        $idContenedor = $this->Datos->idxml;
        $intClass= 'class="contewin"';

        if (isset($this->Datos->__inline) && $this->Datos->__inline == true) {
            $intClass= '';
            $class ='class="Fichagrande2"';
        }

        $style= '';
        if (isset($this->Datos->col2)) $style ='width:'.($this->Datos->col2 - 0.5).'%;';

        $salida  = '<div '.$class.' id="DIVFORM'.$idContenedor.'" style="'.$style.'">'.
                   '<div '.$intClass.' id="INT'.$idContenedor.'" instance="'.$this->Datos->getInstance().'">';
        $salida .= $this->showAbmInt($modoAbm, 'INT'.$idContenedor);
        $salida .= '</div></div>';

        return $salida;
    }



}

?>