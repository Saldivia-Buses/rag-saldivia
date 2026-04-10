<?php
/* 
 * FieldType Class
 * 
 */


/**
 * Define FieldType representation
 *
 * @author luis
 */
class FieldType_geoPoint extends FieldType{

    const ALIGN   = 'left'; // Default Alignment
    const DIR     = 'ltr';  // Text direction
    const INPUT   = 'geoPoint';  // input type
    const HIDDEN  = true;  // input type
    const CUSTOM  = true;     // use custom Construct Method


    public function __construct(&$field=null){
        $this->field = $field;
    }

    public function init(){
        // get mapkey
        $key = $self::getMapKey();

        // add javascript library
        $self::addJavascript_v3($key);
        //

    }

    private function getMapKey(){
        /* CES-MAP */
        $key = 'ABQIAAAAGVe7lkI-epJEXCKWsKR6AhTx9wtw6p3SeGVhkyYVMmElqSVG7xT33Jh6bqccdYly9uL-gFVp28Jdsg';
        return $key;
    }

    private function addJavascript_v3(){
        $js = '<script type="text/javascript" src="http://maps.google.com/maps/api/js?sensor=false"></script>';
        echo $js;
    }


    private static function addJavascript($mapkey){
        $js = '<script src="http://maps.google.com/maps?file=api&v=2&key='.$mapkey.'" type="text/javascript"></script>';
    //    echo $js;

        $javascripts[] = 'geometrycontrols.js';
        $javascripts[] = 'markercontrol.js';
        $javascripts[] = 'polygoncontrol.js';
        $javascripts[] = 'polylinecontrol.js';
        $jspath = '../javascript/';

        foreach($javascripts as $n => $js) {
            if(is_file($jspath.$js)) {
                $js_size = filesize($jspath.$js);
                echo '<script type="text/javascript" src="'.$jspath.$js.'?'.$js_size.'"></script>';
             }
             else {
                echo 'Error Loading Javascript: '.$js.'<br>';
             }
        }
    }



    public static function renderInput( $valor, $field, $arrayAtributos, $uiClass, $opciones=''){

            $inputBox = new Html_textBox($valor, $field->TipoDato);
            $inputBox->Parameters=$arrayAtributos;
            $inputBox->addParameter('ctype', 'geo');
            $inputBox->addParameter('type' , 'hidden');

            $geoMap = new geoMap($valor, self::INPUT);
            $geoMap->Parameters = $arrayAtributos;

            $geoMap->size   = $field->size;

            $salida = $inputBox->show();
            $salida .= $geoMap->show();

            return $salida;
        }

/*
    public function display(){
        $output  = '<div class="geoPoly">';
        $output .= 'Polygon';
        $output .= '</div>';
        return $output;
    }
*/
}
?>
