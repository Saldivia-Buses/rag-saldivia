<?php
/* 
 * DtataType Class
 * 
 */

/**
 * Define FieldType representation
 *
 * @author luis
 */
class FieldType_button extends FieldType{

    const ALIGN   = 'left'; // Default Alignment
    const DIR     = 'ltr';  // Text direction
    const INPUT   = 'button';  // input type


    public function __construct(&$field=null){
        $this->field = $field;
    }



}
?>
