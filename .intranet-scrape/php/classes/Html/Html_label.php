<?php
 class Html_label extends Html_input{
	var $value;
	var $for;

 	function __construct($value, $for=null){
                $this->value  = $value;
 		$this->for    = $for;
 	}

	function show(){
		$styleString= $this->getStyleString();
		if ($styleString != ''){
			$this->addParameter('style', $styleString, true);
		}
  		$attributes = $this->getParametersString();
  		$events 	= $this->getEventsString();

        $output = '<label   '.$attributes.' '.$events.'>'.htmlentities(ucfirst($this->value), ENT_QUOTES, 'UTF-8').'</label>';
		return $output;
	}
 }
?>
