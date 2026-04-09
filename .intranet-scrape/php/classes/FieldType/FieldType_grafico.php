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
class FieldType_grafico extends FieldType{

    const ALIGN = 'left'; // Default Alignment
    const DIR   = 'ltr';  // Text direction


    public function __construct(&$field=null){
        $this->field = $field;
    }

    public static function customValue($valor, &$ObjCampo = '' , $renderParameters =''){
	            $nomcampo = $ObjCampo->NombreCampo;


    	        if (isset($ObjCampo->max)){
                    $max = $ObjCampo->max;
                    $min = 0;
                    $showval=1;
                }
                else  {
                    $dataContainer = $ObjCampo->_DataContainerRef;

                    $max = $dataContainer->TablaTemporal->getMax($nomcampo);
                    $min = $dataContainer->TablaTemporal->getMin($nomcampo);
                }
                $styleIMG=($ObjCampo->style != '' )?'style='.$ObjCampo->style:'';

                if ($ObjCampo->grafico == 'barcode' || $ObjCampo->encode != '') {
                    $encode=($ObjCampo->encode !='')?$ObjCampo->encode:'I25';
                    $imageheight = ($ObjCampo->imageHeight !='')?$ObjCampo->imageHeight:30;
                    if ($valor !='')
                        $valor = '<img '.$styleIMG.' src="../lib/barcode/barcode.php?encode='.$encode.'&bdata='.$valor.'&height='. $imageheight.'&scale=1.5&bgcolor=%23FFFFEC&color=%23333366&file=&type=png" alt="Codigo de Barras">';
                }
                else {
                
            	    $c1 = (isset($ObjCampo->color1))?'&c1='.$ObjCampo->color1:'';
            	    $c2 = (isset($ObjCampo->color2))?'&c2='.$ObjCampo->color2:'';
            	    
                    //$valor = '<img '.$styleIMG.' src="grafico_inline.php?showval='.$showval.$c1.$c2.'&clase=grHoriz&amp;ancho=100&amp;alto=15&amp;max='.$max.'&amp;min='.$min.'&valor='.$valor.'&uid='.UID::getUID('', true).'"  />';
                    $valor = '<img '.$styleIMG.' src="grafico_inline.php?showval='.$showval.$c1.$c2.'&clase=grHoriz&amp;ancho=100&amp;alto=15&amp;max='.$max.'&amp;min='.$min.'&valor='.$valor.'"  />';//
		}
		
	return $valor;

    }


}
?>
