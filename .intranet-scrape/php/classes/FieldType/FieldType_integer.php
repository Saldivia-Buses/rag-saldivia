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
class FieldType_integer extends FieldType{

    const ALIGN = 'left'; // Default Alignment
    const DIR   = 'ltr';  // Text direction
	const INPUT   = 'number';

    public function __construct(&$field=null){
        $this->field = $field;
    }

    public static function inputAttributes($value, $field){
        $attribute['type'] = self::INPUT;
        return $attribute;
    }


}
?>
