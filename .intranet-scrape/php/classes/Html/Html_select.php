<?php
/**
 * Select dropdown
 * @author Luis M. Melgratti
 * Created on 19/01/2008
 *
 */
class Html_select extends Html_input {
      /*
	var $opciones;
	var $xml;
	var $xmlOrig;
        */

    function __construct($opciones, $defaultValue='') {
        $this->opciones = $opciones;
        if ($defaultValue != '')
            $this->value = $defaultValue;

    }

    function addOptionEvent($eventID, $value, $append=false) {
        if ($eventID == false ) unset ($this->OptionEvent[$eventID]);
        if ($value != '')
            $this->OptionEvent[$eventID][$value] = $value;
    }

    function getOptionEventsString() {
        if (isset($this->OptionEvent) && $this->OptionEvent!= '')
            return $this->Array2String($this->OptionEvent);
    }

    function show() {
    //	$styleString= $this->getStyleString();
    //	$this->addParameter('style', $styleString);
        $atributos 	= $this->getParametersString();
        $optiontag = '';
        $this->match=false;
        $color='';
        $optionEvents = $this->getOptionEventsString();
        if (is_array($this->opciones)){
            foreach ($this->opciones as $miop => $key) {
                $sel = '';
                $tag_params= '';
                if (is_array($key)) {
                // recorro las opciones y paso los parametros al lleno cabecera
                    $keyclone = $key;
                    if (count($keyclone) > 1) {
                        $j=0;
                        foreach($keyclone as $nomOpcion => $valOpcion) {
                            if ($j >= 1) {
                            //creo los parametros para el tag option
                                $tag_params .= $nomOpcion.'="'.$valOpcion.'" ';
                                if ($nomOpcion =='Muestra') $color=$valOpcion;
                                $this->addEvent('onchange', 'setCampoSilent(this, \''.$nomOpcion.'\', \''.$this->xml.'\' , \''.$this->xmlOrig.'\'); ', true);
                            }
                            $j++;
                        }
                    }
                    $key = current($key);
                }

                if (isset($this->value) && $this->value == $miop && !$this->match) {
                    $sel = 'selected="selected"';
                    $this->match = true;
                }

                if (isset($this->Formato) && $this->Formato != '')
                    $key = sprintf($this->Formato, $key);
                $optiontag .= "\n\t";
                if ($color!='') $color='background-color:'.$color.';';

                //$optionEvents = $this->getOptionEventsString();
                $styleC = '';
                if($color != '') {
                    $styleC= 'style="'.$color.'"';
                }
                $optiontag .= '<option '.$optionEvents.' '.$styleC.' value="'.$miop.'"  '.$sel.' '.$tag_params.'>'.$key.'</option>';
            }
        }

        $eventos 	= $this->getEventsString();
        $tabindex = (isset($this->tabindex))?$this->tabindex:'';

        $salida = '<select  '.$atributos.' '.$eventos.' '.$tabindex.' >';

        if ($this->match != true) {

            if (is_array($this->opciones)) {
                $defaultValue= key($this->opciones);
                $valor = $defaultValue;

                $this->Campo_valor = $valor;
                $this->Campo_nuevovalor = $valor;
            }
        }

        $salida .= $optiontag;

        $salida .= '</select>';
        return $salida;
    }

}
?>
