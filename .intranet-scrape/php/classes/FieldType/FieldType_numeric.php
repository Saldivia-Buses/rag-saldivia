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
class FieldType_numeric extends FieldType{

    const ALIGN = 'right'; // Default Alignment
    const DIR   = 'rtl';  // Text direction
    const INPUT = 'text';

    public function __construct(&$field=null){
        $this->field = $field;
    }

    // for pdf printing
    public static function printValue($value){
        $value = (float) $value;
        $value = number_format($value, 2, '.', '');

	return $value;

    }


    public static function customValue($value, $field =null, $renderParameters=''){
        $value = (float) $value;
        $value = number_format($value, 2, '.', '');

	   return $value;
    }


}
?>
