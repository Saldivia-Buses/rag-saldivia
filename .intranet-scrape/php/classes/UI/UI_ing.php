<?php
/* 
 * 2009-09-09
 * help popup class 
 */

class UI_ing extends UI_abm {

/**
 * User Interfase constructor
 *
 */
    var $uidAutoFiltros;
    var $registros;
    public function __construct(&$Datacontainer) {
        parent::__construct($Datacontainer);
        
        $this->disabledCheckDefault = false;

        $this->rowDeleteButton = true;
        $this->enableCheckToggle = true;
        $Datacontainer->unserializeParent = 'true';


    }


    // render de complete XML
    public function show($idFormulario = '', $divcont='', $opt='') {
        $salida = '';

        $id = 'Show'.$this->Datos->idxml;

        // id del contenedor (creo)
        $id2= str_replace('.', '_',($divcont != '')?$divcont:$id );

        $style = $this->Datos->style;

        //$clase = 'consultaing';
        $clase = 'ParentingClass';
        if ($this->Datos->clase != '') $clase = $this->Datos->clase;

        if (isset ($this->Datos->CabeceraMov)){
            foreach ($this->Datos->CabeceraMov as $NCabecera => $ContCab) {
                
                $idxmlHeader = ' header="'.$ContCab->idxml.'" ';
            }
        }

        $this->Datos->barraDrag ='false';

        $salidaDatos =  $this->showTabla('NoSql');

        // Si se define explicitamente una clase en el xml
        if ($this->Datos->clase != '') {
            $clase = $this->Datos->clase;
        }

        //   $customjs    = 'Histrix.registerTableEvents(\'TablaInterna'.$this->Datos->idxml.'\');';

        $retorno= '';
        if ($this->Datos->campoRetorno != '') {
            $uidRetorno = $this->Datos->getCampo($this->Datos->campoRetorno)->uid;
            $retorno = ' origen="'.$uidRetorno.'" ';
        }
        if ($this->Datos->__inline == 'true') {
            $salida .= $salidaDatos;
        }
        else {
            $salida .= '<div  class="'.$clase.'" id="'.$id.'" style="'.$style.'" '.$retorno.' '.$idxmlHeader.'>';
            $salida .= '<div class="contewin" >';
            $salida .= $salidaDatos;
            $salida .= '</div>';
            $salida .= '</div>';
        }
        // create Javascript functions
//        $script[]= $customjs;
        //  $script[]= "$('#$id2').draggable({handle:'#dragbar$id2'});";

        $script[]= "Histrix.registroEventos('".$this->Datos->idxml."')";
        $script[]= $this->Datos->getCustomScript();
        $salida .= Html::scriptTag($script);

        return $salida;

    }

    public function editRow($rowNumber) {

        $dataArray = $this->Datos->TablaTemporal->datos();

        //  $salida = '<tr>';
        foreach($dataArray[$rowNumber] as $fieldName => $fieldValue) {
            $field = $this->Datos->getCampo($fieldName);	 // fetch Object
            if (isset($field->noEmpty) && $field->noEmpty == 'true' && $this->Datos->hasValue[$field->NombreCampo] != 'true') {
                $field->Oculto = true;
            }
            if (($field->Oculto))continue;


            $style ='';
            if ($field->noshow == 'true')
                $field->style = 'display:none;';

            if ($field->Formstyle != '')
                $field .= $field->Formstyle;

            if ($field->style != '' || $field->Formstyle != '')
                $style = 'style="'.$field->style.';'.$field->Formstyle.'"';


            $salida .= '<td '.$style.' >';
            $input = $field->renderInput($this, $form, '', $fieldValue, '',  '');
            $salida .= $input;
            $salida .= '</td>';
        }
        //   $salida .= '</tr>';
        return $salida;
    }

    
    protected function deleteColHeader() {
        $output = '';
        if ($this->Datos->deleteRow != 'false')
            $output =  '<th />';
        return $output;

    }


    protected function inlineCrud($idTableForm, $form , $opt, $formini = '', $formfin = '' , $segundaVez = ''){
	if ($this->Datos->sololectura == 'true') return ;

            $output ='<table '.$idTableFrom.' width="100%"  class="form">';
            if ($this->Datos->modificaABM != 'no' &&
                $this->Datos->modificaABM != 'false') {

                $output .= $this->showAbmInLine($form, $segundaVez);
            }
            $output .= $this->showBtnIng($opt);
            $output .='</table>';


            return $output;

    }



    protected function customTotalJavascript($ObjCampo, $value = '') {
        if ($ObjCampo->uid2 != '' || $value != ''){
            $salida = Html::scriptTag('Histrix.calculoTotal($(\'#' . $ObjCampo->uid2 . '\')[0] ,\''.$ObjCampo->uid.'\' , true , {totalvalue: \''.$value.'\', field:\''.$ObjCampo->NombreCampo.'\' , instance:\''.$this->Datos->getInstance().'\'  } );');
            return $salida;
        }
    }
  
    public function updateTotals() {
        $campos = $this->Datos->camposaMostrar();
        foreach ($campos as $nom => $valor) {

            if ($this->Datos->seSuma($valor)) {
                $ObjCampo = $this->Datos->getCampo($valor);                
                $salida .= $this->customTotalJavascript($ObjCampo );
            }
        }
        return $salida;                
    }

    public function showAbmInLine($form = '', $empty='' ) {
	
	if ($this->Datos->sololectura == 'true') return ;

        $subtipo =  $this->Datos->subtipo;
        $initable = '';
        $endtable = '';
        /* recorro los campos */
        // Identifico las filas que contienen los inputs
        $salida = '<tr class="sortbottom" id="TRForm'.$this->Datos->idxml.'">';

        if ($subtipo == 'vertical') {
            $colspan = 'colspan="'.( $this->cantCampos()) . '" ';
            $salida .= '<td '.$colspan.' >';
            $salida .= '<table border="0" cellspacing="0" >';
            $pri = true;
        }
        $previousAutofield = false;
        $campos = $this->Datos->camposaMostrar();
        $cantidad = count($campos);
        $formstyle = '';
        foreach ($campos as $i => $valor) {

            $ObjCampo = $this->Datos->getCampo($valor);	 // fetch Object
            if ($ObjCampo->showInForm=='false') continue;

            if ($empty != '') $ObjCampo->restaurarValores(); // Empty Object

            $style ='';
            if ($ObjCampo->noshow == 'true')
                $ObjCampo->style = 'display:none;';

            if ($ObjCampo->Formstyle != '')
                $formstyle .= $ObjCampo->Formstyle;

            if ($ObjCampo->style != '' || $ObjCampo->Formstyle != '')
                $style = 'style="'.$ObjCampo->style.';'.$ObjCampo->Formstyle.'"';
            if (isset($campos[$i +1]))
                $ProxObj = $this->Datos->getCampo($campos[$i +1]);

            if ($ObjCampo->Parametro['esclave'] == 'true' || $ObjCampo->esClave)
                $esClave = true;

            if (($ObjCampo->esOculto())) {
                continue;
            }

            $modif = '';

            /* Inline autofields generated by a query
             * group inside a table
             */

            if ($ObjCampo->autofield=='true') {
                $previousAutofield = true;
                $autofields[$valor] = $ObjCampo;
                if ($cantidad == $i + 1) {
                    $salida .= $this->autofieldsForm($autofields, $form);
                    $previousAutofield = false;
                }
                continue;
            }

            if ( $previousAutofield == true && ($ObjCampo->autofield !='true' || $cantidad == $i + 1)) {
                $salida .= $this->autofieldsForm($autofields, $form);
                continue;
            }
            $labeltag = '';
                /* valor del campo */
            if ($ObjCampo->noshow != 'true') {

                if ($subtipo == 'vertical' && $ObjCampo->modpos != 'nobr') {
                    if ($pri !== true)
                        $salida .= '</tr cellspacing="0" cellpadding="0" border="0">';
                    $pri = false;
                    $salida .= '<tr>';
                }

                if ($subtipo == 'vertical' && $ObjCampo->Etiqueta != '') {

                    if ($ObjCampo->sincelda != 'true') {

                        $labeltag = $this->fieldLabel($ObjCampo);
                        $salida .= $labeltag;
                        $labeltag = '';
                    }

                }
                $colspan2 = 'colspan="'.$ObjCampo->colspan.'"';
                $rowspan2 = (isset($ObjCampo->rowspan))?' rowspan="'.$ObjCampo->rowspan.'"':'';

                $salida .= '<td '.$modif.' class="sortbottom"  '.$colspan2.$rowspan2.$style.'> ';
            }
            else {

                $salida .= '<td class="sortbottom" border="0" size="0" '.$style.' >';
            }
            $valordelCampo= isset($ObjCampo->valor)?$ObjCampo->valor:'';

            if ($labeltag != '') {
                $initable = '<table><tr>'.$labeltag.'</tr><tr><td>';
                $endtable = '</td></tr></table>';
            }

            $input ='';
            if (isset($ObjCampo->contExterno) && $ObjCampo->esTabla ) {

            // No muestro la tabla en el formulario si existe un vinculo para modificarla extarnametne
                if ($ObjCampo->linkint != '') {
                    $input = ' ';
                }
                else {
                    $ObjCampo->refreshInnerDataContainer($this->Datos);
                    $ObjCampo->contExterno->tabindex = $this->Datos->tabindex +10;
                    $ObjCampo->contExterno->esInterno = true;

                    $UI = 'UI_'.str_replace('-', '', $ObjCampo->contExterno->tipo);
                    $abmDatosDet = new $UI($ObjCampo->contExterno);



                    $ObjCampo->contExterno->xmlpadre = $this->Datos->xml;
                    $ObjCampo->contExterno->xmlOrig  = $this->Datos->xmlOrig;
                    $style = (isset($ObjCampo->style))?'style="'.$ObjCampo->style.'"':'';
                    $class = (isset($ObjCampo->class))?'class="'.$ObjCampo->class.'"':'';

                    $initable = '<div id="'.$ObjCampo->NombreCampo.'" '.$style.' '.$class.'>';
                    unset($style);
                    unset($class);
                    $endtable = '</div>';
                    $input = $abmDatosDet->showTablaInt(null, $ObjCampo->contExterno->xml,'','false',true, 'noform', null, $ObjCampo );
                    // Increase Tabindex
                    $this->Datos->tabindex += $ObjCampo->contExterno->tabindex;
                }
            }

            if ($input == '')
                $input .= $ObjCampo->renderInput($this, $form, '', $valordelCampo, '', $ProxObj, '');

            $salida .= $initable;
            $salida .= $input;
            $salida .= $endtable;
            unset($initable);
            unset($endtable);
            if ($ObjCampo->noshow != 'true') {
                $salida .= '</td>';
            }
            else $salida .= '</td>';

        }
        $salida .= '<td style="display:none;"><input type="hidden" name="Nro_Fila" campo="Nro_Fila" id="Nro_Fila"/></td> ';


        if ($subtipo == 'vertical') {
            $salida .= '</tr>';
            $salida .= '</table >';
            $salida .= '</td  >';

        }

        $salida .= '</tr>';

        // if removed datefield loose helper
        $script[]= "Histrix.registroEventos('".$this->Datos->idxml."')";
        $salida .= Html::scriptTag($script);


        return $salida;
    }



    protected function addRowCell() {
        if ($this->Datos->deleteRow != 'false')
            $output = '<td class="sintotal" />';
        return $output;
    }

    protected  function rowButtons($orden) {
        if (isset($this->rowDeleteButton) && $this->rowDeleteButton == true && $this->Datos->noForm != 'true') {

            if ($this->Datos->deleteRow != 'false') {
            
                $closeImage = new Html_image( '../img/remove2.png',$this->i18n['deleteRow'], $this->i18n['deleteRow']);
                $closeImage->addEvent('onclick', 'deleterow('.$orden.' , \'Form'.$this->Datos->idxml.'\' '.', \''.$this->Datos->xml.'\' , \''.$this->Datos->xmlOrig.'\', this);');
                $imgcerrar = $closeImage->show();

                if ($this->Datos->editable=='true') {
                //$imgedit = '<img rel="edit" src="../img/edit.png" onClick="Histrix.editRow(this);" alt="Editar Fila" title="Editar Fila"/>';
                    $editImage = new Html_image('../img/edit.png',$this->i18n['editRow'], $this->i18n['editRow'] );
                    $editImage->addEvent('onclick', 'Histrix.editRow(this);');
                    $imgedit = $editImage->show();
                }

                //    if ($this->Datos->deleteRow != 'false')
                $TableData = '<td class="delrow" campo="Nro_Fila" valor="'.$orden.'">'.$imgedit.$imgcerrar.$this->Datos->noForm.'</td>';
            }
        }
        return $TableData;
    }

}

?>