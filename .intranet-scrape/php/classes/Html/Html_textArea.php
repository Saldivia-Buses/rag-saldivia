<?php
/**
 * Text Area
 * @author Luis M. Melgratti
 * Created on 19/01/2008
 *
 */
 class Html_textArea extends Html_input {
	var $cols;
 	function __construct(){
 		$this->cols = 47;
 	}

 	function show(){
 		$atributos 	= $this->getParametersString();
// 		$style		= $this->getStyleString();
		$salida = '<textarea '.$atributos.$this->tabindex.'   cols="'.$this->cols.'"  rows="'. ($this->size / 50).'"  '.$this->hide.'>';
	        $valor = $this->value;
	        if (!is_utf8($this->value))
        		$valor = utf8_encode($this->value);            
    		$salida .= $valor; 
		
		$salida .= '</textarea>';
        //        if (isset($this->Parameters['maxlength'])){
          //          $salida .= '<div class="fieldMessage msg" id="_FM_'.$this->Parameters['name'].'"></div>';
          //      }
 		return $salida;
 	}

 }

?>