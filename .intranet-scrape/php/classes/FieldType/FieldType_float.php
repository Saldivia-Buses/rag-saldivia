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
class FieldType_float extends FieldType_decimal{

    const ALIGN = 'right'; // Default Alignment
    const DIR   = 'ltr';  // Text direction
	const INPUT   = 'numeric';

    public function __construct(&$field=null){
        $this->field = $field;
    }

    public static function inputAttributes($value, $field){
        $attribute['type'] = self::INPUT;
        $attribute['step'] = 'any';
        return $attribute;
    }

    // for pdf printing
    public static function printValue($value){
        $value = (float) $value;
        $value = number_format($value, 4, '.', '');

	return $value;

    }


    public static function customValue($value, $field =null, $renderParameters=''){
        $value = (float) $value;
        $value = number_format($value, 4, '.', '');

	   return $value;
    }

}
?>
