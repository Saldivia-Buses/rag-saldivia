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
class FieldType_file extends FieldType{

    const ALIGN = 'center'; // Default Alignment
    const DIR   = 'ltr';  // Text direction
    const INPUT   = 'file';  // input type


    public function __construct(&$field=null){
        $this->field = $field;
    }

    public static function customValue($value, $field, $parameters='')
    {
        if ($value != '') {
        	
			$orden = $parameters['order'];

        	$Container = $field->_DataContainerRef;

            $file  = new Archivo($value,'','' ,$field, $Container);

            $link  = $field->url.$file->link;
            $icono = $file->showIcono();
 

            // Add Object Path
            if ($field->path !='') {
                $pathField = $Container->getCampo($field->path);

                if (is_object($pathField)) {
                    $objUrl = $pathField->ultimo;
                } else {
                    $objUrl = $field->path;
                }

            }
            $value = $file->showInline($orden, $objUrl);
        }
        return $value;
    }	

    public static function renderInput( $valor, $field, $arrayAtributos, $uiClass, $opciones=''){
    




        if ($field->FType != '' ) $tipoFile = $field->FType;
        if ($field->url != '' )   $url = $field->url;

        if ($uiClass->Datos->path != '') {

            // Path for uploading files
            $url .= $uiClass->Datos->path.'/';

            // Create choice button
            $btnBrowse = new Html_button("",'../img/imgfolder.gif' ,"Elegir" );

            if ($field->browse != '') {
                $btnBrowse = new Html_button("Archivo", null ,"Archivo" );
                $btnBrowse->addParameter('pathObj','');
            }

            // Add Object PathrowId
            if ($field->path !='') {
                $pathField = $uiClass->Datos->getCampo($field->path);
                if (is_object($pathField)) {
                    $objUrl = $pathField->valor;
                } else {
                    $objUrl = $field->path;
                }

                if ($objUrl !='') {
                    $url .= $objUrl.'/';
                } else {
                    $btnBrowse->addParameter('pathObj',$pathField->uid2);
                }
            }

	    $uniqid2 = $field->uid2;
            $btnBrowse->addParameter('path',$url);
            $btnBrowse->addEvent('onclick', 'BrowseServer2(\''.$uniqid2.'\', \''.$tipoFile .'\', \''.$url.'\',\''.$uniqid2.'\', this)');
        }

        $arrayAtributos['imgIcon']=UID::getUID('img');

        if ($valor != '') {
            $arrayImgs = unserialize($valor);
            //special case
            $spValue = unserialize(urldecode($_GET[$field->NombreCampo]));
            if (is_array($spValue)) {
                $valor = $spValue;
                $arrayImgs = $spValue;
            }
            if (!is_array($arrayImgs)) {
                $arrayImgs[$valor]= $valor;
            }

            foreach ($arrayImgs as $valor => $label) {

                $file = new Archivo($valor,'','');

                $link = $field->url.$file->link;
                if (!is_file($link))
                    $link = $file->link;

                $icono = $file->showIcono();

                $ancho=($field->Size != '')?$field->Size:50;
                if ($file->imagen) {

                    if ($uiClass->Datos->path != '') {
                        $filePath .= $uiClass->Datos->path;
                    }

                    if ($field->path !='') {
    //                                $objUrl =
                        $pathField = $uiClass->Datos->getCampo($field->path);
                if (is_object($pathField)) {
                            $objUrl = $pathField->valor;
                } else {
                    $objUrl = $field->path;
                    }

                    $filePath .= '/'.$objUrl;
                }

                    if (!is_file($link))
                        $link  = '../database/'.$_SESSION['datapath'].'xml/'.$filePath.'/'.$link;

                    $browsePref = '';
                    $link_alt = basename($link);
                    $icoFile .= '<img   style="float:left;" id="'.$arrayAtributos['imgIcon'].'" src="thumb.php?url='.$link.'&ancho='.$ancho.'" title="'.$valor.'" alt="'.$link_alt.'"/>';

                } else {

                    if (file_exists($_SERVER["DOCUMENT_ROOT"].$link)) {
                        $downloadLink = $file->downloadLink();
                        $icoFile .= '<a href="'.$downloadLink.'" target="_blank">'.$icono.$valor.'</a>';

                    }
                }
            }

        } else {

            if ($field->url == '') {
                $filePath= $uiClass->Datos->path;
            }
            // Add Object Path
            if ($field->path !='') {
                $pathField = $uiClass->Datos->getCampo($field->path);
                if (is_object($pathField)) {
                    $objUrl = $pathField->valor;
                } else {
                    $objUrl = $field->path;
                }

                $filePath .= '/'.$objUrl;
            }

            $link  = '../database/'.$_SESSION['datapath'].'xml/'.$filePath.'/'.$link;
            $link2 = '../database/'.$_SESSION['datapath'].'xml/'.$filePath.$nonutflink;

            $imageSelect = '$(\'#'.$arrayAtributos['imgIcon'].'\').attr(\'src\', \'thumb.php?alto=40&url='.$link.'\' + this.value );';
            $file = new Archivo('vacio','','', $field, $uiClass->Datos);
            $file->imgId = $arrayAtributos['imgIcon'];
            $icoFile = $file->showIcono();

        }




        $btnBrowse->addParameter('style', 'float:left;');
        $browse = '';
        $dropclassname = '';
        $legend = '';
        $clrbtntxt = '';
        
        if ($field->deshabilitado != 'true') {
            $browse = $btnBrowse->show();
            $dropclassname = 'dropfile';
            $legend = $uiClass->i18n['dropfile'];

            // clear button
            $clrbtn = new Html_button('', "../img/cancel.png" ,"Borrar" );
            $clrbtn->addParameter('style', 'float:left; '   );
            $clrevent = '$(\'#'.$arrayAtributos['id'].'\' ).val(\'\'); ';
            $clrevent .= '$(\'#'.$arrayAtributos['imgIcon'].'\').attr(\'src\',\'\'); ';
            $clrbtn->addEvent('onclick', $clrevent );
            $clrbtn->tabindex = $uiClass->tabindex();

            $clrbtntxt =  $clrbtn->show();
        }
         
        $inputBox = new Html_textBox($valor, $field->TipoDato);
        $inputBox->Parameters=$arrayAtributos;
        $inputBox->addParameter('image', true);
        $inputBox->addEvent('onchange', $imageSelect, true);

        $inputBox->value    = $valor;
    //                $inputBox->hide   = 'hidden';
        $inputBox->addParameter('disabled', 'disabled');

	$maxwidth = isset($field->maxWidth)?' maxWidth="'.$field->maxWidth.'"' :'';
        $salida .= '<div class="'.$dropclassname.'" pathObj="'.$pathField->uid2.'" fileinput="'.$arrayAtributos['id'].'" '.$maxwidth.'  destinationDir="'.$url.'" style="height:75px;">';

        $salida .='<div style="width:45px;float:left;">';
        $salida .= $browse;

        $salida .= $clrbtntxt;
        $salida .='</div>';
        $salida .='<div style="float:left;height:60px; width:100px;">';
//        $imgUrl = 'href="'.$file->imageLink() .'"';
        $imgUrl = 'href="thumb.php?url='.$link .'"';
        $salida .= '<a rel="lightbox" '.$imgUrl.'>'.$icoFile.'</a><br/>';
        $salida .= $inputBox->show();
        $salida .='</div><br/>';
        $salida .='<div style="float:left;text-align:center;width:100%;">';
        $salida .= $legend;
        $salida .='</div>';

        $salida .= '</div>';
 
        return $salida;
    }
}
?>
