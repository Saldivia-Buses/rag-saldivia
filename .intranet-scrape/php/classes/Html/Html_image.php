<?php
 class Html_image extends Html_input{
	var $title;
	var $imgsrc;
	var $imgalt;
 	function __construct($imgsrc, $title=null, $imgalt=null){
 		//$this->title  = htmlentities($title, ENT_QUOTES, 'UTF-8');
                $this->title  = $title;
 		$this->imgsrc = $imgsrc;
 		$this->imgalt = $imgalt;
 	}

	function show(){
		$styleString= $this->getStyleString();
		if ($styleString != ''){
			$this->addParameter('style', $styleString, true);
		}
  		$atributos 	= $this->getParametersString();
  		$eventos 	= $this->getEventsString();

                $salida = '<img title="'.$this->title.'" src="'.$this->imgsrc.'" alt="'.$this->imgsrc.'" '.$atributos.' '.$eventos.' /> ';
		return $salida;
	}
 }
?>
