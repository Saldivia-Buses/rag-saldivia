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
class FieldType_varchar extends FieldType_char{

    const ALIGN   = 'left'; // Default Alignment
    const DIR     = 'ltr';  // Text direction
    const TYPE    = 'text';

    public function __construct(&$field=null){
        $this->field = $field;
    }




}
?>
