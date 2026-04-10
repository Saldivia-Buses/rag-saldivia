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
class FieldType_month extends FieldType_varchar{

    const ALIGN   = 'left'; // Default Alignment
    const DIR     = 'ltr';  // Text direction
    const INPUT   = 'text';

    public function __construct(&$field=null){
        $this->field = $field;
    }

    public static function extraData(){
        return 'varchar';
    }

    public static function inputAttributes($value, $field){
        $attribute['class'] = 'month';
        return $attribute;
    }
}
?>
