<?php
/*
 * 2009-09-09
 * help popup class
 */

class UI_abmmini extends UI_abm
{
/**
 * User Interfase constructor
 *
 */
    public function __construct(&$dataContainer)
    {
        parent::__construct($dataContainer);

        $this->disabledCheckDefault = true;
        $this->hasForm = true;
        $this->muestraCant = true;
        $dataContainer->unserializeParent = 'false'   ;
    }

    // render de complete XML
    public function show($idFormulario = '', $divcont='', $opt='')
    {
      /* si es un detalle de otra consulta no pongo el div principal */

        $id = 'Show'.$this->Datos->idxml;

        // id del contenedor (creo)
        $id2= str_replace('.', '_',($divcont != '')?$divcont:$id );

        if ($this->Datos->detalle !='' && $this->Datos->inline != 'true')
            $retrac = true;

        $clasedetalle = 'detalle';

        $style = $this->Datos->style;

        // Columns
        $width = ($this->Datos->width != '')?$this->Datos->width : $this->Datos->ancho;

        if ($width != '') {
            $this->Datos->col1=$width;
            $this->Datos->col2=100 - $width;
            $style.='width:'.$this->Datos->col1.'%;';
        }

        $clase		= 'consultaing2';
        $barraDrag	= false;
        if ($this->Datos->detalle !=''  && $this->Datos->inline != 'true') {
            $clase		= 'consulta';
            
            $barraSlide  = $this->showSlider($id, $retrac);
        }
        $salidaDatos = $this->showTabla();

        // Si se define explicitamente una clase en el xml
        if ($this->Datos->clase != '') {
            $clase = $this->Datos->clase;
        }

        // create Utility dragBar
        if ($this->Datos->barraDrag != 'false') {

            $paramsDrag = $this->dragBarParameters();
            $salidaDrag = $this->barraDrag2($id2,null, $paramsDrag ,$barraDrag, null);
        }

        if ($this->Datos->__inline == 'true') {

            $salida .= $salidaDatos;
        } else {
            if ($this->Datos->campoRetorno != '') {
                $uidRetorno = $this->Datos->getCampo($this->Datos->campoRetorno)->uid;
                $retorno = ' origen="'.$uidRetorno.'" ';
            }

            $salida .=  '<div  class="'.$clase.'" id="'.$id.'" style="'.$style.'" '.$retorno.'>';
            $salida .= $salidaDrag;
            $salida .= '<div class="contewin" >';
            $salida .= $salidaDatos;
            $salida .= '</div>';
            $salida .= '</div>';
        }

        // Incorporo la barra vertical para slide
        $salida .= $barraSlide;

        // Add Detail div
        $salida .= $this->detailDiv($clasedetalle);

        // Graficos
        if ($this->Datos->grafico != '') {
            $salida .= $this->showGraficos();
        }

        $script[]= "sortables_init()";

        //        $script[]= "$('#$id2').draggable({handle:'#dragbar$id2'});";
        $script[]= "Histrix.registroEventos('".$this->Datos->idxml."')";

        $salida .= Html::scriptTag($script);

        return $salida;

    }

    protected function inlineCrud($idTableForm, $form , $opt, $formini = '', $formfin = '')
    {
	if ($this->Datos->sololectura == 'true' ) return;

            $output .= $formini.'<table '.$idTableForm.' width="100%"  class="form" >';
            $output .= $this->showAbmInLine($form);
            $output .= $this->showAbmMini($opt);
            $output .='</table>'.$formfin;

            return $output;

    }

    public function showAbmMini()
    {
    
        $form 	= 'Form'.$this->Datos->idxml;
        $xmlId 	= $this->Datos->xml;
        if ($this->Datos->__inlineid != '') $xmlId = $this->Datos->__inlineid;
        /* BOTONES PARA GRABAR */

        $cantcampos = $this->cantCampos();
        $salida    .= $this->tooltip($cantcampos);

        $salida    .= '<tr class="filabotones">';
        $salida    .= '<td colspan="'.$cantcampos.'" class="btn_mini">';

        $optionsArray['instance'] = '"'.$this->Datos->getInstance().'"';
        $optionsArray['button'] = 'this';

         if ($this->Datos->validations != '') {
            foreach ($this->Datos->validations as $val => $messages) {
                $validations[] = trim(addslashes($val));
            }
            $optionsArray['validations'] = "new Array('" . implode("','", $validations) . "')";
            $optionsArray['messages'] = "new Array('" . implode("','", $this->Datos->validations) . "')";
            $optionsArray['validationOptional'] = "new Array('" . implode("','", $this->Datos->validationType) . "')";
        }

        $postOptions = Html::javascriptObject($optionsArray,'"');
        $postOptions = htmlspecialchars($postOptions);
        
        //$postOptions = htmlspecialchars(json_encode($optionsArray));


        // add toolbar buttons
        if ($this->toolbarButtons != '') {
            $salida .= implode('', $this->toolbarButtons);
        }


        $btn_delete_habil = false;
        $creamod= $this->i18n['create'].'&nbsp;';
        $btnGrabar = new Html_button($creamod, "../img/filesave.png" ,$creamod );
        $btnGrabar->addParameter('idform', $this->Datos->idxml);
        $btnGrabar->addParameter('name', 'Grabar');
        $btnGrabar->addParameter('accion', 'insert');
        $btnGrabar->addEvent('onclick', 'grabaABM('."'".$form."'".', '."this.getAttribute('accion')".' '.", '".$xmlId."' ".' , \''.$this->Datos->xml.'\' , null, null, '.$postOptions.' )');
        $btnGrabar->tabindex = $this->tabindex();
        $salida .= $btnGrabar->show();

        if ($this->Datos->borra != 'no' && $this->Datos->borra != 'false') {
            $btnDelete = new Html_button($this->i18n['delete'], "../img/edittrash.png" ,$this->i18n['delete']);
            if ($btn_delete_habil !=  true)
                $btnDelete->addParameter('disabled', "disabled");
            $btnDelete->addParameter('idform', $this->Datos->idxml);
            $btnDelete->addParameter('name', 'delete');
            $btnDelete->addEvent('onclick', 'grabaABM('."'".$form."'".', '."'delete'".' '.", '".$xmlId."' ".' , \''.$this->Datos->xml.'\' , null, null, '.$postOptions.')');
            $btnDelete->tabindex = $this->tabindex();
            $salida .= $btnDelete->show();

        }
        if ($this->Datos->cancel != 'false') {
            $btnCancel = new Html_button($this->i18n['cancel'], "../img/button_cancel.png" ,$this->i18n['cancel'] );
            $btnCancel->addParameter('idform', $this->Datos->idxml);
            $btnCancel->addParameter('name', $this->i18n['cancel']);
            $btnCancel->addParameter('accion', 'cancel');
            $btnCancel->addEvent('onclick', 'Histrix.clearForm(\''.$this->Datos->xml.'\', true, this)');
            $btnCancel->tabindex = $this->tabindex();
            $salida .= $btnCancel->show();
        }

        $salida .= '</td>';
        $salida .= '</tr>';

        return $salida;

    }

    public function showAbmInLine($form = '', $empty='' )
    {
        $subtipo =  $this->Datos->subtipo;

        /* recorro los campos */
        // Identifico las filas que contienen los inputs
        $salida .= '<tr class="sortbottom" id="TRForm'.$this->Datos->idxml.'">';

        if ($subtipo == 'vertical') {
            $colspan = 'colspan="'.( $this->cantCampos()) . '" ';
            $salida .= '<td '.$colspan.' >';
            $salida .= '<table border="0" cellspacing="0" >';
            $pri = true;
        }
        $previousAutofield = false;
        $campos = $this->Datos->camposaMostrar();
        $cantidad = count($campos);
        foreach ($campos as $i => $valor) {

            $objCampo = $this->Datos->getCampo($valor);	 // fetch Object
            if ($objCampo->showInForm=='false') continue;

            if ($empty != '') $objCampo->restaurarValores(); // Empty Object

            $style ='';
            if ($objCampo->Parametro['noshow'] == 'true')
                $objCampo->style = 'display:none;';

            if ($objCampo->style != '' || $objCampo->Formstyle != '')
                $style = 'style="'.$objCampo->style.';'.$objCampo->Formstyle.'"';

            $ProxObj = $this->Datos->getCampo($campos[$i +1]);

            if (($objCampo->esOculto())) {
                continue;
            }

            $modif = '';

            /* Inline autofields generated by a query
             * group inside a table
             */

            if ($objCampo->autofield=='true') {
                $previousAutofield = true;
                $autofields[$valor] = $objCampo;
                if ($cantidad == $i + 1) {
                    $salida .= $this->autofieldsForm($autofields, $form);
                    $previousAutofield = false;
                }
                continue;
            }

            if ( $previousAutofield == true && ($objCampo->autofield !='true' || $cantidad == $i + 1)) {
                $salida .= $this->autofieldsForm($autofields, $form);
                continue;
            }

                /* valor del campo */
            if ($objCampo->Parametro['noshow'] != 'true') {

                if ($subtipo == 'vertical' && $objCampo->modpos != 'nobr') {
                    if ($pri !== true)
                        $salida .= '</tr cellspacing="0" cellpadding="0" border="0">';
                    $pri = false;
                    $salida .= '<tr>';
                }

                if ($subtipo == 'vertical' && $objCampo->Etiqueta != '') {

                    if ($objCampo->sincelda != 'true') {
                    $salida .= $this->fieldLabel($objCampo);
                    }

                }

                $colspan2 = (isset($objCampo->colspan))?' colspan="'.$objCampo->colspan.'"':'';
                $rowspan2 = (isset($objCampo->rowspan))?' rowspan="'.$objCampo->rowspan.'"':'';

                $salida .= '<td '.$modif.' class="sortbottom" '.$colspan2.$rowspan2.$style.'> ';
            } else {

                $salida .= '<td class="sortbottom" border="0" size="0" '.$style.' >';
            }
            $valordelCampo= $objCampo->valor;

            if ($labeltag != '') {
                $initable = '<table><tr><td>'.$labeltag.'</td></tr><tr><td>';
                $endtable = '</td></tr></table>';

            }

            $input ='';
            if (isset($objCampo->contExterno) && $objCampo->esTabla ) {

            // No muestro la tabla en el formulario si existe un vinculo para modificarla extarnametne
                if ($objCampo->linkint != '') {
                    $input = ' ';
                } else {
                    $objCampo->refreshInnerDataContainer($this->Datos);
                    $objCampo->contExterno->tabindex = $this->Datos->tabindex +10;
                    $objCampo->contExterno->esInterno = true;
                    $UI = 'UI_'.str_replace('-', '', $objCampo->contExterno->tipo);
                    $abmDatosDet = new $UI($objCampo->contExterno);
                    //$abmDatosDet->esInterno = true;

                    $objCampo->contExterno->xmlpadre = $this->Datos->xml;
                    $objCampo->contExterno->xmlOrig  = $this->Datos->xmlOrig;
                    $initable = '<div id="'.$objCampo->NombreCampo.'">';
                    $endtable = '</div>';
                    $input = $abmDatosDet->showTablaInt(null, $objCampo->contExterno->idxml,'','false',true, 'noform', null, $objCampo );
                    // Increase Tabindex
                    $this->Datos->tabindex += $objCampo->contExterno->tabindex;
                }
            }

            if ($input == '')
                $input = $objCampo->renderInput($this, $form, '', $valordelCampo, '',  '');

              // add method to show intput on toolbar
            if (strpos($objCampo->display, 'toolbar') !== false) {
                $this->toolbarButtons[] = $input;
                unset($input);
            }

            $salida .= $initable;
            $salida .= $input;
            $salida .= $endtable;

            if ($objCampo->Parametro['noshow'] != 'true') {
                $salida .= '</td>';
            } else $salida .= '</td>';

        }
        $salida .= '<td style="display:none;"><input type="hidden" name="Nro_Fila" campo="Nro_Fila" id="Nro_Fila"/></td> ';

        if ($subtipo == 'vertical') {
            $salida .= '</tr>';
            $salida .= '</table >';
            $salida .= '</td  >';

        }

        $salida .= '</tr>';

        $script[]=  "Histrix.registroEventos('".$this->Datos->idxml."');";
        $script[]= $this->Datos->customScript;
        $salida .= Html::scriptTag($script);

        return $salida;
    }

}
