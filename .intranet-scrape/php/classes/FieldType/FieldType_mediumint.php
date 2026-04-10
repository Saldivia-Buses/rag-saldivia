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
class FieldType_mediumint extends FieldType_int{

    const ALIGN = 'left'; // Default Alignment
    const DIR   = 'ltr';  // Text direction


    public function __construct(&$field=null){
        $this->field = $field;
    }



}
?>
