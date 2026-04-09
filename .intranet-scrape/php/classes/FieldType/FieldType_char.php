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
class FieldType_char extends FieldType{

    const ALIGN   = 'left'; // Default Alignment
    const DIR     = 'ltr';  // Text direction

    public function __construct(&$field=null){
        $this->field = $field;
    }

    public static function extraData(){
        return 'varchar';
    }


}
?>
