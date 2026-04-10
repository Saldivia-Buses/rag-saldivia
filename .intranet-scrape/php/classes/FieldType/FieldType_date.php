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
class FieldType_date extends FieldType_time{

    const ALIGN = 'center'; // Default Alignment
    const DIR   = 'ltr';  // Text direction


    public function __construct(&$field=null){
        $this->field = $field;
    }

/*
    public static function extraField(){
        if ($hide == '' || $deshabilitado != 'true') {

            if (isset($objCampo->dateselector) && $objCampo->dateselector =='false') {
                 $dat = '';
            }
            else     $dat = '<a href="#"><img inputid="'.$uniqid2.'" class="selFecha" align="absbottom" title="Ver Calendario" style="border:0; margin:1px;"  src="../img/cal.gif"  id="imgCal'.$id.'" /></a>';
        }
        return 'date';
    }

*/
    public static function inputAttributes($value, $field){
        $attribute['class']="date ";
        return $attribute;
    }
}
?>
