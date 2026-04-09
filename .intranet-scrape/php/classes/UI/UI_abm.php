<?php

/*
 * 2009-09-09
 * help popup class
 */

class UI_abm extends UI_consulta
{
    /**
     * User Interfase constructor
     *
     */
    public $nosel;
    public $autocomplete;
    public $preFetch;
    public $esAyuda;

    public function __construct(&$Datacontainer)
    {
         parent::__construct($Datacontainer);

        $this->disabledCellId = false;
        $this->disabledCheckDefault = true;

        $this->hasForm = true;
        $this->formClass = 'consulta_der';
        $this->muestraCant = true;

    $this->defaultClass = 'consulta';

        $this->updateButton = (isset($Datacontainer->modifica) && $Datacontainer->modifica == 'false')?false:true;

        if (isset( $this->Datos->hasFieldset))
                    $this->hasFieldset = $this->Datos->hasFieldset;
    }

    // render de complete XML
    public function show($idFormulario = '', $divcont='', $opt='')
    {
        $id = 'Show' . $this->Datos->idxml;

        // id del contenedor (creo)
        $id2 = str_replace('.', '_', ($divcont != '') ? $divcont : $id);

        $style = (isset($this->Datos->style)) ? $this->Datos->style : '';

        // Columns
        $ancho = (isset($this->Datos->ancho)) ? $this->Datos->ancho : null;
        $width = (isset($this->Datos->width)) ? $this->Datos->width : $ancho;

        if ($width != null) {
            $this->Datos->col1 = $width;
            $this->Datos->col2 = 100 - $width;
            $style.='width:' . $this->Datos->col1 . '%;';
        }

        $clasedetalle = 'detalle2';
        $clase = $this->defaultClass;
        $salidaDatos = $this->showTabla();
        $barraSlide = $this->showSlider($id);

        $clase_der = (isset($this->formClass))?'class="'.$this->formClass.'"':'class="consulta_der"';

        $salidaAbm = $this->showAbm(null, $clase_der);

        // Si se define explicitamente una clase en el xml
        if ($this->Datos->clase != '') {
            $clase = $this->Datos->clase;
        }

        $retorno = '';
        if ($this->Datos->campoRetorno != '') {
            $uidRetorno = $this->Datos->getCampo($this->Datos->campoRetorno)->uid;
            $retorno = ' origen="' . $uidRetorno . '" ';
        }

        $salida = '<div  class="' . $clase . '" id="' . $id . '" style="' . $style . '" ' . $retorno . '>';
        $salida .= '<div class="contewin" >';
        $salida .= $salidaDatos;
        $salida .= '</div>';
        $salida .= '</div>';

        // El Abm
        $salida .= $salidaAbm;
        // Incorporo la barra vertical para slide
        $salida .= $barraSlide;

        // Add Detail div
        $salida .= $this->detailDiv($clasedetalle);

        // create Javascript functions
        $script[] = "Histrix.registroEventos('" . $this->Datos->idxml . "')";
//        $script[] = "Histrix.calculoAlturas('" . $this->Datos->idxml . "', null ); ";
        $script[]= $this->Datos->customScript;

        $salida .= Html::scriptTag($script);

        return $salida;
    }

    /**
     * Show input FORM Containder DIVs
     */
    public function showAbm($modoAbm = '', $clase = '')
    {
        if ($clase != '')
            $class = $clase;
        else {
            if ($modoAbm == 'readonly')
                $class = 'class="Fichadetalle"';
            else
                $class = 'class="Ficha2"';

            if (isset($this->defaultClass)) {
                $class = 'class="'.$this->defaultClass.'"';
            }

        }
        $idContenedor = $this->Datos->idxml;

        $style = '';
        $strStyle = '';
        if ($this->Datos->col2 != '')
            $style .='width:' . ($this->Datos->col2 - 0.5) . '%;';

        if ($style != '')
            $strStyle = ' style="' . $style . '" ';

        $salida = '<div ' . $class . ' id="DIVFORM' . $idContenedor . '" ' . $strStyle . ' >' .
                '<div class="contewin" id="INT' . $idContenedor . '">';
        $salida .= $this->showAbmInt($modoAbm, 'INT' . $idContenedor);
        $salida .= '</div></div>';

        return $salida;
    }

    /**
     * Fieldlabel
     * creates a field label
     */
    protected function fieldLabel($field)
    {


        $field->idlbl = UID::getUID('', true);
        $uid2 = (isset($field->uid2)) ? $field->uid2 : '';


        $output = $field->renderLabel($uid2);

        return $output;

    }

   protected function topButtons(){}

   private function formSearchButtons()
   {
        $btnHlp = new Html_button($this->i18n['search'], "../img/find.png", $this->i18n['search']);
        $btnHlp->addParameter('name', $this->i18n['search']);
        $btnHlp->addParameter('accesskey', "(F2)");
        $btnHlp->addEvent('onclick', 'ayudaFicha(this)');
        $htmlhelp = $btnHlp->show();

        $btnCancel = new Html_button($this->i18n['cleanRecord'], "../img/eraser.png", $this->i18n['cancel']);
        $btnCancel->addParameter('name', $this->i18n['cancel']);
        $btnCancel->addParameter('accesskey', "(F4)");
        $btnCancel->addEvent('onclick', 'Histrix.clearForm(\'' . $this->Datos->xml . '\', true, this, {instance:\''.$this->Datos->getInstance().'\'})');
        $htmlCancel .= $btnCancel->show();
        $output = Html::tag('div', $htmlhelp . $htmlCancel, array('class' => 'barraBusqueda'));

        return $output;
   }
    /**
     * Show FORM Content
     */
    public function showAbmInt($modoAbm = '', $idContenedor = '', $opt='')
    {
        $onkeydown = '';

        if ($idContenedor == '')
            $idContenedor = 'INT' . $this->Datos->xml;

        $form = 'Form' . $this->Datos->idxml;
        $form = str_replace('.xml', '', $form);

        // si es una ficha se autoreferencia
        if ($this->tipo == 'ficha' || $this->tipo == 'cabecera')
            $contenedorDestinoSQL = $idContenedor; //'Ficha';
        else
            $contenedorDestinoSQL = $this->Datos->xml; //'TablaConsultada';

            $rsPrivado = $this->Datos->resultSet;

        $salida = '';

        // Add Search Buttons to Form
        if (isset($this->Datos->conBusqueda) && $this->Datos->conBusqueda == 'true') {
            $onkeydown = ' searchForm="true" ';
            $salida .= $this->formSearchButtons();
        }

         $salida .= $this->topButtons();

        if ($this->getTitulo() != '') {

            // Help link to the manual web page.

            $helpId  = $this->Datos->getHelpLink();
            $helpLink ='<a target="_blank" class="internalLink" href="'.$helpId.'" >';

            $obs = (isset($this->Datos->obs)) ? $this->Datos->obs : '';
            $titulo = '<legend title="' . $obs . '">'.$helpLink . htmlentities(ucfirst($this->getTitulo()), ENT_QUOTES, 'UTF-8') . '</a> <b style="font-size:20px;">'.$obs.'</b></legend>';
        }

        if (!isset($this->hasFieldset) || $this->hasFieldset != 'false') {
            $salida .= '<fieldset>'. $titulo;
        }

       $orig = '';
        if (isset($this->Datos->xmlOrig))
            $orig = 'original="' . $this->Datos->xmlOrig . '"';

        $salida .= '<form  id="' . $form . '" onsubmit="return false;" xml="' . $this->Datos->xml . '"  maininstance="' . $this->Datos->mainInstance . '" instance="' . $this->Datos->getInstance() . '" xmlOrig="' . $this->Datos->xmlOrig . '" action="" ' . $onkeydown . ' name="' . $form . '" ' . $orig . ' tipo="' . $this->Datos->tipoAbm . '">';

        $salida .= '<table width="100%" class="form" cellspacing="0"> ';
        $styleTbody = '';
//        if (isset($this->Datos->formTbody))
//            $styleTbody = 'style="' . $this->Datos->formTbody . '"';
        $tbody = '<tbody id="tformbody_' . $this->Datos->idxml . '" ' . $styleTbody . ' >';

        /* recorro los campos */
        if (isset($rsPrivado)) {
            $row = _fetch_array($rsPrivado);
            if (($row))
                foreach ($row as $clarow => $valrow) {
                    $Campo = $this->Datos->getCampo($clarow);
                    if ($Campo != null)
                        $Campo->setValor($valrow);
                }
        }

        $rs = $this->Datos->resultSet;
        if (isset($rs))
            $campos = _num_fields($rs);
        else
            $campos = $this->cantCampos();
        $i = 0;
        $nombres = $this->Datos->nomCampos();
        $pri = false;

        while ($i < $campos) {
            $i++;
            if ($rs) {
                $nom = _field_name($rs, $i);
                if ($i < $campos)
                    $nomprox = _field_name($rs, $i + 1);
                else
                    $nomprox = '';
            } else {
                $nom = $nombres[$i - 1];
                if ($i < $campos)
                    $nomprox = $nombres[$i];
                else
                    $nomprox = '';
            }

            // busco los campos por la etiqueta
            if ($this->Datos->tablas[$this->Datos->TablaBase] != '') {
                $nombre_campo = '';
                $nombre_campo_prox = '';

                if (isset($this->Datos->tablas[$this->Datos->TablaBase]->etiquetas_reverse[$nom]))
                    $nombre_campo = $this->Datos->tablas[$this->Datos->TablaBase]->etiquetas_reverse[$nom];

                if (isset($this->Datos->tablas[$this->Datos->TablaBase]->etiquetas_reverse[$nomprox]))
                    $nombre_campo_prox = $this->Datos->tablas[$this->Datos->TablaBase]->etiquetas_reverse[$nomprox];

                if ($nombre_campo_prox != '')
                    $nomprox = $nombre_campo_prox;
                if ($nombre_campo != '')
                    $nom = $nombre_campo;
            }

            if ($nomprox != '')
                $proximoObjeto = $this->Datos->getCampo($nomprox);

            $objNombre = $this->Datos->getCampo($nom);

            if (isset($objNombre->valAttribute) && $objNombre->valAttribute != '') {
                foreach ($objNombre->valAttribute as $attribID => $attrib) {

                    $valAtt = $this->Datos->getCampo($attrib)->valor;

                    if ($attribID == 'oculto') {
                        if ($valAtt === 'true' || $valAtt === true)
                            $objNombre->setOculto($valAtt);
                    }
                    $objNombre->{$attribID} = (string) $valAtt;
                    $objNombre->Parametro[$attribID] = (string) $valAtt;
                }
            }

            // Bandera que marca si es un campo clave o no
            if ((isset($objNombre->Parametro['esclave']) && $objNombre->Parametro['esclave'] == 'true') ||
                    (isset($objNombre->esClave) && $objNombre->esClave))
                $esClave = true;

            // No muestro los campos Ocultos
            if ($objNombre != null)
                if (($objNombre->esOculto())) {
                    continue;
                }

            $modpos = (isset($objNombre->modpos)) ? $objNombre->modpos : '';

            if ($modpos != 'nobr') {
                if ($pri)
                    $tbody .= '</tr >';
                $tbody .= '<tr >';
                $pri = true;
            }
            $style = '';
            $none = '';

            // Si el campo activa otros campos

            if (isset($objNombre->Parametro['noshow']) && $objNombre->Parametro['noshow'] == 'true') {
                $none = 'display:none;';
                $style = 'style="' . $none . '"';
            }

            $tmpStyle = (isset($objNombre->style)) ? $objNombre->style : '';

            $formStyle = (isset($objNombre->Formstyle)) ? $objNombre->Formstyle : '';
            $formStyle .= (isset($objNombre->formStyle)) ? $objNombre->formStyle : '';

            $style = 'style="' . $none . $tmpStyle .$formStyle. '"';
            unset($tmpStyle);

            $colspan = (isset($objNombre->colspan) && $objNombre->colspan > 1) ? ' colspan="' . $objNombre->colspan . '" ' : '';
            $rowspan = (isset($objNombre->rowspan) && $objNombre->rowspan > 1) ? ' rowspan="' . $objNombre->rowspan . '" ' : '';

            // Add Field Label
            if (isset($objNombre->Etiqueta) && $objNombre->Etiqueta != '') {
                if (isset($objNombre->contExterno) && isset($objNombre->esTabla) && $objNombre->esTabla && $modpos != 'force') {
                // do not show label
        } else {
                $tbody .= $this->fieldLabel($objNombre);
                }
            }

            // valor del campo
            $nocell = isset($objNombre->sincelda) ? $objNombre->sincelda : '';
            if ($nocell != 'true') {
                $tbody .= '<td ' . $colspan . $rowspan . $style . ' > ';
            }

            $valor = $objNombre->valor;
            if (!is_utf8($objNombre->valor))
                $valor = utf8_encode($objNombre->valor);

            if ($objNombre->TipoDato == 'date' && strpos($valor, '-') == 4) {
                if ($valor == '0000-00-00')
                    $valor = '';
                else
                    $valor = date("d/m/Y", strtotime($valor));
            }

            if (isset($objNombre->aletras) && $objNombre->aletras == true) {
                if (is_numeric($valor))
                    $valor = NumeroALetras($valor);
            }

            if (isset($this->Datos->sololectura) && $this->Datos->sololectura == true) {
                $objNombre->deshabilitado = 'true';
            }

            $currentRow[$nom] = $valor;

            // Si el Campo recorrido tiene dentro un contenedor le cargo los parametros y lo muestro
            $inputText = '';
            if (isset($objNombre->contExterno) && $objNombre->esTabla) {

                $objNombre->refreshInnerDataContainer($this->Datos);

                $objNombre->contExterno->tabindex = (isset($this->Datos->tabindex)) ? $this->Datos->tabindex : 0 + 10;
                $objNombre->contExterno->esInterno = true;
                $UI = 'UI_' . str_replace('-', '', $objNombre->contExterno->tipo);
                $abmDatosDet = new $UI($objNombre->contExterno);

                if (isset($objNombre->titulo))
                    $tbody .= $objNombre->titulo;



                if ($objNombre->contExterno->tipoAbm == 'ing' ||
                        $objNombre->contExterno->tipoAbm == 'grid' ||
                        (isset($objNombre->contExterno->noTabs) && $objNombre->contExterno->noTabs == "true")) {

                    if ($objNombre->contExterno->tipoAbm == 'grid') {
                        $this->Datos->grillasContenidas[$objNombre->contExterno->xml] = $objNombre->contExterno->getInstance();
                    }

                    // DO NOT REMOVE (usefull for temporal tables import)
                    $contReferente = new ContDatos("");
                    if (!$objNombre->contExterno->isInner) { // if is an inner table do not unserialize previous container

                        $anterior = Histrix_XmlReader::unserializeContainer($objNombre->contExterno);

                        if ($anterior != '')
                        $objNombre->contExterno = $anterior;
                    }

                    // do not reload data if data is recovered form a previous container
                    if ($this->Datos->__savedState==true){

//                        $objNombre->contExterno = Histrix_XmlReader::unserializeContainer($objNombre->contExterno);                      
                        $objNombre->contExterno->llenoTemporal = 'false';
                    }

                    $this->Datos->tabindex++;
                    $objNombre->contExterno->tabindex = $this->Datos->tabindex;
                    /*
                    $objNombre->contExterno->esInterno = true;
                    $UI = 'UI_' . str_replace('-', '', $objNombre->contExterno->tipo);
                    $abmDatosDet = new $UI($objNombre->contExterno);
                    */
                    $objNombre->contExterno->xmlpadre = $this->Datos->xml;
                    $objNombre->contExterno->xmlOrig = $this->Datos->xmlOrig;
                    //     $inputText = $abmDatosDet->showTablaInt($opt, $objNombre->contExterno->idxml, null, 'false', true, 'Form'.$this->Datos->xml);
                    $inputText = $abmDatosDet->showTabla($opt);
                    $inputText = '<div name="' . $nom . '" id="' . $nom . '" style="'.$objNombre->objStyle.'">' . $inputText . '</div>';
                    $this->Datos->tabindex = ($objNombre->contExterno->tabindex + 1 );
                } else {
                    $tablabel = ($objNombre->Etiqueta != '')?$objNombre->Etiqueta:$objNombre->contExterno->titulo;
                    $tabsInternos[$tablabel] = $abmDatosDet->showTablaInt($opt, $objNombre->contExterno->idxml, null, 'false', true, 'Form' . $this->Datos->xml);
                }

                Histrix_XmlReader::serializeContainer($objNombre->contExterno);
            } else {
                if (isset($objNombre->linkint) && $objNombre->linkint != '' && $objNombre->deshabilitado != 'false') {

                    $currentRow[$nom] = $valor;
                    $params = $this->generateLinkParameters($objNombre, $currentRow);
//TODO TEST!!!!!!!
                    $inputText = $this->linkButton($objNombre, $valor, '', $params, $formatoCampo);
                   // $customAttrib['hidden'] ='true';

        //            $inputText .= $objNombre->renderInput($this, $form, '', $valor, $modoAbm,  $idContenedor, $customAttrib);

                } else {
                    if (isset($nombres[$i])) {
                        $fieldObject = $this->Datos->getCampo($nombres[$i]);
                    }
                    if (is_object($objNombre))
                        $inputText = $objNombre->renderInput($this, $form, '', $valor, $modoAbm,  $idContenedor, null);
                }
            }

            // add method to show intput on toolbar
            if (isset($objNombre->display) && strpos($objNombre->display, 'toolbar') !== false) {
                $this->toolbarButtons[$nom] = $inputText;
                unset($inputText);
            }

            $tbody .= $inputText;

            if ($proximoObjeto != '') {
                $nocell = isset($proximoObjeto->sincelda) ? $proximoObjeto->sincelda : '';

                if ($nocell != 'true')
                    $tbody .= '</td>';
            } else
                $tbody .= '</td>';
        }

        // Close Last cell
        //    $tbody .= '</td>';
        $pie = false;
        // Tabs Internas
        $tabsbody = '';
        if (isset($tabsInternos) && $tabsInternos != '') {
            $cantTabs = count($tabsInternos);
            $pie = true;
            $tabsbody .= '<tr class="filatabs">';
            $marginTabs = (isset($this->Datos->marginTabs)) ? $this->Datos->marginTabs : '';
            if ($marginTabs != 'false')
                $tabsbody .= '<td class="trans"></td>';

            $tabsbody .= '<td colspan="10" class="celdaTabs" >';

            $tabsbody .= '<div  id="tabs_ficha">';
            $uidtabs = UID::getUID($this->Datos->idxml, true);
            $tabsbody .= '<ul class="tabs" id="' . $uidtabs . '" style="background-color:transparent;">';
            $lis = 0;
            foreach ($tabsInternos as $nombre => $TablaInt) {
                if ($lis == 0)
                    $claseLi = "activo";
                else
                    $claseLi = "inactivo";
                $uidname = str_replace('.', '_', $nombre);
                $uidname = str_replace(' ', '_', $uidname);
                $tabsbody .='<li id="LI' . $uidname . '" class="' . $claseLi . '"><span  onclick="Histrix.activartab(\'' . $uidname . '\', \'' . $uidtabs . '\')">';
                $tabsbody .=$nombre;
                $tabsbody .='</span></li>';
                $lis++;
            }
            $tabsbody .= '</ul></div>';
            $tabsbody .= '<div style="position:relative; float:left; margin:0px ;padding:0px;width:100%;">';

            $lis = 0;
            if ($cantTabs != 1)
                $tabsbody .='<ul class="continternos">';
            foreach ($tabsInternos as $nombre => $TablaInt) {
                $uidname = str_replace('.', '_', $nombre);
                $uidname = str_replace(' ', '_', $uidname);

                if ($lis == 0)
                    $visible = "visible";
                else
                    $visible = 'hidden; display:none';
                if ($cantTabs != 1)
                    $tabsbody .='<li  id="' . $uidname . '" style="visibility:' . $visible . '; height:auto">';

                $tabsbody .= $TablaInt;

                if ($cantTabs != 1)
                    $tabsbody .='</li>';
                $lis++;
            }
            if ($cantTabs != 1)
                $tabsbody .= '</ul>';
            $tabsbody .= '</div>';
            $tabsbody .= '</td></tr>';
        }

        $tbody .= '</tr >';

        if ($this->tipo == 'abm') {
            $tbody .= $tabsbody;
            $tabsbody = '';
        }

        $tbody .= '</tbody>';

        $tfoot = '<tfoot id="tformfoot_' . $this->Datos->idxml . '">';

        /* SAVE BUTTONS */

        if ($modoAbm != 'readonly' &&
                $this->tipo != 'cabecera' &&
                $this->tipo != 'fichaing' &&
                $this->Datos->sololectura != 'true') {

            $pie = true;
            if ($tabsbody != '')
                $prefix = '<td class="trans" ></td>';
            $tfoot .= $this->tooltip(10, $prefix);
        }

        if ($opt != 'noecho')
            if ($modoAbm != 'readonly'
                    && $this->tipo != 'cabecera'
                    && $this->tipo != 'fichaing') {

                $optionButton = $this->btnOptions();
                $tfoot .= '<tr class="filabotones">';
                if ($tabsbody != '')
                    $tfoot .= '<td class="trans" >' . $optionButton . '</td>';

                $tfoot .= '<td colspan="10"   align="center">';

                if ($this->Datos->sololectura != 'true') {
                    $pie = true;
                    $btn_delete_habil = false;
                    if ($this->Datos->modificar == 'true') {
                        $creamod = $this->i18n['save'];
                        $btn_delete_habil = true;
                    } else
                        $creamod=$this->i18n['create'];

                    if ($btn_delete_habil == true)
                        $btn_hab = '';
                    else
                        $btn_hab = "disabled";

                    $defaultAction = 'insert';
                    if (isset($this->Datos->TotalRegistros))
                        $cant = $this->Datos->TotalRegistros;
                    else
                        $cant = $this->Datos->resultSet->num_rows;

                    if ($this->Datos->modificar == 'true') {
                        $creamod = $this->i18n['save'];
                        $defaultAction = 'update';
                    } else {
                        $this->i18n['create'];
                    }

                    if (($this->Datos->preFetch == 'true' || $this->preFetch == true ) && $cant != 0) {
                        $defaultAction = 'update';
                        $creamod = $this->i18n['save'];
                        // $this->Datos->onDuplicateKey = "true";
                    }

                    $optionsArray['instance'] = '"'.$this->Datos->getInstance() .'"';
                    $optionsArray['parentInstance'] = '"'.$this->Datos->parentInstance .'"';
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

                    //$postOptions = Html::javascriptObject($optionsArray);
                    //$postOptions = htmlspecialchars(json_encode($optionsArray));

                    // add toolbar buttons
                    if ($this->toolbarButtons != '') {
                        $tfoot .= implode('',$this->toolbarButtons);
                    }

                    if ($this->Datos->inserta == 'false') {
                        $defaultAction = 'update';
                        $creamod = $this->i18n['save'];
                    }

                    if ($this->updateButton == true || $this->Datos->inserta != 'false') {

                        $creamod = (isset($this->Datos->createButtonLabel)) ? $this->Datos->createButtonLabel : $creamod;
                        $createIcon = (isset($this->Datos->createButtonIcon)) ? $this->Datos->createButtonIcon : 'filesave.png';

                        $btnIns = new Html_button($creamod, "../img/" . $createIcon, $creamod);
                        $btnIns->addParameter('name', 'Grabar');
                        $btnIns->addParameter('accion', $defaultAction);
                        $btnIns->addParameter('idform', $this->Datos->idxml);

                        $btnIns->addEvent('onclick', 'grabaABM(' . "'" . $form . "'" . ', ' . "this.getAttribute('accion')" . ' ' . ", '" . $contenedorDestinoSQL . "' " . ' , \'' . $this->Datos->xml . '\', null, null, ' . $postOptions . ' )');
                        $btnIns->tabindex = $this->tabindex();
                        $tfoot .= $btnIns->show();
                    }
                    if (isset($this->Datos->borra) && ($this->Datos->borra == 'no' || $this->Datos->borra == 'false')) {
			// do not delete
		    } else {
			// delete button by default
                        $btnDelete = new Html_button($this->i18n['delete'], "../img/edittrash.png", $this->i18n['delete']);
                        if ($btn_hab != '' && isset($btnIns))
                            $btnIns->addParameter('disabled', $btn_hab);
                        $btnDelete->addParameter('name', 'delete');
                        $btnDelete->addEvent('onclick', 'grabaABM(' . "'" . $form . "'" . ', ' . "'delete'" . ' ' . ", '" . $contenedorDestinoSQL . "' " . ' , \'' . $this->Datos->xml . '\' , null, null, ' . $postOptions . ' )');
                        $btnDelete->tabindex = $this->tabindex();
                        $tfoot .= $btnDelete->show();
                    }
                    
                    if ($this->Datos->cancel != 'false') {
                        $btnCancel = new Html_button($this->i18n['cancel'], "../img/cancel.png", $this->i18n['cancel']);
                        $btnCancel->addParameter('name', $this->i18n['cancel']);
                        $btnCancel->addParameter('accion', 'cancel');

                        $btnCancel->addEvent('onclick', 'Histrix.clearForm(\'' . $this->Datos->xml . '\', true, this)');
                        $btnCancel->tabindex = $this->tabindex();
                        $tfoot .= $btnCancel->show();
                    }
                    
                }
                $formPrint = 'FormPrintFicha' . $this->Datos->idxml;
                $doPrint = (isset($this->Datos->imprime)) ? $this->Datos->imprime : '';
                if ($doPrint != 'false' && $this->tipo == 'ficha') {
                    $btnImprimir = new Html_button($this->i18n['print'], "../img/printer1.png", $this->i18n['print']);
                    $btnImprimir->addEvent('onclick', 'Histrix.imprimirpdf(\'' . $this->Datos->xml . '\', \'' . $this->Datos->getTitulo() . '\', \'' . $formPrint . '\' ,null, null,\''.$this->Datos->getInstance().'\' )');
                    $btnImprimir->tabindex = $this->tabindex();

                    $tfoot .= $btnImprimir->show();

                    $btnImpOpc = new Html_button(null, "../img/down.png", $this->i18n['print']);
                    $btnImpOpc->addEvent('onclick', 'Histrix.showopprint(this, \'' . $formPrint . '\', false)');
                    $btnImpOpc->addParameter('title', "Parametros de Impresion");
                    //TODO REMOVE ADD AS CLASS
                    $btnImpOpc->addStyle('height', '24px');
                    $btnImpOpc->addStyle('padding-left', '0px');
                    $btnImpOpc->addStyle('padding-right', '0px');
                    $btnImpOpc->addStyle('margin-left', '-2px');
                    $tfoot .= $btnImpOpc->show();
                    $tfoot .= $this->printOptions($formPrint);

                    $pie = true;
                }

                $tfoot .= '</td></tr>';
            }
        $tfoot .= $tabsbody;
        $tfoot .= '</tfoot>';

        $salida .= $tbody;
        if ($pie)
            $salida .= $tfoot;

        $salida .= '</table>';

        $salida .= $this->addFormButtons();

        $salida .= '</form>';

        if (!isset($this->hasFieldset) || $this->hasFieldset != 'false') {
            $salida .= '</fieldset>';
        }

        return $salida;
    }

    protected function addFormButtons()
    {
        return '';
    }

    public function importDataButton()
    {
	$botones2='';
        // Genero el Control para importar datos de otros XML externos
        if (isset($this->Datos->importadatos) && $this->Datos->importadatos != '') {
            $selimpButton = '';
            // Selector de Origen
            $selimp = '<ul class="makeMenu">';
            $selimp .= '<li ><img src="../img/importar.png" alt="Importar" border="0"  />Importar';
            $selimp .= '<ul>';

            $cant = 0;
            foreach ($this->Datos->importadatos as $clavexml => $arrayparams) {
                $xmlimporta = (string) $clavexml;
                $label = (string) $arrayparams['label'];
                $dirimportacion = (string) $arrayparams['dir'];
                $widthimportacion = ((string) $arrayparams['width'] != '')?(string) $arrayparams['width']:'90%';
                $accesskey = (string) $arrayparams['accesskey'];
                $ventanaPadre = $this->Datos->xml;
                if ($this->Datos->xmlOrig != '')
                    $ventanaPadre = $this->Datos->xmlOrig;

                $paramstring = '&amp;';
                if (isset($arrayparams['campos']))
                    foreach ($arrayparams['campos'] as $npar => $fieldName) {
                        /* busco el campo en la/s cabecera/s */
                        $fieldName = (string) $fieldName;
                        if ($fieldName == '')
                            $fieldName = $npar;
                        $paramstring .= '_param_in[]=' . $fieldName . '&amp;';
                    }
                $tit_importa = $arrayparams['titulo'];
                $dirimportacion = ($dirimportacion != '')?$dirimportacion : $this->Datos->dirxml;
                $paramstring .='&amp;_xmlreferente=' . $this->Datos->xml;
                $paramstring .='&amp;parentInstance=' . $this->Datos->getInstance();

                $paramstring .='&dir=' . $dirimportacion;
                $esBoton = '';

                $esBoton = (string) $arrayparams['boton'];
                if ($esBoton == 'true') {

                    $btnImportar2 = new Html_button($label, "../img/importar.png", $label);
                    $btnImportar2->addParameter('acceskey', $accesskey);
                    $btnImportar2->addParameter('name', 'Importar');
                    $btnImportar2->addEvent('onclick', 'Histrix.ventInt(\'' . $ventanaPadre . '\' , \'' . $xmlimporta . '\', \'' . $paramstring . '\', \'' . $tit_importa . '\', {modal:true, height: \'95%\', width:\'99%\', parentInstance: \''.$this->Datos->getInstance().'\'})');
                    $btnImportar2->tabindex = $this->tabindex();
                    $selimpButton .= $btnImportar2->show();

                    continue;
                }
                $cant++;

                $btnImportar = new Html_button($label, "../img/importar.png", $label);
                $btnImportar->addParameter('acceskey', $accesskey);
                $btnImportar->addParameter('name', 'Importar');
                $btnImportar->addEvent('onclick', 'Histrix.ventInt(\'' . $ventanaPadre . '\' , \'' . $xmlimporta . '\', \'' . $paramstring . '\', \'' . $tit_importa . '\', {modal:true, height: \'95%\', width:\'95%\', parentInstance: \''.$this->Datos->getInstance().'\'})');
                $btnImportar->tabindex = $this->tabindex();
                $botones2 = $btnImportar->show();

                $selimp .= '<li><div class="itemmenu" onClick="Histrix.ventInt(\'' . $ventanaPadre . '\' , \'' . $xmlimporta . '\', \'' . $paramstring . '\', \'' . $tit_importa . '\', {modal:true, width:\''.$widthimportacion.'\',parentInstance: \''.$this->Datos->getInstance().'\'})" >' . $label . '</div></li>';
            }
            $selimp .= '</ul>';

            $selimp .= '</li>';
            $selimp .= '</ul>';

            if ($cant > 1)
                $botones2 = $selimp;

            $botones2 = '<span style="float:left;clear;none;" >' . $botones2 . $selimpButton . '</span>';
        }

        return $botones2;
    }

    protected function tooltip($colspan=20, $prefix = '')
    {
/*        $botones1 = '';
        // BOTONES PARA GRABAr
        if ($this->Datos->tooltip != 'false') {
            $botones1 = '<tr>';
            $botones1 .= $prefix;
            $botones1 .= '<td colspan="' . $colspan . '" class="tooltip" id="tooltip' . $this->Datos->idxml . '" > ';
            $botones1 .= '</td>';
            $botones1 .= '</tr>';
        }

        return $botones1;
        */
    }

    public function showBtnIng()
    {
        $form = 'Form' . $this->Datos->idxml;
        $motivo = '';

        // add toolbar buttons
        if ($this->toolbarButtons != '') {
            $botones2 .= implode('',$this->toolbarButtons);
        }
        
        if ($this->Datos->noForm != 'true') {

            $botones1 = $this->tooltip();

//            $botones2 .= $this->importDataButton();
            $refrescaOrig = '';

            if ((isset($this->Datos->xmlOrig) && $this->Datos->xmlOrig != '' ) && 
		$this->Datos->xml == $this->Datos->xmlpadre) {
                $form2 = '';
                if ($this->Datos->xmlReferente != '' && $this->Datos->xmlReferente == $this->Datos->xmlOrig)
                    $refrescaOrig = 'grabaABM(\'Form' . $this->Datos->xmlOrig . '\', \'update\',\'' . $this->Datos->xmlOrig . '\' , \'' . $this->Datos->xmlOrig . '\', false ' . $form2 . ');';
            }

            $displayCleanButton = false;

            unset($optionsArray);

            /*
            $optionsArray['instance'] = $this->Datos->getInstance() ;
            $optionsArray['xmlpadre'] = $this->Datos->xmlpadre ;
            $postOptions = htmlspecialchars(json_encode($optionsArray));            
            */

            $optionsArray['instance'] = '"'.$this->Datos->getInstance().'"' ;
            $optionsArray['parentInstance'] = '"'.$this->Datos->parentInstance.'"' ;
            $optionsArray['xmlpadre'] = '"'.$this->Datos->xmlpadre.'"' ;
            $optionsArray['clickedbutton'] = 'this' ;

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

            
            

            if ($this->Datos->insertaABM != 'no' && $this->Datos->insertaABM != 'false'
            //&& $this->Datos->autoing != 'true'
) {
                $form2 = ', null ';
                if ($this->Datos->xmlpadre != '')
                    $form2 = ',\'Form' . $this->Datos->xmlpadre . '\'';

                $btnNuevo = new Html_button($this->i18n['insert'], "../img/filenew.png", $this->i18n['insert']);
                $btnNuevo->addParameter('name', $this->i18n['insert']);
                $btnNuevo->addEvent('onclick', 'grabaABM(' . "'" . $form . "'" . ', ' . "'insert'" . ' , \'' . $this->Datos->xml . '\'  , \'' . $this->Datos->xml . '\', true' . $form2 . ' ,  ' . $postOptions . '  ); ' . $refrescaOrig);
                $btnNuevo->tabindex = $this->tabindex();
                $botones2 .= $btnNuevo->show();
                $displayCleanButton = true;
            }

            if ($this->Datos->modificaABM != 'no' && $this->Datos->modificaABM != 'false' && $this->Datos->editable != 'true') {
                $btnMod = new Html_button($this->i18n['modify'], "../img/filesave.png", $this->i18n['modify']);
                $btnMod->addParameter('name', $this->i18n['modify']);
                $btnMod->addEvent('onclick', 'grabaABM(' . "'" . $form . "'" . ', ' . "'update'" . ' , \'' . $this->Datos->xml . '\'  , \'' . $this->Datos->xml . '\', null, null, ' . $postOptions . '  ); ' . $refrescaOrig);
                $btnMod->tabindex = $this->tabindex();
                $botones2 .= $btnMod->show();
                $displayCleanButton = true;
            }
            /*
            if ($displayCleanButton) {
                $btnClear = new Html_button('', "../img/eraser.png", $this->i18n['eraser']);

                $btnClear->addParameter('name', 'eraser');
                $btnClear->addParameter('title', $this->i18n['eraser']);
                $btnClear->addEvent('onclick', 'softClearForm(this );');
                $btnClear->addStyle('float', 'left');
                $cleanbtn = $btnClear->show();
            }
            */

        }
        $salida = $botones1;

        /* test de botones de GRABACION */
        $salida .= '<tr class="filabotones">';
        $salida .= '<td colspan="20" align="center">';

        $salida .= $cleanbtn.$botones2;

        if (isset($this->Datos->CabeceraMov))
            foreach ($this->Datos->CabeceraMov as $NCabecera => $ContCab) {
                $xmlcabecera = $ContCab->xml;
            }

        if ($this->Datos->procesa != 'false') {

            // Motivos para deshabilitar el boton Confirmar
            $disabled = '';

            // SIN MINUTA CONTABLE
            if ($this->Datos->Minuta != '')
                if (!$this->Datos->Minuta->balanceado()) {
                    $disabled = 'disabled';
                    $motivo = 'Minuta contable No V&aacute;lida';
                }

            // Campos no validables
            $campos = $this->Datos->camposaMostrar();
            foreach ($campos as $nom => $nombre) {

                $objCampo = $this->Datos->getCampo($nombre);
                if ($objCampo) {

                    if (isset($objCampo->validar) && $objCampo->validar == 'true') {
                        if ($objCampo->valor == 'true' ||
                                $objCampo->valor == 1 ||
                                $objCampo->ultimo == 'true' ||
                                $objCampo->ultimo == 1) {
                            $disabled = '';
                            $motivo = '';
                        } else {
                            // EL campo no valido correctamente
                            $disabled = 'disabled';
                            $motivo = $objCampo->invalid;
                        }
                    }
                }
            }

            if ($this->tipo != 'fichaing') {
                // SIN DATOS A PROCESAR
                if ($this->Datos->TablaTemporal != '') {
                    if ($this->Datos->TablaTemporal->countRows() == 0) {
                        $disabled = 'disabled';
                        //$motivo   = 'No existen Registros';
                        $motivo = '';
                    }
                } else {
                    $disabled = 'disabled';
                    $motivo = '';
                }
            }

            if (isset($this->Datos->xmlImpresion) && $this->Datos->xmlImpresion != '')
                $impExterna = 'true';

            if (isset($this->Datos->btnconfirma) && $this->Datos->btnconfirma != '')
                $btnConf = $this->Datos->btnconfirma;
            else
                $btnConf= $this->i18n['accept'];

            $btnConf ='F9 - '.$btnConf;
            $cierraproceso = '';
            if ($this->Datos->cierraproceso != '') {
                $cierraproceso = $this->Datos->cierraproceso;
                $cierraprocesoDir = $this->Datos->cierraprocesoDir;
                $cierraprocesoCondition = $this->Datos->cierraprocesoCondition;

            }
            if ($this->Datos->grillasContenidas != '') {
                $arraygrillas = ", new Array('" . implode("','", $this->Datos->grillasContenidas) . "')";
            } else
                $arraygrillas = ", null ";

            // TODO remove all this jsparam
            //    $jsparam[4]=" this";
            //   $jsparam[5]="'".trim($this->Datos->confirmacion)."'";
            //  $jsparam[6]="'".$this->Datos->dirxml."'";
            //   $jsparam[7]=",'".$cierraproceso."'";
            //    $jsparam[8]=$arraygrillas;
            if ($this->Datos->validations != '') {
                foreach ($this->Datos->validations as $val => $messages) {
                    $validations[] = trim(addslashes($val));
                }
                $validation = " new Array('" . implode("','", $validations) . "')";
                $messages = ", new Array('" . implode("','", $this->Datos->validations) . "')";
                $validationType = ", new Array('" . implode("','", $this->Datos->validationType) . "')";

                $jsparam[9] = $validation;
                $jsparam[10] = $messages;
                $jsparam[11] = $validationType;
            } else {
                $jsparam[9] = ' null';
                $jsparam[10] = ', null';
                $jsparam[11] = ', null';
            }
            $jsparam[12] = ', null';
            if ($this->Datos->__inlineid != '') {
                $jsparam[12] = " , '" . $this->Datos->__inlineid . "'"; // div_destino (used in inline
            }

            $options = "{ xmldatos:'{$this->Datos->xml}' , xmlcabecera:'$xmlcabecera', button:this ";
            $options .= ", mensajeconf:'" . trim($this->Datos->confirmacion) . "'";
            $options .= ", dir:'{$this->Datos->dirxml}'";
            $options .= ", cierraproceso:'$cierraproceso'";
            $options .= ", cierraprocesoDir:'$cierraprocesoDir'";
            $options .= ", cierraprocesoCondition:'$cierraprocesoCondition'";

            if ($this->Datos->grillasContenidas != '') {
//                $options .= ", arrayGrillas: new Array('" . implode("','", $this->Datos->grillasContenidas) . "')";
        $options .= ", instanceGrids: new Array(";
            $quote = ' ';

            foreach ($this->Datos->grillasContenidas as $xmlGrid => $instanceGrid) {
                    $options .= $quote."'".$instanceGrid."'";
                    $quote = ' , ' ;
                }
        $options .=  ")";

        $options .= ", arrayGrillas: new Array(";
            $quote = ' ';

            foreach ($this->Datos->grillasContenidas as $xmlGrid => $instanceGrid) {
                    $options .= $quote."'".$xmlGrid."'";
                    $quote = ' , ' ;
                }
        $options .=  ")";

            }

            if ($this->Datos->validations != '') {
                foreach ($this->Datos->validations as $val => $messages) {
                    $validations2[] = trim(addslashes($val));
                }
                $options .= ", validations: new Array('" . implode("','", $validations2) . "')";
                $options .= ", messages: new Array('" . implode("','", $this->Datos->validations) . "')";
                $options .= ", validationOptional: new Array('" . implode("','", $this->Datos->validationType) . "')";
            }
            $options .= ", xmlOrig:'{$this->Datos->xmlOrig}'";

            if ($this->Datos->__inlineid != '') {
                // div_destino (used in inline
                $options .= ", div_destino:'{$this->Datos->__inlineid}'";
            }

            $options .= ", instance:'{$this->Datos->getInstance()}'";

            $options .= '}';

            $jsprocesar = 'procesar(' . $options . ');';

            $btnProcesar = new Html_button($btnConf, "../img/run.png", $btnConf);
            if ($disabled != '')
                $btnProcesar->addParameter('disabled', 'disabled');
            $btnProcesar->addParameter('name', 'Procesar');
            $btnProcesar->addParameter('title', $motivo);
            $btnProcesar->addEvent('onclick', $jsprocesar);
            $btnProcesar->tabindex = $this->tabindex();
            $salida .= $btnProcesar->show();
            if ($motivo != '') {
                $salida .= '<div class="motivo">' . $motivo . '</div>';
            }
        }

        $salida .= $this->xmlEditor();
        $salida .= '</td>';
        $salida .= '</tr>';

        return $salida;
    }

}
