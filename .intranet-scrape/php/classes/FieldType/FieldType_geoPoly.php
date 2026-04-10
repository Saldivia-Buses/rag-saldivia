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
class FieldType_geoPoly extends FieldType_geoPoint{

    const ALIGN   = 'left'; // Default Alignment
    const DIR     = 'ltr';  // Text direction
    const INPUT   = 'geoPoly';  // input type
    const HIDDEN  = true;  // input type
    const CUSTOM  = true;     // use custom Construct Method


    public function __construct(&$field=null){
        $this->field = $field;
    }


}
?>
