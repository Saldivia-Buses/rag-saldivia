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
class FieldType_tinyint extends FieldType_int{

    const ALIGN = 'left'; // Default Alignment
    const DIR   = 'ltr';  // Text direction


    public function __construct(&$field=null){
        $this->field = $field;
    }



}
?>
