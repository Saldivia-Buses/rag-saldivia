<?php
/**
 * input Text Box
 * @author Luis M. Melgratti
 * Created on 19/01/2008
 *
 */
class Html_textBox extends Html_input 
{
        /*
	var $Campo_valor;
	var $Formato;
        var $prefijo;
*/

    function __construct($valor, $TipoDato) {

        $this->value 	= $valor;
        $this->TipoDato = $TipoDato;
        $this->sufijo    = '';
    }

    function show() {

    //	$style		= $this->getStyleString();
        $style = '';
        $maxsizeStr = '';
        $size		= $this->size;

        $div = '';
        $maxsize = '';
        $valor = $this->value;

        /// FORMATEO EXPLICITO 
        if (isset($this->Formato)) {
            // las fechas se formatean diferente
            if ($this->TipoDato =='date') {
                $valor = date($this->Formato, strtotime($this->Campo_valor));
                $size = strlen($this->value);
            } else {
                $valor = sprintf($this->Formato, $this->value);
                $size  = strlen(sprintf($this->Formato, $valor));
            }
        }

        $this->addParameter('value', $valor);

        if ($this->valorCampo != '')
            $this->addParameter('value', $this->valorCampo);


        // limito el tama�o por que sino no entran
        if ($size > 77 && $this->tipoAbm != 'ficha')
            $size = 77;

        if ($this->TipoDato== 'numeric') {
            $maxsize = $size * 2;
            $size = $size * 2 / 3 ;

        }

        if(isset( $this->maxsize) ) $maxsize = $this->maxsize;



        if ($size > 40 && $this->tipoAbm == 'ing' && $this->forceSize != 'true')
            $size = 40;

        $eventos 	= $this->getEventsString();
        $atributos 	= $this->getParametersString();
        $strmaxsize = ($maxsize != '')?  ' maxlength="'.$maxsize.'" ' :'';
    
        $input = '<input '.$atributos.' '.$eventos .' '.$this->tabindex.' size="'.$size.'"  '.$strmaxsize.'/>';

        if ( ( isset($atributos['disabled']) && $atributos['disabled'] != '' ) || ( isset($atributos['readonly']) && $attributos['readonly'] != '')) {
            if ($this->deshabilitado == 'true') {
                $maxsize = $size;
                //$size = $this->tammax;
                //
                //$tabindex= '';
                if($size=='') $size=$maxsize;

            }
            if (isset($this->Formstyle))
                $style.=$this->Formstyle;

            if (isset($this->maxsize)) {
                $maxsizeStr = ' maxsize="'.$this->maxsize.'" ';
            }

            $input = '<input '.$atributos.'style="'.$style.'" '.$eventos.' '.$this->tabindex.' size="'.$size.'" '.$maxsizeStr.' />';
        }

        $prefix = (isset($this->prefijo))?$this->prefijo:'';

        $input = $prefix.$input.$this->sufijo;
        return $input;
    }

}
?>
