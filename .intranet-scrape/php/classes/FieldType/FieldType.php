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
class FieldType  {

    const ALIGN = 'left'; // Default Alignment
    const DIR   = 'ltr';  // Text direction

    public function __construct(&$field){
        $this->field = $field;
    }


    public static function extraData(){
        return '';
    }

    public static function customCellParameters(){
        return '';
    }

    public static function inputAttributes($value, $field){
        $attribute['type'] = 'text';
        return $attribute;
    }

}
?>
