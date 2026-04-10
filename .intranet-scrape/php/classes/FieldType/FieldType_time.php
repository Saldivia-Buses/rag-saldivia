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
class FieldType_time extends FieldType_varchar{

    const ALIGN   = 'center'; // Default Alignment
    const DIR     = 'ltr';  // Text direction
    const INPUT   = 'text';
    
    public function __construct(&$field=null){
        $this->field = $field;
    }



}
?>
