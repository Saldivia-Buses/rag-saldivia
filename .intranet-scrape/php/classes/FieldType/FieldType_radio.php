<?php
/* 
 * DtataType Class
 * 
 */

/**
 * Define FieldType representation
 *
 * @author luis
 */
class FieldType_radio extends FieldType{

    const ALIGN   = 'left'; // Default Alignment
    const DIR     = 'ltr';  // Text direction
    const INPUT   = 'radio';  // input type


    public function __construct(&$field=null){
        $this->field = $field;
    }


    public static function renderInput( $valor, $field, $arrayAtributos, $uiClass, $opciones){
    
        $inputBox = new Html_radio($opciones, $valor);
        $inputBox->Parameters=$arrayAtributos;
        $inputBox->addParameter('type', self::INPUT);

        //$inputBox->addEvent('onchange', $actualizarSelect2, true);

        $salida = $inputBox->show();
                
 
        return $salida;
    }

}
?>
