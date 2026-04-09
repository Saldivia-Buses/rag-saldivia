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
class FieldType_custom_numeric extends FieldType_numeric{

    const ALIGN = 'right'; // Default Alignment
    const DIR   = 'rtl';  // Text direction


    public function __construct(&$field=null){
        $this->field = $field;
    }


    public static function customValue($value, $field='', $renderParameters=''){
        if (is_numeric($value))
                $value = number_format($value, 2, ',', '.');

	   return $value;
    }


}
?>
