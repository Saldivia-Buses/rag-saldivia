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
class FieldType_datetime extends FieldType_time{

    const ALIGN   = 'center'; // Default Alignment
    const DIR     = 'ltr';  // Text direction

    public function __construct(&$field=null){
        $this->field = $field;
    }



}
?>
