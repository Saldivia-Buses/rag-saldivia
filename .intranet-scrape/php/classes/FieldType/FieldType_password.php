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
class FieldType_password extends FieldType_varchar{

    const ALIGN = 'left'; // Default Alignment
    const DIR   = 'ltr';  // Text direction
    const TYPE  = 'text';  // input type

    public function __construct(&$field=null){
        $this->field = $field;
    }


    public static function customValue($value, $field, $parameters='')
    {
                if ($value != '' ){

                    return $value;
                }
    }

    public static function inputAttributes($valor, $field){
        $attribute['type'] = self::TYPE;
        return $attribute;
    }
}
?>
