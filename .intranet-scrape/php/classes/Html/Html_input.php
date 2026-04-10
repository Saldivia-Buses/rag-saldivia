<?php
/**
 * Base input class
 * @author   Luis M. Melgratti
 * @category Histrix
 * Created on 19/01/2008
 *
 */

class Html_input {
    /*
    var $Parameters;
    var $Events;
    var $Style;
    var $value;
    var $dataType;
    var $tabIndex;
    var $label;
    var $lblid;
    var $id;
    var $nombre;
    var $OptionEvent;
    var $Eventos;
    var $tabindex;
    var $Formato;
    var $Formstyle;
*/
    function __construct() {

    }

    function addStyle($key, $value) {
        if ($value != '')
        $this->Style[$key]=$value;

        return $this; // make chainable
    }


    function addParameter($key, $value, $append=false) {
        if ($append){
            if (isset($this->Parameters[$key]))
                $this->Parameters[$key] .=$value;
            else $this->Parameters[$key]=$value;
        }
        else
            $this->Parameters[$key]=$value;

        return $this; // make chainable
    }

    /**
     * add Event to input field
     * @param mixed  $eventID event name
     * @param string  $value   javascript event 
     * @param boolean $append  append or replace events
     */
    function addEvent($eventID, $value, $append=false) {

        if (!is_array($eventID))
            $eventArray[]=$eventID;
        else 
            $eventArray = $eventID;

        foreach ($eventArray as $event) {
            if ($append == false ) 
                unset ($this->Eventos[$event]);

            if ($value != '')
                $this->Eventos[$event][$value] = $value;
                

            if ($append) 
                $this->Events[$event] .=$value;
            else
                $this->Events[$event] = $value;
        }
        return $this; // make chainable        
    }


    /**
     * remove events from input
     * @param  mixed $event events
     * @return none
     */
    public function removeEvent($events){
        if (!is_array($events))
            $eventArray[]=$events;
        else 
            $eventArray = $events;

        foreach ($eventArray as $value) {
            unset($this->Parameters[$value]);
            unset($this->Events[$value]);
            unset($this->Eventos[$value]);
        }
        
    }


    function setValue($value) {
        $this->value = $value;

        return $this; // make chainable
    }

    function getValue() {
        return $this->value;
    }

    function getStyleString() {
        $style ='';
        if (isset($this->Style) && is_array($this->Style))
            foreach ($this->Style as $claveArray => $valarray) {
                $style .= $claveArray.':'.$valarray.'; ';
            }
        return $style;
    }


    function getEventsString() {
        if (isset($this->Eventos))
        return $this->Array2String($this->Eventos);
    }

    function getParametersString() {
        return $this->Array2String($this->Parameters);
    }

    /**
    * Array2String converto an array to string parameters
    
    * @param array attribute-value array
    * @return string atributes
    */
    static function Array2String($arrayAtributos) {
        $atributos ='';
        if (is_array($arrayAtributos))
            foreach ($arrayAtributos as $claveArray => $valarray) {
                if (is_array($valarray)) {
                    $stratrib = '';
                    foreach ($valarray as $valarray2) {
                        $stratrib .= $valarray2.' ';
                    }
                    if (trim($stratrib) != '')
                        $atributos .= ' '.$claveArray.'="'.trim($stratrib).'" ';
                }
                else
                    if (trim($valarray) != '')
                        $atributos .= ' '.$claveArray.'="'.trim($valarray).'" ';
            }
        return $atributos;
    }


}


?>