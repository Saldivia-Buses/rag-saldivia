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
class FieldType_decimal extends FieldType_numeric{

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

}
?>
