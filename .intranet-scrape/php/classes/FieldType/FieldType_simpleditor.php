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
class FieldType_simpleditor extends FieldType_editor{

    const ALIGN   = 'left'; // Default Alignment
    const DIR     = 'rtl';  // Text direction
    const INPUT   = 'simpleditor';  // input type


    public function __construct(&$field=null){
        $this->field = $field;
    }

    public static function customCellParameters(){
        return 'editor="true"';
    }


}
?>
