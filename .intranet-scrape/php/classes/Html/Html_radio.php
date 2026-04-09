<?php
 /**
 * Radio button options
 * @author Luis M. Melgratti
 * Created on 19/01/2008
 *
 */

class Html_radio extends Html_input {
    /*
    var $opciones;
    var $xml;
    var $xmlOrig;
    */

    function __construct($opciones, $value= '', $extra='') {
        $this->opciones = $opciones;
        $this->value 	= $value;
        $this->extra 	= $extra;
    }

    function show() {
        $styleString= $this->getStyleString();
        $this->addParameter('style', $styleString);
        $this->addParameter('valauto', 'true');


        $this->match=false;
        $color='';
        $cont= 0;
        $eventos 	= $this->getEventsString();
        $salida = '';
        $optiontag = '';
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

            if ($this->value == $miop) {
                $sel = 'checked="checked"';
                $this->match = true;
            }

            if (isset($this->Formato) && $this->Formato != '')
                $key = sprintf($this->Formato, $key);

            $optiontag .= "\n\t";
            if ($color!='') $color='background-color:'.$color.';';

            $this->addParameter('id', $this->Parameters['name'].'_'.$cont);
            $atributos 	= $this->getParametersString();
            $tabindex = (isset($this->tabindex))? $this->tabindex:'';
            
            $optiontag .= '<input '.$atributos.' '.$sel.' value="'.$miop.'"  '.$eventos.$tabindex.' /><label for="'.$this->Parameters['id'].'">';
            if ($this->extra != '') {
                $optiontag .= $this->extra[$miop];
            }
            else $optiontag .= htmlentities($key,null, 'UTF-8');

            $optiontag .='</label>';
            $cont++;
        }


        if ($this->match != true) {
            $defaultValue= key($this->opciones);
            $valor = $defaultValue;

            $this->Campo_valor = $valor;
            $this->Campo_nuevovalor = $valor;
        }

        $salida .= $optiontag;

        return $salida;
    }

}
?>
