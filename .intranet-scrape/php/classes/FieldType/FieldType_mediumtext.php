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
class FieldType_mediumtext extends FieldType_varchar{

    const ALIGN   = 'left'; // Default Alignment
    const DIR     = 'ltr';  // Text direction
	const INPUT   = 'text';

    public function __construct(&$field=null){
        $this->field = $field;
    }



}
?>
