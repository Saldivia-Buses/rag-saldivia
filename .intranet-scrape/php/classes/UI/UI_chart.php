<?php
/* 
 * 2009-09-09
 * help popup class 
 */

class UI_chart extends UI_consulta {

/**
 * User Interfase constructor
 *
 */
    public function __construct($Datacontainer) {
       parent::__construct($Datacontainer);
       

    }

    public function show($idFormulario = '', $divcont='', $opt='') {

        $id = 'Show'.$this->Datos->idxml;

        // id del contenedor (creo)
        $id2=($divcont != '')?$divcont:$id;

        $id2  = str_replace('.', '_', $id2);

        $style = $this->Datos->style;


        $clase		 = 'consultaing2';
        $BarraDrag	 = false;

        $extra_data['top'][] = $this->setObjectives();

        $salida .= $this->showFiltrosXML(false,true, $extra_data);

        $salida .= '<div id="'.$this->Datos->idxml.'" ></div>';

        $noshow = $this->showTabla();


        // Graficos
        if ($this->Datos->grafico != '') {
            $salida .= $this->showGraficos();
        }


        // add events
        $salida .= $this->eventScripts();
        return $salida;

    }

    function setObjectives(){
        $btn = new Html_button($this->i18n['objectives'], "../img/add.png" ,$this->i18n['objectives'] );
        $btn->addStyle('float','right');
        
        foreach ($this->Datos->grafico as $key => $value) {
            
            $option_name = $value['option_name'];
        }
        

        $target = 'histrixLoader.php?xml=htxoption_crud.xml&dir=histrix/prefs&option_name='.$option_name;   
        $parent = 'DIV'.$this->Datos->xml;   
        $opcetiq = '\'\'';          
        $btn->addEvent('onclick', 'Histrix.loadInnerXML(\'htxoption_crud.xml\', \''.$target.'\', '.$opcetiq.',\''.$this->i18n['objectives'].'\', \''.$parent.'\',\''.$uid.'\' )');
        $salida .= $btn->show();

        return $salida;
    }

}

?>