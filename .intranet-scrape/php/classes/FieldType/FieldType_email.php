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
class FieldType_email extends FieldType_varchar{

    const ALIGN = 'left'; // Default Alignment
    const DIR   = 'ltr';  // Text direction
    const TYPE  = 'email';


    public function __construct(&$field=null){
        $this->field = $field;
    }


    public static function customValue($value, $field, $parameters='')
    {
                if ($value != '' ){
                    if (filter_var($value, FILTER_VALIDATE_EMAIL) !== false) {
                        $link = 'mailto:'.$value;
                        $icono = '<img align="top" border="0" src="../img/mimetypes/mail_generic.png" />';
                        $value = '<a href="'.$link.'" target="download">'.$icono.'</a>'.$value;
                    }
                    return $value;
                }
    }

    public static function inputAttributes($valor, $field){
        $attribute['type'] = self::TYPE;
        return $attribute;
    }
}
?>
