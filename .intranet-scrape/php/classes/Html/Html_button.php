<?php
 class Html_button extends Html_input{
	var $label;
	var $imgsrc;
	var $imgalt;
        var $tag;
 	function __construct($label=null, $imgsrc = null, $imgalt=null){
 		//$this->label  = htmlentities($label, ENT_QUOTES, 'UTF-8');
        $this->label  = $label;
 		$this->imgsrc = $imgsrc;
 		$this->imgalt = $imgalt;
 		$this->addParameter('type', 'button');
        $this->tag = 'button';
 	}

	function show(){
		$styleString= $this->getStyleString();
		if ($styleString != ''){
			$this->addParameter('style', $styleString, true);
		}
  		$atributos 	= $this->getParametersString();
  		$eventos 	= $this->getEventsString();
                $tabindex = (isset($this->tabindex))? $this->tabindex:'';
 		$salida = '<'.$this->tag.' '. $tabindex .$atributos.$eventos.'>';
 		if ($this->imgsrc != ''){
			$salida .= '<img src="'.$this->imgsrc.'" alt="'.$this->imgalt.'" /> ';
 		}
 		$salida .= $this->label;
 		$salida .= '</'.$this->tag.'>';
		return $salida;
	}
 }
?>
