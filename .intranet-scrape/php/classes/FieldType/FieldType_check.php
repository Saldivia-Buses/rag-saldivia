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
class FieldType_check extends FieldType{

    const ALIGN = 'center'; // Default Alignment
    const DIR   = 'ltr';  // Text direction


    public function __construct(&$field=null){
        $this->field = $field;
    }

    public static function customValue($value, $field, $parameters='')
    {
    		if (isset($parameters['order']))
    				$orden = $parameters['order'];
    		$this_field = $field;	
    		$nomcampo = $this_field->NombreCampo;
            $idchk = 'chk'.$nomcampo.'_'.$orden;

            $Container = $this_field->_DataContainerRef;
            
            $refresco = (isset($this_field->refresh) && $this_field->refresh)?'true':'false';
            $jsch = 'onClick="setCampoTabla('.$orden.' , \''.$nomcampo.'\', this , \''.$Container->xml.'\', '.$refresco.', \''.$Container->xmlOrig.'\',  \''.$Container->getInstance().'\');"';

            if ($value == 1) $checked= 'checked="true"';
            	else $checked= '';

            $chkdisabled= '';
            if ($UI->disabledCheckDefault == true)
                $chkdisabled = ' disabled="disabled" ';

            if ( $this_field->deshabilitado == 'true')
                $chkdisabled = ' disabled="disabled" ';

            $valor = '<input '.$chkdisabled.$checked.' refresh="'.$refresco.'" id="'.$idchk.'" type="checkbox" '.$jsch.' >';
            return $valor;

    }


    public static  function inputAttributes($value, $field){
        if ($value == 1) {
            $attribute['checked'] = "true";
            $attribute['default'] = "1";
        } 

        $attribute['type'] = "checkbox";
        
        
        return $attribute;
    }
}


?>
