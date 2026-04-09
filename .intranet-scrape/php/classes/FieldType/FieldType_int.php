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
class FieldType_int extends FieldType_integer{

    const ALIGN = 'left'; // Default Alignment
    const DIR   = 'ltr';  // Text direction
    const INPUT = 'integer';

    public function __construct(&$field=null){
        $this->field = $field;
    }



}
?>
