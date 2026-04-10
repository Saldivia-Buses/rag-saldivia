<?php
/*
 * 2009-09-09
 * basic table for information display
*/

class UI_consulta extends UI
{
    /**
     * User Interfase constructor
     *
     */
    public $nosel;
    public $contFiltro;
    public $esAyuda;
    public $autocomplete;
    public function __construct(&$dataContainer)
    {
        parent::__construct($dataContainer);

        $this->muestraCant = true;
        $this->disabledCheckDefault = true;
        $this->disabledCellId = true;

        $this->defaultClass = 'consultaing2';
        $this->uidAutoFiltros = UID::getUID('AF',true);
        $dataContainer->unserializeParent = 'false';

    }

    public function setFiltro($cont_filtro, $campoFiltro)
    {
        $this->contFiltro  = $cont_filtro;
        $this->campoFiltro = $campoFiltro;

        /* ejecuto el query */
        // TODO remove from HERE
        return $this->contFiltro->Select();
    }

    /**
     *  Render de complete XML
     * @param  string $idFormulario
     * @param  string $divcont
     * @param  array  $opt
     * @return string
     */
    public function show($idFormulario = '', $divcont='', $opt='')
    {
        $salida = '';

        $barraSlide = '';
        /* si es un detalle de otra consulta no pongo el div principal */

        $id = 'Show'.$this->Datos->idxml;

        // id del contenedor (creo)
        $id2= str_replace('.', '_',($divcont != '')?$divcont:$id );

        if (isset($this->Datos->detalle) && $this->Datos->inline != 'true')
            $retrac = true;

        $style = (isset($this->Datos->style))?$this->Datos->style:'';

        // Columns
        $ancho = (isset($this->Datos->ancho))?$this->Datos->ancho : '';
        $width = (isset($this->Datos->width))?$this->Datos->width : $ancho;

        if ($width != '') {
            $this->Datos->col1=$width;
            $this->Datos->col2=100 - $width;
            $style.='width:'.$this->Datos->col1.'%;';
        }

        $clase	= $this->defaultClass;

        if ((isset($this->Datos->detalle) && $this->Datos->inline != 'true' ) ||
                (isset($this->Datos->grafico) && $this->Datos->grafico != '')) {
            $clase	= 'consulta';
            $barraSlide = $this->showSlider($id, $retrac);
        }

        // Si se define explicitamente una clase en el xml
        if (isset($this->Datos->clase) && $this->Datos->clase != '') {
            $clase = $this->Datos->clase;
        }

        // display Table
        $salidaDatos = $this->showTabla();

        if (isset($this->Datos->__inline) && $this->Datos->__inline == 'true') {
            $salida .= $salidaDatos;
        } else {
            $retorno = '';
            if (isset($this->Datos->campoRetorno)) {
                $uidRetorno = $this->Datos->getCampo($this->Datos->campoRetorno)->uid;
                $retorno = ' origen="'.$uidRetorno.'" ';
            }

            $salida .=  '<div  class="'.$clase.'" id="'.$id.'" style="'.$style.'" '.$retorno.'>';
            $salida .= '<div class="contewin" >';
            $salida .= $salidaDatos;
            $salida .= '</div>';
            $salida .= '</div>';
        }

        // Incorporo la barra vertical para slide
        $salida .= $barraSlide;

        // Add Detail div
        $salida .= $this->detailDiv('detalle');

        // Graficos
        if (isset($this->Datos->grafico)) {
            $salida .= $this->showGraficos();
        }
        // add events
        $salida .= $this->eventScripts();

        return $salida;

    }

    /**
     * Javascript events to run after display content
     */
    protected function eventScripts($script='')
    {
        // create Javascript functions

        $script[]= "Histrix.registroEventos('".$this->Datos->idxml."');";
        $resize = (isset($this->Datos->resize))?$this->Datos->resize:'';
        $resizeTable = (isset($this->resizeTable))?$this->resizeTable:'';

//        if ($resize != "false" && $resizeTable != "false")
  //          $script[]= "Histrix.calculoAlturas('".$this->Datos->idxml."', null ); ";

//	$script[]= $this->Datos->customScript;
        $script[]= $this->Datos->getCustomScript();

        $output = Html::scriptTag($script);

        return $output;
    }

    public function detailDiv($detailClass)
    {
        $salida = '';
        if (isset($this->Datos->detalle)  && $this->Datos->inline != 'true') {
            // solo muestro la cabecera de lo consultado si se pide explicitamente
            if ($this->Datos->showCab == 'true')
                $salidaDetalle .= $this->showAbm('readonly');

            if ($this->Datos->col1 != '') $style_2='left:'.($this->Datos->col1 + 0.5).'%';
            $salidaDetalle .= '<div  class="'.$detailClass.'" target="'.$this->Datos->detalle.'" id="Det'.$this->Datos->idxml.$this->Datos->iddetalle.'"   style="'.$style_2.'">';
            $salidaDetalle .= '<div class="esperareloj" >'.$this->i18n['selectRowToDetail'].'</div>';
            $salidaDetalle .= '</div>';

            $salida .= $salidaDetalle;
        }

        return $salida;
    }

    public function autoFiltros()
    {
        $salida = '<div class="filtro" id="'.$this->uidAutoFiltros.'" style="display:none;">';

        $listaCampos = $this->Datos->camposaMostrar();

        foreach ($listaCampos as $nNombre => $nombrelista) {
            $objCampo = $this->Datos->getCampo($nombrelista);
            $nombre =  htmlentities(ucfirst($objCampo->Etiqueta), ENT_QUOTES, 'UTF-8');
            if ($nombre != '')
                $options[$nombrelista] = $nombre;
        }
        if (count($options) > 0) {
            $inputSelect = new Html_select($options);
            $inputSelect->addParameter('id', '_Autofiltro'.$this->Datos->idxml);
            $inputSelect->addParameter('name', '_Autofiltro');
            $salida .= $inputSelect->show();
        }

        $operators = array('=' => $this->i18n['='].' (=)' ,
                '&gt;' =>$this->i18n['>'].' (&gt;)',
                '&gt;=' =>$this->i18n['>='].' (&gt;=)',
                '&lt;' =>$this->i18n['<'].' (&lt;)',
                '&lt;=' =>$this->i18n['<='].' (&lt;=)',
                'like' =>$this->i18n['like'],
                '!=' =>$this->i18n['!=']);
        $inputSelect = new Html_select($operators);
        $inputSelect->addParameter('id', '_AutoOperador'.$this->Datos->idxml);
        $inputSelect->addParameter('name', '_AutoOperador');
        $salida .= $inputSelect->show();

        $btn = new Html_button($this->i18n['add'], "../img/add.png" ,"Autofiltro" );
        $btn->addEvent('onclick', 'addfiltro(\''.$this->Datos->xml.'\' , \''.$this->Datos->xmlOrig.'\' )');
        $btn->tabindex = $this->tabindex();
        $salida .= $btn->show();

        $salida .= "</div>";

        return $salida;
    }

    public function showFiltro()
    {
        $campos = _num_fields($this->contFiltro->resultSet);
        $i = 0;
        for ($j = 1; $j <= $campos; $j ++) {
            if (_field_name($this->contFiltro->resultSet, $j) != 'ROWID') {
                $nom[$j] = _field_name($this->contFiltro->resultSet, $j);
                $i ++;
            }
        }

        $filterField = $this->Datos->getCampo($this->Datos->campofiltroPrincipal);
        while ($row = _fetch_array($this->contFiltro->resultSet)) {
            $key    = $row[$nom[1]];
            $value  = $row[$nom[2]];
            $filterField->opcion[$key]=$value;
        }
        $inputSelect = new Html_select($filterField->opcion);
        $inputSelect->tabindex = $this->tabindex();
        $inputSelect->addParameter('id', 'filtro'.$filterField->uid);
        $inputSelect->addParameter('name', 'filtro');
        $inputSelect->addEvent('onchange', 'filtrar('."'".$this->campoFiltro."'".
                ', this.value, '."'='".', \''.$this->Datos->xml.'\' , \''.$this->Datos->xml.'\', this , \''.$this->Datos->xmlOrig.'\');');
        $salida .= $inputSelect->show();

        return $salida;
    }

    /*
     * Muestra los filtros de las consultas
     * (Rehacer para no pasar dos veces para implementar los grupos)
    */

    public function showFiltrosXML($display=true, $buttons=true, $extra_data = null)
    {
        if (isset($this->Datos->resultSet))
            $campos = $this->cantCampos();

        $idFiltros   = 'Filtros'.$this->Datos->idxml;
        $formFiltros = 'FForm'.$this->Datos->idxml;

        //	if ($nodiv !='nodiv')
        $salida ='';
        $style= '';
        if ($display == false)
            $salida .= '<fieldset style="display:none;">';
        else
            $salida .= '<fieldset>';
    $class = ($this->Datos->autoFilter == 'true')?' class="autofilter" ':'';

        $salida .= '<form '.$class.'  name="'.$formFiltros.'" id="'.$formFiltros.'" action="" onsubmit="return false;" tipo="filter" instance="'.$this->Datos->getInstance().'">';

        // Add top Extra data
        if (isset($extra_data['top'])){
            
            $salida .= implode(' ', $extra_data['top']);

        }

        $salida .= '<table >';
        $hayFiltros = false;
        $enableFilters = false;

        if ($this->contFiltro) {
            $hayFiltros = true;
            $enableFilters = true;
            $salida .= "<tr>";
            $salida .= '<td><b>';
            $salida .= $this->contFiltro->getTitulo();
            $salida .= '</b></td>';
            $salida .= '<td colspan="2">';
            $salida .= $this->showFiltro();
            $salida .= '</td>';
            $salida .= "</tr>";
        }

        $i = 0;
        $pri=false;
        if (isset ($this->Datos->filtros) && $this->Datos->filtros) {
            $hayFiltros = true;
            // 1ro Agrupo los filtros por grupos
            $atributos	='';
            $gruposFiltros= [];

            foreach ($this->Datos->filtros as $nomfiltro => $objFiltro) {
                // hide filters if there is not any enable
                if ($objFiltro->deshabilitado != 'true') {
                    $enableFilters = true;
                }
                $grupo = $objFiltro->grupo ?$objFiltro->grupo:'default';
                $gruposFiltros[$grupo][]= $objFiltro;

            }

            foreach ($gruposFiltros as $grupoFiltro => $objfiltros) {

                // Solo muestro los grupos Explicitos en el xml, no los automaticos
                if ($grupoFiltro != 'default') {

                    $salida  .= '<tr><td colspan="10">';
                    $uidfiltro = UID::getUID($grupoFiltro, true);
                    $salida .= '<fieldset id="'.$grupoFiltro.'" >';
                    $salida .=  '<legend><span class="activafiltro"><input type="checkbox" onchange="toggleID(\''.$uidfiltro.'\'); onoffAtt(\'deshab\', \'grupo\' , \''.$grupoFiltro.'\', \''.$formFiltros.'\' );"> - '.$grupoFiltro.'</span></legend>';
                    $salida .= '<table style="display:none" id="'.$uidfiltro.'">';
                    $deshabilitar=true;
                    $atributos['deshab'] ='true';
                    $atributos['grupo']  = $grupoFiltro;
                }

                foreach ($objfiltros as $clv => $objFiltro) {
                    $remfiltro = UID::getUID('rf', true);
                    $i ++;
                    $campoFiltro = $this->Datos->getCampo($objFiltro->campo);

                    if (!is_object($campoFiltro)) {
                        echo '<div class="error"> error al filtrar '. $objFiltro->campo.'</div>';
                        continue;
                    }

                    if (isset($campoFiltro->deshabilitado))
                        $destemp = $campoFiltro->deshabilitado;

                    $campoFiltro->deshabilitado = $objFiltro->deshabilitado;

                    if ($campoFiltro == null)
                        continue;
                    $operadores = $campoFiltro->getOperadores();
                    $atributos['operador']  = htmlentities($objFiltro->operador);

                    if ($objFiltro->opcion == 'auto') {
                        $filterClass = 'class="autoFilter"';

                    }

                    // por ahora
                    if (!isset($objFiltro->modpos))
                        $objFiltro->modpos = $campoFiltro->modpos;
                    if ($objFiltro->modpos != 'nobr') {
                        if ($pri)
                            $salida .= '</tr>';
                        $salida .= '<tr id="'.$remfiltro.'" '.$filterClass.'>';
                        $pri = true;
                    }

		    $filterStyle= ' style="'.$campoFiltro->filterStyle.'" ';
                    $salida .= '<td '.$filterStyle.'>';
                    if ($objFiltro->opcion == 'auto') {
                        $salida .= '<img alt="remover" src="../img/remove2.png" onClick="delautofiltro(\''.$this->Datos->xml.'\',\''.$remfiltro.'\', \''.$objFiltro->uid.'\' , \''.$this->Datos->xmlOrig.'\');" title="Remover Campo de busqueda">';
                        $campoFiltro->deshabilitado="false";
                    }

                    $salida .= '<span><b>'.htmlentities($objFiltro->label,null, 'UTF-8').'</b></span>';
                    $salida .= "</td>";

                    $colspan = ($objFiltro->colspan != '')?' colspan="'.$objFiltro->colspan.'" ':'';
                    $prefijo = '';
                    $idOPE = $objFiltro->campo;
                    if ($objFiltro->opcion == 'xml' || $objFiltro->opcion == 'auto') {

                    } else {
                        $salida .= "<td $colspan>";
                        $inputSelect = new Html_select($operadores);
                        $inputSelect->addParameter('id', 'Cond'.$idOPE);
                        $inputSelect->addParameter('name', 'operador');
                        $inputSelect->value 	= $objFiltro->operador;
                        $salida .= $inputSelect->show();
                        $salida .= "</td>";
                    }

                    $salida .= "<td $colspan>";
//                    $opciones = $campoFiltro->opcion;

                    // copias en js de un objeto a otro, no esta bien resuelto, REHACER
                    $campoFiltro->copia = $objFiltro->copia;

                    $salida .= $campoFiltro->renderInput($this, $formFiltros, $idOPE, $objFiltro->valor,  'filtro', '', $atributos);

                    if (isset($destemp))
                        $campoFiltro->deshabilitado = $destemp ;
                    unset($destemp);

                    $salida .= "</td>";

                }
                $salida .= "</tr>";

                if ($grupoFiltro != 'default') {
                    $salida  .= '</table>';
                    $salida .= '</fieldset>';
                }
            }

            // $salida .= "</tr>";
        }

        if ($i > 0 && $buttons) {
            $hayFiltros = true;
            $salida .= '<tr><td colspan="20" align="center">';

            $btnBuscar = new Html_button($this->i18n['search'], "../img/find.png" ,$this->i18n['search']);
            $btnBuscar->addParameter('class', 'searchButton');
            $btnBuscar->addEvent('onclick', 'filtracampos('."'".$formFiltros."'".' , \''.$this->Datos->xml.'\' , \''.$this->Datos->xml.'\' , \''.$this->Datos->xmlOrig.'\', \''.$this->Datos->getInstance().'\' )');
            $btnBuscar->tabindex = $this->tabindex();
            $salida .= $btnBuscar->show();

            $salida .= '</td></tr>';

        }
        $salida .= "</table>";
        // Add top Extra data
        if (isset($extra_data['bottom'])){
            $salida .= implode(' ', $extra_data['bottom']);
        }
        $salida .= "</form>";
        $salida .= '</fieldset>';
        $salida .= '</div>';

        if ($hayFiltros) {
            if ($enableFilters == false) $style = ' style="display:none;" ';
            $salida = '<div id="'.$idFiltros.'" '.$style.' class="filtro">'.$salida;

            return $salida;
        }

    }

    public function showTabla($opt = '')
    {
        $idTabla = $this->Datos->idxml;

        if (isset($this->Datos->imprime) && $this->Datos->imprime == 'false') $bottom = 0;
        else $bottom = 31;

        if ( !isset($this->Datos->esInterno) || $this->Datos->esInterno != true)
            $estiloPriv= 'position:absolute;top:0px;bottom:'.$bottom.'px; left:0px;right:0px; overflow:auto;';

        $autofilter = (isset($this->Datos->autofiltro))? $this->Datos->autofiltro : 'true';

        if ($this->contFiltro || ( isset($this->Datos->filtros)  && $this->Datos->filtros)  || $autofilter != 'false') {

            $filtros = $this->showFiltrosXML();

            if ($autofilter != 'false')
                $filtros .= $this->autoFiltros();

        }

        $salidaTabla = $this->showTablaInt($opt, $idTabla);
        if (isset($this->Datos->__inline) && $this->Datos->__inline == 'true') return $salidaTabla;

        $tablaInt = Html::tag('div', $salidaTabla,
                array('id' => $idTabla, 'class' => 'contTablaInt', 'instance' => $this->Datos->getInstance()));
        $propDiv  = array('id'=>'IMP'.$idTabla, 'class'=>'TablaDatos',
                'cellpadding'=>0, 'cellspacing'=>0,'style'=>$estiloPriv );
        $salida = Html::tag('div', $filtros.$tablaInt, $propDiv );

        $isInner = (isset($this->Datos->esInterno)) ? $this->Datos->esInterno:'';

        // add toolbar buttons
        if ($this->toolbarButtons != '') {
            $toolbarButons = implode('',$this->toolbarButtons);        
        }

        if ( $isInner != true)
            $salida .= $this->botonera($toolbarButons);

        return $salida;
    }

    public function showUL()
    {
        $form = 'Form'.$this->Datos->idxml;

        if ($this->Datos->xmlpadre != '')
            $form = 'Form'.$this->Datos->xmlpadre;

        $originalFieldName= $parentObject->NombreCampo;
        $tablaTemp = $this->Datos->TablaTemporal->datos();
        $searchString = $this->Datos->lastSearchString;
        $replaceStr = '';

        // Generate Column labels
        $i = 0;
        if ($tablaTemp)
            foreach ($tablaTemp as $order => $row) {
                $i++;
                $pos=9999;
                $destino='';
                foreach ($row as $fieldName => $value) {

                    $field = $this->Datos->getCampo($fieldName);
                    if (($field->Oculto))continue;

                    if ($field->Detalle != '') {
                        foreach ($field->Detalle as $index => $name) {

                            $destino[] = 'var node = {name: "'.$name.'" , value: "'.addslashes($value).'" }';
                        }
                    }

                    if ($field->Parametro['noshow'] != 'true') {
                        $visible[]=$value;
                    }

                }
                if ($visible != '') $line = implode(',',$visible);
                if ($destino != '') $line .= '|'.implode('|', $destino);
                unset($hidden);
                unset($visible);
                $posicion[$i] = $pos;
                $li[$i] = $line;
            }

        if ($li != '') {
            asort($posicion);
            foreach ($posicion as $linepos => $order) {
                $salida .= $li[$linepos]."\n";
            }
        }

        return $salida;
    }

    public function btnNotificar()
    {
        $uid= UID::getUID('notif', true);
        $contenido = htmlentities('Se Notifican cambios en el modulo: '.$this->Datos->titulo);
        $contetiq = 'histrixLoader.php?xml=mensajes_notificar.xml&dir=histrix/mensajeria&contenido='.urlencode($contenido);
        $opcetiq = '\'\'';
        if ($this->Datos->xmlOrig != '')
            $padre = 'DIV'.$this->Datos->xmlOrig;
        else $padre = 'DIV'.$this->Datos->xml;

        $btn = new Html_button("Notificar", "../img/mail-forward.png" ,"Notificar" );
        $btn->addEvent('onclick', 'Histrix.loadInnerXML(\'mensajes_notificar.xml\', \''.$contetiq.'\', '.$opcetiq.',\'Notificar\', \''.$padre.'\',\''.$uid.'\' )');
        $btn->addParameter('id',$uid);
        $btn->tabindex = $this->tabindex();

        $salida = $btn->show();

        return $salida;
    }

    /*
     * Export Buttons
    */
    public function btnExportar()
    {
        $idDiv= UID::getUID();
        //$btn = new Html_button($this->i18n['export'], "../img/exportar.png" ,$this->i18n['export'] );
	$export = '<span class="btnExportText">'.$this->i18n['export'].'</span>';
        $btn = new Html_button($export, "../img/exportar.png" ,'');
        $btn->addParameter('title',$this->i18n['exportTitle']);
        $salida = Html::tag('div', $btn->show(), array('onclick'=>'$(\'#'.$idDiv.'\').slideToggle()'));

        // EXPORT FORMATS
        $formats['xls']= array("../img/spreadsheet.png" ,"Excel");
        $formats['ods']= array("../img/ximian-openoffice-calc.png" ,"Ods");
        $formats['xml']= array("../img/xml.png" ,"XML");
        $formats['csv']= array("../img/text-x-generic.png" ,"CSV");

        if ($this->Datos->getFieldByAttribute(array('wp_title'=>'true')))
            $formats['wp']= array("../img/text-x-generic.png" ,"wordpress");

        //$formats['bas']= array("../img/text-x-generic.png" ,"BAS");

        if (isset($this->Datos->ldifExport) && $this->Datos->ldifExport)
            $formats['ldif']= array("../img/text-x-generic.png" ,"LDIF");
        $botones = '';
        foreach ($formats as $format => $details) {
            $btn = new Html_button($details[1], $details[0] ,$details[1]);
            $btn->addEvent('onclick', 'Histrix.exportFile(\''.$this->Datos->getInstance().'\', \''.$this->Datos->getTitulo().'\', \''.$format.'\', \''.$this->Datos->xmlOrig.'\', \''.$_SESSION['DAT'].'\' );');
            $btn->addEvent('onclick', 'Histrix.toggle(\''.$idDiv.'\');', true);
            $btn->addParameter('title', $this->i18n['export'.strtoupper($format).'Title']);
            $btn->addStyle('width', '99%');
            $botones .= $btn->show();
        }
        $salida .= Html::tag('div', $botones, array('id'=>$idDiv, 'class'=>'botonesExportar optionPanel', 'style'=>'display:none;'));

        return $salida;
    }

    /*
     * Option Buttons
    */
    public function btnOptions()
    {
        $idDiv= UID::getUID();
        $btn = new Html_button('', "../img/configure.png" , '' );
        $btn->addParameter('title',$this->i18n['optionsTitle']);
        $salida = Html::tag('div', $btn->show(), array('onclick'=>'$(\'#'.$idDiv.'\').slideToggle()'));

        if (isset($this->uidAutoFiltros)) {

            $btn = new Html_button($this->i18n['filters'], '../img/filtros.png' ,$this->i18n['filters']);
            $btn->addEvent('onclick', 'toggleID(\''.$this->uidAutoFiltros.'\');');
            $btn->addParameter('title', $this->i18n['filterTitle']);
            $btn->addStyle('width', '100%');
            $btn->addStyle('text-align', 'left');
            $botones = $btn->show();
        }

        $btn = new Html_button($this->i18n['chart'], '../img/x-office-spreadsheet.png' ,$this->i18n['chart']);
        $btn->addEvent('onclick', 'addGraficos(\''.$this->Datos->xml.'\');');
        $btn->addEvent('onclick', '$(\'#'.$idDiv.'\').slideToggle();', true);
        $btn->addParameter('title', $this->i18n['chart']);
        $btn->addStyle('width', '100%');
        $btn->addStyle('text-align', 'left');
        $botones .= $btn->show();

        if (isset($this->isTree) && $this->isTree == true) {
            $btn = new Html_button($this->i18n['treeview'], '../img/view_tree1.png' ,$this->i18n['treeview']);
            $btn->addEvent('onclick', 'addInnerGraphWindow(\''.$this->Datos->xml.'\' , \''.$this->Datos->titulo.'\' , \'ventint\');');
            $btn->addEvent('onclick', '$(\'#'.$idDiv.'\').slideToggle();', true);
            $btn->addParameter('title', $this->i18n['treeview']);
            $btn->addStyle('width', '100%');
            $btn->addStyle('text-align', 'left');
            $botones .= $btn->show();
         }

        if ($_SESSION['EDITOR'] == 'editor') {
            $file = $this->Datos->xml;

            if ($this != '') {
                $dir = (isset($this->Datos->dirxml))?$this->Datos->dirxml:'' ;
                $titulo = 'EDITOR '.$file;
                $titulo2 = 'Last SQL '.$file;

                $btn = new Html_button($this->i18n['edit'], '../img/edit.png' ,$this->i18n['edit']);
                $btn->addEvent('onclick', "xmlLoader('Edit_$file', '&url=codeEditor.php&file=$file&dir=$dir', {title:'$titulo', loader:'cargahttp.php'});");
                $btn->addEvent('onclick', '$(\'#'.$idDiv.'\').slideToggle();', true);
                $btn->addParameter('title', $this->i18n['edit'].' '.$dir.'/'.$file);
                $btn->addStyle('width', '100%');
                $btn->addStyle('text-align', 'left');

                $botones .= $btn->show();

                $btn = new Html_button($this->i18n['sql'], '../img/sql_ico.jpg' ,$this->i18n['sql']);
                $btn->addEvent('onclick', "Histrix.loadInnerXML('Debug_$file', 'debug.php?xml=$file&instance=".$this->Datos->getInstance()."', '$titulo2', '$titulo2',null,null, 'cargahttp.php');");
                $btn->addEvent('onclick', '$(\'#'.$idDiv.'\').slideToggle();', true);
                $btn->addParameter('title', $this->i18n['sql'].' '.$dir.'/'.$file);
                $btn->addStyle('width', '100%');
                $btn->addStyle('text-align', 'left');

                $botones .= $btn->show();

            }
        }

        if (isset($this->Datos->log) && $this->Datos->log  != '') {
            $paramstring= '&dir=histrix/menu&xmlprog='.$this->Datos->xml;
            $btn = new Html_button($this->i18n['history'], '../img/appointment-new.png' ,$this->i18n['history']);
            $btn->addEvent('onclick', 'Histrix.ventInt(\''.$this->Datos->idxml.'\',\'log_cons.xml\', \''.$paramstring.'\', \'Historial\', { parentInstance: \''.$this->Datos->getInstance().'\'});');
            $btn->addEvent('onclick', '$(\'#'.$idDiv.'\').slideToggle();', true);
            $btn->addParameter('title', $this->i18n['history']);
            $btn->addStyle('width', '100%');
            $btn->addStyle('text-align', 'left');
            $botones .= $btn->show();
        }

        if ($_SESSION['administrator'] == 1) {
            $btn = new Html_button($this->i18n['notifications'], '../img/mail_generic.png' ,$this->i18n['notifications']);
            $_menuId = (isset($this->Datos->_menuId))?$this->Datos->_menuId:'';
            $btn->addEvent('onclick', "Histrix.loadInnerXML('htxnotifuser.xml' , 'histrixLoader.php?xml=htxnotifuser.xml&amp;dir=histrix/menu&amp;menuId=".$_menuId."', '','Notificaciones', 'DIVhtxnotifuser.xml','Notification' , {width:'580px', height:'260px'});");
            $btn->addEvent('onclick', '$(\'#'.$idDiv.'\').slideToggle();', true);
            $btn->addParameter('title', $this->i18n['notifications']);
            $btn->addStyle('width', '100%');
            $btn->addStyle('text-align', 'left');

            $botones .= $btn->show();

            $btn = new Html_button($this->i18n['permissions'], '../img/emblem-readonly.png' ,$this->i18n['permissions']);
            $btn->addEvent('onclick', "Histrix.loadInnerXML('profileAuthItem_xml', 'histrixLoader.php?xml=profileAuthItem.xml&amp;dir=histrix/menu&amp;menuId=".$_menuId."', '','Permisos', 'DIVprofileAuthItem.xml','Permisos' , {width:'580px', height:'260px'});");
            $btn->addEvent('onclick', '$(\'#'.$idDiv.'\').slideToggle();', true);
            $btn->addParameter('title', $this->i18n['permissions']);
            $btn->addStyle('width', '100%');
            $btn->addStyle('text-align', 'left');
            $botones .= $btn->show();

            $btn = new Html_button($this->i18n['users'], '../img/system-users.png' ,$this->i18n['users']);
            $btn->addEvent('onclick', "Histrix.loadInnerXML('userAuthItem_xml', 'histrixLoader.php?xml=userAuthItem.xml&amp;dir=histrix/menu&amp;menu_id=".$_menuId."', '','Permisos', 'DIVuserAuthItem.xml','Permisos' , {width:'580px', height:'260px'});");
            $btn->addEvent('onclick', '$(\'#'.$idDiv.'\').slideToggle();', true);
            $btn->addParameter('title', $this->i18n['users']);
            $btn->addStyle('width', '100%');
            $btn->addStyle('text-align', 'left');
            $botones .= $btn->show();

        }

        if ($this->Datos->helpLink != 'false'){

          $helpId  = $this->Datos->getHelpLink();

        //$helpId  = 'cargahttp.php?url='.str_replace('?', '&', $helpId);
          $btn = new Html_button($this->i18n['manual'], '../img/Manual.png' ,$this->i18n['manual']);

      //  $btn->addEvent('onclick', "Histrix.loadInnerXML('userAuthItem_xml', '$helpId', '','', '','Permisos' , {modal:true});");
          $btn->addEvent('onclick', '$(\'#'.$idDiv.'\').slideToggle();', true);
          $btn->addParameter('title', $this->i18n['manual']);
          $btn->addParameter('class', 'internalLink');
          $btn->addParameter('href', $helpId);

          $btn->addStyle('width', '100%');
          $btn->addStyle('text-align', 'left');
          $botones .= $btn->show();
        }
        /*
            if (in_array('idgraf', $opt))
                $graf  = '<img style="cursor:pointer;" src="../img/x-office-spreadsheet.png"  alt="gr&aacute;ficos" title="Modificar Gr&aacute;fico"  onClick="addGraficos(\''.$this->Datos->xml.'\',  \''.$idgraf.'\');" />';

                          if (in_array('rss', $opt)) {

                $rsslink = 'exportrss.php?db=saldivia&user='.$_SESSION['usuario'].'&dir='.$this->Datos->subdir.'&xmldatos='.$this->Datos->xml.'&p='.md5($_SESSION['pass']);
                $rss = '<a style="border:0px;" target="blank" href="'.$rsslink.'" title="Puede suscribirse a este enlace mediante un lector de noticias" >'.'<img  style="border:0px;" src="../img/rss.png" />'.'</a>';
            }
        */
        $salida .= Html::tag('div', $botones, array('id'=>$idDiv, 'class'=>'botonesExportar optionPanel', 'style'=>'display:none;'));

        return $salida;
    }

    public function pagination($opt)
    {
        // Pagination
        $salida = '';
        if (isset($this->Datos->paginar) && $this->Datos->paginar != '' && $opt != 'noecho') {
            $maxlinks = 10;
            $paginaActual = (isset($this->Datos->paginaActual))?$this->Datos->paginaActual:0;
            $paginar      = $this->Datos->paginar;
            $pages      = ($this->Datos->TotalRegistros / $paginar );
            $totalPaginas = $pages + 1;
            if ($totalPaginas - 1  > 1) {

                $textoprefijo = ($this->Datos->prefijoResultados != '')?$this->Datos->prefijoResultados:$this->i18n['records'].':';
                $textosufijo  = ($this->Datos->sufijoResultados  != '')?$this->Datos->sufijoResultados:'';

                $paginacion .= '<div id="_paginar_'.$this->Datos->idxml.'" class="paginar" ><span style="float:left; clear:none;">'.$textoprefijo.$this->Datos->TotalRegistros.$textosufijo.'</span>';
                if ($paginaActual > 0)
                    $paginacion .= '<a onclick="paginar(\''.$this->Datos->xml.'\','.($paginaActual - 1 ).', \''.$this->Datos->xmlOrig.'\' , \''.$this->Datos->getInstance().'\');" > « </a>';
//                    $paginacion .= '<a onclick="paginar(\''.$this->Datos->xml.'\','.($paginaActual - 1 , \''.$this->Datos->getInstance().'\').' , \''.$this->Datos->xmlOrig.'\');" > « </a>';

                $init = 1;

                if ($paginaActual > $maxlinks)
                    $init = $paginaActual -  $maxlinks;

                for ($i= $init ; $i<=$totalPaginas; $i++) {
                    if ( $i > $paginaActual + $maxlinks - $init + 1  ) continue;
                    if (($paginaActual + 1 ) !== $i) {
                        $paginacion .= '<a onclick="paginar(\''.$this->Datos->xml.'\','.($i - 1 ).', \''.$this->Datos->xmlOrig.'\' , \''.$this->Datos->getInstance().'\');" >'.$i.' </a> ';
                    } else {
                        $paginacion .='<span class="res">'.$i.'</span> ';
                    }
                }

                if (($paginaActual + 1 ) < intval($totalPaginas))
                    $paginacion .= '<a onclick="paginar(\''.$this->Datos->xml.'\','.($paginaActual + 1 ).', \''.$this->Datos->xmlOrig.'\' , \''.$this->Datos->getInstance().'\');" > » </a>';
                $paginacion .= '</div>';

                $salida .= $paginacion;
            }
        }

        return $salida;
    }

    public function showTablaInt($opt = '', $idTabla = '', $segundaVez = '', $nocant='', $div=false, $form=null, $pdf=null, &$parentObject=null)
    {
        $defaultForm = 'Form'.$this->Datos->idxml;
        $formini = '';
        $formfin = '';
        // nombre del form
        if ($form == null) {
            $form = $defaultForm;
        }

        $form = str_replace('.', '_', $form);
        $isInner = (isset($this->Datos->isInner))?$this->Datos->isInner:'';
        $xmlorig = '';
        // Si es un subForm interno estos valores NO coinciden y no escribo el tag form
        if ($form == $defaultForm && $isInner != 'true') {
            if (isset($this->Datos->xmlOrig) && $this->Datos->xmlOrig != '')
                $xmlorig = ' original="'.$this->Datos->xmlOrig.'"';
            $formini = '<form id="'.$form.'" name="'.$form.'" onsubmit="return false;" action="" tipo="'.$this->tipo.'" '.$xmlorig.' instance="'.$this->Datos->getInstance().'">';
            $formfin = '</form>';
        }

        $salida = '';
        $xmlcabecera = '';
        $abming = false;
        if ( $this->Datos->tipoAbm == 'ing' || $this->Datos->tipoAbm == 'grid')
            $abming = true;


        $llenoTemporal = (isset($this->Datos->llenoTemporal))?$this->Datos->llenoTemporal:'';
        $preload        = (isset($this->Datos->preloadData))?$this->Datos->preloadData:'';


        $this->TIEMPO_CONSULTA= processing_time();
        if ($llenoTemporal != "false" && $segundaVez == '' && $opt !='noselect') {
            if ($this->nosel == 'true') {
                // 'no hago select';

            } else {

                if (!isset($this->Datos->tempTables[$llenoTemporal]) && ($llenoTemporal == '' || $llenoTemporal == 'true')) {

                    if ($preload != "false") {
                        $this->Datos->Select();
                    }
                    $this->Datos->preloadData = "true";

                    if ($this->Datos->resultSet)
                        $this->cantCampos = _num_fields($this->Datos->resultSet);
                    else $campos = $this->cantCampos();

                }
                // Cargo tabla temporal con el resultado del select ODBC
                // Tarda un poco mas, SI, pero despues lo trato mas facil en la temporal :D
                // Y puedo Paginar sin tener en cuenta restricciones en el motor SQL
                // YA SE que es mas lento, pero bueno, velocidad x interoperabilidad
                // Que se le va a hacer...
                //                loger('cargoTemporal');

                $this->Datos->CargoTablaTemporal();
//die();
            }
        }
        /* Show an Horizontal Grid
          * instead of a Vertical One
        */
        if ($this->tipo == 'horizontalGrid') {
            return $this->showHorizontalGrid($opt, $parentObject);
        }

        // unorderer List used by Ajax autocomplete
        if ($this->autocomplete) {
            return $this->showUL();
        }
        $paginacion = $this->pagination($opt);
        $salida .= $paginacion;

        // Asigno Estilos

        $styleTable =(isset($this->Datos->styleTable))?'style="'.$this->Datos->styleTable.'"':'style="float:left;clear:left;"';
//	REMOVED DUE TO CHANGES IN BROWSER BEHAVIOUR
//        $styleTbody =(isset($this->Datos->styleTbody))?'style="'.$this->Datos->styleTbody.'"':'';

        $tableClass = $this->Datos->tipoAbm.'Class ';
        if (isset($this->Datos->swap) && $this->Datos->swap == 'true') $tableClass .= 'dnd';

        $tableProp = 'idForm="'.$form.'" xml="'.$this->Datos->xml.'" type="'.$this->Datos->tipoAbm.'" ';
        $tableProp .= 'instance="'. $this->Datos->getInstance() .'"';

        if (isset($this->Datos->xmlOrig))
            $tableProp .= ' xmlOrig="'.$this->Datos->xmlOrig.'" ';

        if (isset($this->Datos->tableBorder))
            $tableProp .= ' border="'.$this->Datos->tableBorder.'" ';

        /* IF TABLE IS UPDATEABLE*/

        if ( isset($this->hasForm) && $this->hasForm  &&
                ($this->Datos->modificaABM != 'no' && $this->Datos->modificaABM != 'false') &&
                ($this->Datos->form != 'false')) {
            $tableProp .= '  fillForm="true" ';

        }

        $customClass = $this->Datos->tableClass;

        if (isset($this->Datos->detalle)) {
            $tableProp .= 'detail="true"';
        }
        if ($opt == 'txt') {
            $salida .= '<table width="99%" >';
        } else
        if ($opt == 'noecho' || $opt == 'micro') {

            $salida .= '<table class="microTabla '.$tableClass.' '.$customClass.'" '.$tableProp.$styleTable.' >';
        } else {
            $wrapperIni = '<div class="tablewrapper" '.$styleTable.'>';
            $wrapperEnd = '</div>';
            $salida .= $wrapperIni.'<table class="sortable resizable '.$tableClass.' '.$customClass.'"
                        id="TablaInterna'.$idTabla.'"  width="100%" cellspacing="0" '.$styleTable.' '.$tableProp.'>';
        }
        // TODO put fillform event here

        $detallado = (isset($this->Datos->detallado))?$this->Datos->detallado:'';
        if ($detallado != 'false') {
            $salida .= $this->showCabecera( $opt);
            $salida .= '<tbody '.$styleTbody.' id="tbody_'.$this->Datos->idxml.'">';

            $contenido = $this->showDatos($idTabla, $opt);

            if ($contenido =='') return '';
            $salida .= $contenido;
            $salida .= '</tbody>';

        }

        if ($opt != 'micro' || $this->Datos->showTotals =="true") {

            $salida .= '<tfoot id="tfoot_'.$this->Datos->idxml.'">';
            $salida .= $this->showTotales();
            $salida .= '</tfoot>';
        }

        $salida .= '</table>'.$wrapperEnd;

        if ($opt == 'micro') return $salida;

        $tiempo = '('.processing_time($this->TIEMPO_CONSULTA).' '.$this->i18n['secs'].')';

        if (isset($this->Datos->showCantidad) && $this->Datos->showCantidad == 'false') $nocant = 'true';

        $tablaRegistros = '';
        if ($this->muestraCant && $nocant =='') {

            $tablaRegistros = '<table id="totales_'.$this->Datos->idxml.'" class="totales" border="0" width="100%" cellpadding="0" cellspacing="0">';
            $tablaRegistros .= $this->showCantidad($tiempo);
            $tablaRegistros .= '</table>';
        }

        if ($paginacion != '') {
            $salida .= $paginacion;
        } else {
            $salida .= $tablaRegistros;
        }

        if ($opt != 'noecho') {
            $idTableForm = 'id="inTableForm'.$this->Datos->idxml.'" instance="'.$this->Datos->getInstance().'"';
            $salida .= $this->inlineCrud($idTableForm, $form , $opt, $formini , $formfin, $segundaVez );

        }
	// impòrt buttons
        $salida .= $this->importDataButton();


        if ($opt != 'noecho') {
            $jscode[] = 'Histrix.registerTableEvents(\'TablaInterna'.$this->Datos->idxml.'\') ';
            $jscode[] = "sortables_init();";
            if (isset ($this->Datos->CabeceraMov))
                foreach ($this->Datos->CabeceraMov as $NCabecera => $ContCab) {
                    $xmlcabecera = ', \''.$ContCab->xml.'\'';
                }

            $resize = (isset($this->Datos->resize))?$this->Datos->resize:'';
            $resizeTable = (isset($this->resizeTable))?$this->resizeTable:'';

//            if ($resize != "false" && $resizeTable != "false")
  //              $jscode[] = "Histrix.calculoAlturas('".$this->Datos->idxml."', null ".$xmlcabecera." );";

            $salida .= Html::scriptTag($jscode);
        }
        if ($abming) {
            if ($this->Datos->xmlOrig != '')
                $xmlorig = ' original="'.$this->Datos->xmlOrig.'" ';
            $salida =  '<form instance="' . $this->Datos->getInstance() . '"  id="'.$form.'" name="'.$form.'" tipo="'.$this->Datos->tipoAbm .'" '.$xmlorig.' >' . $salida .'</form>';
        }

        if ($div==true) {

            $salida = '<div id="'.$this->Datos->idxml.'" instance="'.$this->Datos->getInstance().'">'.$salida.'</div>';
        }


        return $salida;
    }

    public function importDataButton(){
    }


    protected function inlineCrud($idTableForm, $form , $opt, $formini = '', $formfin = '')
    {
        return '';
    }

    /**
     * Muestra la cabecera de las tablas
     */
    private function showCabecera( $opt)
    {
        $hayLabels= false;
        $salidaTH = '';
        // COLS
        $salida = '<colgroup>';
        $campos = $this->Datos->camposaMostrar();

        // triangle col
        if (isset( $this->Datos->inline) && $this->Datos->inline == 'true') {
            $salida .= '<col style="width:20px;" />';
        }
        // la fila de la izquierda para subir o bajar datos
        if (isset($this->Datos->swap) && $this->Datos->swap == 'true')
            $salida .= '<col style="width:25px;" />';

        foreach ($campos as $nom => $valor) {

            $this->Suma[$nom] = 0;
            $objCampo = $this->Datos->getCampo($valor);

            if (isset($objCampo->noEmpty) && $objCampo->noEmpty == 'true' && !isset($this->Datos->hasValue[$objCampo->NombreCampo])) {

                continue;
            }

            if (isset($objCampo->Oculto))continue;

            $style 	  = (isset($objCampo->colstyle) && $objCampo->colstyle != '')? $objCampo->colstyle :'';

            $class = ( isset($this->Datos->sumaCampo[$objCampo->NombreCampo]) ||
                            isset($this->Datos->acumulaCampo[$objCampo->NombreCampo]))?'celsum':'';
            $class 	  .= (isset($objCampo->colclass) && $objCampo->colclass != '')? $objCampo->colclass :'';

            //$classsum = ($this->Datos->seSuma($valor) || $this->Datos->seAcumula($valor))?'class="celsum"':'';

            if (isset($objCampo->Parametro['noshow']) && $objCampo->Parametro['noshow'] =='true') {
                $style = 'display:none;';
                $objCampo->colstyle = 'display:none;';
                //$style = 'visibility:hidden;';

            }

            // TODO : si activo esto se VE bien en IE, pero no se como responde en los abm, controlar
            //else{

                $salida .= '<col  campo="'.$valor.'"  ';
                if ($style !='') $salida .=  'style="'.$style.'" ' ;
                if ($class !='') $salida .=  'class="'.$class.'" ' ;
                $salida .= '/>';
            //}

            if (isset($objCampo->colstyle) && $objCampo->colstyle != 'display:none;') {
                $style = (isset($objCampo->colstyle) && $objCampo->colstyle != '')? $objCampo->colstyle :'';
                $style = (isset($objCampo->thstyle) && $objCampo->thstyle != '')? $objCampo->thstyle :$style;
            }

            $th = '';
            if ( isset($objCampo->Parametro['noshow']) && $objCampo->Parametro['noshow'] == 'true') {
                $thProp = array('scope'=>'col', 'border'=>0, 'noprint' => 'true', 'size'=>0, 'style'=>$style);
            } else {
                $chk='';
                $class = 'colHeader';
                if ($objCampo->TipoDato == "check" && isset($this->enableCheckToggle) && $this->enableCheckToggle == true
                ) {
                    $class= 'checkToggle';
                }

                if (trim($objCampo->Etiqueta) == '') {
                    $thProp = array('scope'=>'col',
                            'noprint' => 'true',
                            'colName'=> $objCampo->NombreCampo,
//                            'style'=>"border:0px; heigth:0px; padding:0px; margin:0px;cellspacing:0px;",
                            'style'=>$style,

                            'class'=> $class);
                } else {
                    $hayLabels=true;
                    $print = (isset($objCampo->print) && $objCampo->print == 'false') ? 0 : 1;
                    $labelText = htmlentities(ucfirst($objCampo->Etiqueta), ENT_QUOTES, 'UTF-8');
                    if (isset($objCampo->Htmllabel))
                        $labelText = htmlentities($objCampo->Htmllabel, ENT_QUOTES, 'UTF-8');

                    $th = '<span >'.$labelText.'</span>';
                    $title = (isset($objCampo->ayuda))?$objCampo->ayuda:'';
                    $thProp = array('scope'=>'col',
                        /*'align'=>'center',*/
                            'title' => $title,
                            'style'=>$style,
                            'colName'=> $objCampo->NombreCampo,
                            'class'=> $class,
                            'print' => $print
                    );
                }
            }

            if (isset($objCampo->defaultValue))
                    $thProp['value'] = $objCampo->defaultValue;

            $salidaTH .= Html::tag('th', $th, $thProp);

        }
        //loger($salidaTH);
        // Agrego una columna para los ingresos
        if ($this->Datos->tipoAbm =='ing' || $this->Datos->tipoAbm =='grid') {
            $salida .= '<col class="helpersTabla" campo="Nro_Fila" />';
        }
        $salida .= '</colgroup>';

        $salidaHead = '<thead id="thead_'.$this->Datos->idxml.'">';
        $salidaHead .= '<tr >';

        if ((isset($this->Datos->swap) && $this->Datos->swap == 'true') || (isset( $this->Datos->inline ) && $this->Datos->inline == 'true')) {
            $salidaHead .= '<th border="0" scope="col" size="0" noprint="true"></th>';
        }
        $salidaHead .= $salidaTH;

        $salidaHead .= $this->deleteColHeader();

        $salidaHead .= '</tr>';
        $salidaHead .= '</thead>';

        $salida .=$salidaHead;

        return $salida;
    }

    protected function deleteColHeader()
    {
    }

    /**
     * Sor TempTable
     */
    public function sortTempTable()
    {
        $rev = false;
        if ($this->Datos->orden != '') {
            foreach ($this->Datos->orden as $ncampo => $campo) {
                $objCampo = $this->Datos->getCampo($campo);
                if ($objCampo) {
                    if ($objCampo->tipoOrden =='DESC') $rev = true;
                }
            }
        }
        $this->Datos->TablaTemporal->ordenar($this->Datos->orden, $rev);
    }

    public function createChart($row, $y)
    {
        //recorro cada grafico
        foreach ($this->Datos->grafico as $id_grafico => $graf) {
            if ($graf['series'] !='') {
                foreach ($graf['series'] as $nomcolserie => $valseries) {
                    if (isset($row[$nomcolserie])) {

                        if ($y == 1) $arrayserie[$nomcolserie]='';
                        else
                            $arrayserie[$nomcolserie] = $this->Datos->grafico[$id_grafico]['series'][$nomcolserie];

                        $arrayserie[$nomcolserie][($y - 1 )]= $row[$nomcolserie];
                        $this->Datos->grafico[$id_grafico]['series'][$nomcolserie] = $arrayserie[$nomcolserie];
                        $this->Datos->grafico[$id_grafico]['series'][$nomcolserie]['titulo'] = $this->Datos->getCampo($nomcolserie)->Etiqueta;
                    }
                }
            }
            if ($graf['datos'] != '')
                if (isset($row[$graf['datos']])) {

                    if ($y == 1) unset($this->Datos->grafico[$id_grafico]['valores']);

                    $this->Datos->grafico[$id_grafico]['valores'][]=$row[$graf['datos']];

                }
            if ($graf['etiquetas'] != '') {

                if ($y == 1) unset($this->Datos->grafico[$id_grafico]['leyendas']);
                if (isset($row[$graf['etiquetas']])) {
                    $valetiq = $row[$graf['etiquetas']];
                    if ($this->Datos->getCampo($graf['etiquetas'])->opcion != '') {
                        $valetiq= $this->Datos->getCampo($graf['etiquetas'])->opcion[$valetiq];
                        if(is_array($valetiq))
                            $valetiq =  current($valetiq);
                    }
                    $this->Datos->grafico[$id_grafico]['leyendas'][]=$valetiq;

                }
            }

        }
    }

    public function importButton($fieldObject , $order, $value = '')
    {
        $xmlOrig =$this->Datos->xmlOrig;
        $referente = '&amp;_xmlreferente='.$this->Datos->xml.'&amp;parentInstance='.$this->Datos->getInstance();

        $paramstring = '&amp;__row='. $order;
        $paramstring .= '&amp;__col='. $fieldObject->NombreCampo;

        if(isset($fieldObject->importacion['campos']))
            foreach ($fieldObject->importacion['campos'] as $npar => $Nomcampo) {
                $paramstring .= '&amp;_param_in[]='. $npar;
            }

		$dir = (isset($fieldObject->importacion['dir']) && $fieldObject->importacion['dir'] != '' ) ?$fieldObject->importacion['dir']:$this->Datos->dirxml;
	
        $btnImp = new Html_button($fieldObject->importacion['label'], null ,null );
        $btnImp->addEvent('onclick', 'Histrix.ventInt(\''.$xmlOrig.'\',\''.$value.'\', \''.$referente.$paramstring.'&dir='.$dir.'\', \''.$fieldObject->importacion['label'].'\', {modal:true, parentInstance: \''.$this->Datos->getInstance().'\'})');
        $btnImp->addParameter('value', $value);
        $btnImp->tabindex = $this->tabindex();

        $td  = '<td campo="'.$fieldObject->NombreCampo.'">';
        if ($fieldObject->noshow != 'true'){
            $td .= $btnImp->show();
        }
        $td .= '</td>';

        return $td;

    }

    protected function displayInnerTable($objCampo, $row, $valCampo, $orden, $x, $y , $parametros)
    {
        $tag     = (isset($this->tag))?$this->tag:'td';
        $rowTag  = (isset($this->rowTag))?$this->rowTag:'tr';

        $objCampo->refreshInnerDataContainer($this->Datos, $row);

        $objCampo->contExterno->tabindex = $this->Datos->tabindex +10;
        if (is_array($valCampo)) {
            $objCampo->contExterno->innerTablaData = $valCampo;
        }

        $objCampo->contExterno->esInterno = true;

        Histrix_XmlReader::serializeContainer($objCampo->contExterno);
        $UI = 'UI_'.str_replace('-', '', $objCampo->contExterno->tipo);
        $abmDatosDet = new $UI($objCampo->contExterno);

        if ($objCampo->contExterno->tipoAbm == 'ing') {
            $contenidoTd = $abmDatosDet->showTablaInt('micro','','','false',true, 'noform' );
        } else {

        if ($objCampo->contExterno->showProcessButton == 'true') {
                $contenidoTd = $abmDatosDet->showTabla();
            } else {

                $contenidoTd = $abmDatosDet->showTablaInt('micro', null, null, null, null, 'Form'.$this->Datos->idxml, null , $objCampo);
            }

        }

        $td = '<'.$tag.' campo="'.$objCampo->NombreCampo.'" style="padding:0px;cellspacing:0px;cellpadding:0px;margin0px;">'.$contenidoTd.'</'.$tag.'>';

        if ($objCampo->showValor=='true') {
            $tablita  = '<'.$tag.' style="vertical-align:top;"><table class="microTabla">';
            $tablita .= '<tr>';
            $tablita .=  $objCampo->renderCell($this , $objCampo->NombreCampo, $valCampo, $orden, $x, $y, $parametros);
            $tablita .= '</tr>';
            if ($contenidoTd !='')
                $tablita .= '<tr>'.$td.'</tr>';
            $tablita .= '</table></'.$tag.'>';
            $td = $tablita;
        }

        return $td;
    }

    // TODO: REFACTOR METHOD
    /**
     * Show Table Data
     * @param  string  $idTabla (UNUSED - TO BE REMOVED)
     * @param  string  $opt     - Options (to removed soon)
     * @param  integer $fila    $row
     * @return string
     */
    public function showDatos($idTabla=null, $opt=null , $fila=null)
    {
        $this->valcorte     = null;
        $listaCampos        = $this->Datos->camposaMostrar();
        $idContenedorForm   = "Form".$this->Datos->xml;
        $tag     = (isset($this->tag))?$this->tag:'td';
        $rowTag  = (isset($this->rowTag))?$this->rowTag:'tr';
        if (isset($this->fillTextArray) && $this->fillTextArray) $tag = '';
        $salida = '';
        if ($this->Datos->ordenaTemporal) {
            $this->sortTempTable();
        }

        $gridStyle = '';
        if (isset($this->Datos->grid)) {
            $gridValues = explode(',' , $this->Datos->grid);
            $gridStyle = ' style="width:'.$gridValues[0].';height:'.$gridValues[1].';" ';
        }

        $tablaTemp = $this->Datos->TablaTemporal->datos();

        if ($fila != null) {
            $temprow = $fila;

            if ($opt == 'noRowInformation')
                $temprow= 0;

            $tablaTempAux = $tablaTemp[$temprow];
            unset($tablaTemp);
            $tablaTemp[$fila] = $tablaTempAux;
            unset($tablaTempAux);
        }

        //     $paginaActual = (isset($this->Datos->paginaActual;
        $totalPaginas = count($tablaTemp);

        // inicializo los campos indexados
        $this->indices = $this->Datos->tablas[$this->Datos->TablaBase]->indices;

        // Ver para la impresion PDF
        if (isset($this->indices))
            foreach ($this->indices as $ncampoind => $campoindexado) {
                $this->camposIndexados[$campoindexado]['actual'] = NULL;
                $this->camposIndexados[$campoindexado]['anterior'] = NULL;
            }
        $hayGrafico = false;
        if (isset($this->Datos->grafico)) $hayGrafico = true;

        $y=0;
        $contadorPaginas = 0;
        $tableOrder=0;
        $close = '';
        if ($totalPaginas == 0) {
            if ($this->tipo == 'ayuda') {
                $btnclose = new Html_button($i18n['close'], "../img/cancel.png" ,$i18n['close'] );
                $btnclose->addEvent('onclick', 'cerrarVent(\'HLP'.$this->Datos->idxml.'\');');
                $btnclose->addParameter('title', $this->i18n['close']);

                $close = $btnclose->show();
                $close .='<script type="text/javascript">setTimeout(\'cerrarVent("HLP'.$this->Datos->idxml.'");\',3000);</script>';
            }
            if ($opt == 'micro') {
                return '';
            } else
                $salida .=  '<tr><td colspan="'.count($listaCampos).'"><span class="encontrados">'.$this->i18n['noRecords'].$close.'</span></td></tr>';
        } else
        if (($tablaTemp)) {
            if (isset($this->Datos->maxshowresult) && $this->Datos->maxshowresult < count($tablaTemp) &&  $this->Datos->maxshowresult !='') {
                $salida .=  '<tr><td colspan="'.count($listaCampos).'"><span class="encontrados">'.count($tablaTemp).' Registros Encontrados</span></td></tr>';
            } else
                foreach ($tablaTemp as $orden => $row) {

                    $y++;
                    if ($hayGrafico) {
                        $this->createChart($row, $y);
                    }

                    // Initialize VAriables
                    $this->det  = '';
                    $tableData  = '';
                    $bloqueado  = false;
                    $break      = false;
                    $x          = 0;
                    if (isset($rowParam))
                        unset($rowParam);

                    // Comandos para swap
                    if (isset($this->Datos->swap) && $this->Datos->swap == 'true') {
                        $tableData .= '<'.$tag.' class="dragHandle"></'.$tag.'>';
                    }

                    //$valCampo='';
                    foreach ($listaCampos as $nNombre => $nombrelista) {
                        $valCampo    ='';
                        $parametros  ='';
                        $displayType  = 'cell';
                        // Get Value
                        if (isset($row[$nombrelista]))
                            $valCampo = $row[$nombrelista];

                        // get Object (clean memory first
                        if (isset($objCampo)) unset($objCampo);
                        $objCampo = $this->Datos->getCampo($nombrelista);

                        // remove value if repeated
                        if (isset($objCampo->repeat) && $objCampo->repeat=='false') {
                            if ($orden > 0) {
                            if ($row[$nombrelista] == $tablaTemp[$orden - 1][$nombrelista])
                                $valCampo = '';
                            }
                        }

                        // for Empty columns
                        if (isset($objCampo->noEmpty) && $objCampo->noEmpty == 'true' && !isset($this->Datos->hasValue[$objCampo->NombreCampo])) {
                            continue;
                        }

                        // Set Attributes from row values
                        $objCampo->setAttributes($row);

                        // custom Style
                        if (isset($objCampo->getStyle) && $objCampo->getStyle !='') {
                            $objCampo->style = $row[$objCampo->getStyle];
                        }

                        // get LinkInt Parameters
                        if (isset($objCampo->paring) && $objCampo->paring != '') {
                            $parametros .= $this->generateLinkParameters($objCampo, $row);
                        }

                        if (isset($objCampo->clasefila) && $objCampo->clasefila  != '')
                            $rowParam['class'][] = $valCampo;

                        // Si el Campo recorrido tiene dentro un contenedor le cargo los parametros y lo muestro
                        if (isset($objCampo->contExterno) && isset($objCampo->esTabla) && $objCampo->showObjTabla == 'true') {
                            $displayType = 'innerTable';
                        }

                        if (isset($objCampo->Parametro['bloqueafila']) && $objCampo->Parametro['bloqueafila']=='true') {
                            if ($valCampo == 1 || $valCampo == 'true' )
                                $bloqueado = true;
                        }

                        // Para las grillas editables

                        $modif = '';
                        if ( isset($objCampo->editable) && $objCampo->editable == 'true') {

                            if ( (isset($objCampo->esClave) && $objCampo->esClave)) {
                                $modif = ' class="esclave" ';
                            }
                            $displayType = 'editable';

                        }

                        // Inserta en la tabla un boton para importar datos
                        if (isset($objCampo->importacion) && $objCampo->importacion != '' && $valCampo != '') {
                            $displayType = 'importButton';
                        }

                        switch ($displayType) {
                            case 'importButton':
                                $td = $this->importButton($objCampo, $orden, $valCampo);
                                break;
                            case 'editable':
                                if ($valCampo == '0000-00-00') {
                                    $valCampo = '';
                                } else {
                                    if ($objCampo->TipoDato == 'date') {
                                        $modif = ' class="field'.$objCampo->TipoDato.'"';
                                        if ($valCampo != '') {
                                            $valfecha = $valCampo;
                                            $valCampo = date("d/m/Y", strtotime($valCampo));
                                        }
                                    }
                                }

                                $this->_rowId = $orden;
                                /* update field inner containers*/

                                // reset Options
                                $deleteInnerContainer = false;
                                if ($objCampo->isSelect) {

                                    //set each datacontainer
                                    // this will refresh every data container or every row ALL the time.
                                    if ($objCampo->helperXml != '') {
                                        $xmlReader = new Histrix_XmlReader($this->Datos->dirXmlPrincipal, $objCampo->helperXml, true, $this->Datos->xml, $objCampo->helperDir,true);
                                        $micont = $xmlReader->getContainer();

                                        $micont->xml = $helperXml;
                                                $objCampo->contExterno  = $micont;
                                                $deleteInnerContainer = true;
                                    } else {

                                    }

                                        // get new options
                                    if (isset($objCampo->contExterno)) {
                                            unset($objCampo->opcion);
                                            $objCampo->refreshInnerDataContainer($this->Datos, $row);
                                            $objCampo->llenoOpciones('false',$orden);

                                    }

                                    $objCampo->customRowName = true;
                                }

                                $td ='<'.$tag.' '.$modif.' campo="'.$nombrelista.'">'.$op.$objCampo->renderInput($this, $nombreForm, $prefijoId, $valCampo).'</'.$tag.'>';

                                if ($deleteInnerContainer === true) {
                                    unset($objCampo->contExterno);
                                }

                                unset($this->_rowId);
                                break;
                                break;
                            case 'innerTable':
                            // Obtengo los Datos de los parametros definidos para el Contenedor Externo Embebido
                                $td = $this->displayInnerTable($objCampo, $row, $valCampo, $orden, $x, $y , $parametros );

                                break;
                            case 'cell':

//$objCampo->deshabilitado="true";

                                $td = $objCampo->renderCell($this , $objCampo->NombreCampo, $valCampo, $orden, $x, $y, $parametros, $tag);
                                break;

                            default:

                        }

                        // add method to show intput on toolbar
                        if (isset($objCampo->display) && strpos($objCampo->display, 'toolbar') !== false) {
                            $this->toolbarButtons[$nom] = $td;
//                            $td = '';
                        }


                        if (isset($objCampo->Oculto) && $objCampo->Oculto) continue;

                        // Break by Data
                        if (isset($this->Datos->hasBreak) && $this->Datos->hasBreak) {

                            if ($objCampo->suma=='true')
                                $partialSum[$nombrelista] = $valCampo ;
                            else $partialSum[$nombrelista] = ' ';

                            if (isset($objCampo->break) && $objCampo->break == 'true') {
                                if ($orden != 0 && $oldData[$nombrelista] != $valCampo) {
                                    $break = true;
                                }
                                $oldData[$nombrelista] = $valCampo;
                            }
                        }

                        if (isset($this->fillTextArray) && $this->fillTextArray && $objCampo->print != 'false') {
                            if ($td != '')
                                $this->textArray[$y - 1][$objCampo->Etiqueta] = $td;
                        }

                        $tableData .= $td;
                        $x++;

                    } // end each field

                    $tableData .= $this->rowButtons($orden);

                    // ROW EVENTS BLOCK
                    unset($rowEvents);
                    if (isset($this->Datos->forzado) && $this->Datos->forzado) {

                        $rowEvents['onclick'][]= 'fillForm(this, null);';
                        $rowEvents['onclick'][]= 'cerrarVent(\'HLP_aux_'.$this->Datos->idxml.'\');';
                    }

                    if (isset($this->Datos->editable) && $this->Datos->editable=='true' && $this->dlbClickEditRow != false) {
                        unset($rowEvents['onclick']);
                        $rowEvents['ondblclick'][]= 'Histrix.editRow(this);';
                    }

                    //TODO: change this to bubbling model event
                    if ($this->esAyuda == true) {
                        unset($rowEvents['onclick']);
                        $rowEvents['onclick'][]= 'cargoValor(this, \''.$this->campoRetorno.'\', \''.$this->Datos->xmlOrig.'\'  );';
                    }
                    // END ROW EVENT BLOCK

                    if ($bloqueado==true) {
                        $rowParam['block'][] ='true';
                        $rowParam['class'][] = 'grisado ';
                    }

                    // Generate detail link and parameters
                    $showDetail   = $this->hasDetail($row);
                    $detailButton = $this->detailLink($orden, $showDetail);
                    if ($showDetail) {
                        $tempParams = $this->getDetailParams($row);
                        $rowParam['detailpar']= $tempParams['detailpar'];
                        $rowParam['detaildiv']= $tempParams['detaildiv'];

                        if (isset($this->Datos->showCab) && $this->Datos->showCab == 'true') {
                            //VER
                            $rowEvents['onclick'][]= 'fillForm(this, null);';
                        }
                    }

                    // TODO : remove
                    if (isset($row['ROWID']))
                        $rowParam['id'][] = $row['ROWID'];

                    if (isset($this->Datos->onRowClick))
                        $rowParam['onRowClick'] = $this->Datos->onRowClick;

                    if (isset($this->Datos->inline) && $this->Datos->inline == 'true' && $showDetail )
                        $rowParam['class'][] = 'inline ';

		    if (isset($this->rowClass))
                        $rowParam['class'][]=$this->rowClass;

                    $rowParameters = '';
                    if (isset($rowParam)) {
                        $rowParameters = Html_input::Array2String($rowParam);
                        unset($rowParam);
                    }

                    // Break data
                    if (isset($this->Datos->hasBreak) && $this->Datos->hasBreak) {

                        if ($break) {
                            if ($this->Datos->detalle != '')
                                $breakData = '<td></td>';

                            foreach ($listaCampos as $nNombre => $nombrelista) {
                                $objCampo = $this->Datos->getCampo($nombrelista);
                                if ($objCampo->oculto=="true") continue;
                                $value = '';
                                if ($this->Datos->seSuma($nombrelista))
                                    $value = $breakTotal[$nombrelista];
                                $parametros['sum']= 'false';
                                
                                $state = $objCampo->deshabilitado;
                                $objCampo->deshabilitado = 'true';
                                $breakData .= $objCampo->renderCell($this , $nombrelista, $value , $orden, $x, $y, $parametros);
                                $objCampo->deshabilitado = $state;

                            }

                            // break by
                            $salida .=  '<tr class="partialSum" >';
                            $salida .=  $breakData;
                            $salida .=  '</tr>';
                            $breakData = '';
                            unset($breakTotal);
                        }

                        foreach ($partialSum as $n => $val) {
                            $breakTotal[$n] += $val;
                        }
                        unset($partialSum);
                    }

                    if ($fila != null) {
                        $salida .=  $detailButton.$tableData;
                    } else {
                        $jsEvents = '';
                        if (isset($rowEvents))
                            $jsEvents = Html_input::Array2String($rowEvents);

                        if ($opt == 'noRowInformation') {
                            $salida .=  $detailButton.$tableData;
                        } else {
                            $salida .=  '<'.$rowTag.'  '.$gridStyle.$rowParameters.$jsEvents.' o="'.$tableOrder.'" >';
                            $salida .=  $detailButton.$tableData;
                            $salida .=  '</'.$rowTag.'>';
                        }

                        $tableOrder++;

                    }

                    // Last break by case
                    $hasBreak = (isset($this->Datos->hasBreak))?$this->Datos->hasBreak:'false';
                    if ($orden + 1 == $totalPaginas && $hasBreak == 'true') {
                        if ($this->Datos->detalle != '')
                            $breakData = '<td></td>';

                        foreach ($listaCampos as $nNombre => $nombrelista) {
                            $objCampo = $this->Datos->getCampo($nombrelista);
                            if ($objCampo->oculto=="true") continue;
                            $value = '';
                            if ($this->Datos->seSuma($nombrelista))
                                $value = $breakTotal[$nombrelista];
                            $parametros['sum']= 'false';
                            $breakData .= $objCampo->renderCell($this , $nombrelista, $value , $orden, $x, $y, $parametros);
                        }

                        // break by
                        $salida .=  '<tr class="partialSum" >';
                        $salida .=  $breakData;
                        $salida .=  '</tr>';
                        $breakData = '';
                        unset($partialSum);
                    }

                    $clasefila = '';
                    $this->registros++;
                } // end each row

        }
        if ($hayGrafico)
            foreach ($this->Datos->grafico as $id_grafico => $grafico) {
                $_SESSION[$id_grafico]=$grafico;

            }

        return $salida;

    }
    /**
     * define if row has a related Detail
     * @param  array   $row
     * @return boolean
     */
    public function hasDetail($row)
    {
        $showDetail = false;
        if (isset($this->Datos->detalle)) {
            $showDetail = true;
            $hasDetail = $this->Datos->hasDetail;
            if ($hasDetail != '') {
                if ($row[$hasDetail] != 0)  $showDetail = true;
                else $showDetail = false;
            }
        }

        return $showDetail;
    }

    /**
     * get detail parameters for function call
     * @return Array Detail parameters
     */
    public function getDetailParams(&$row)
    {
        $div = 'Det'.$this->Datos->idxml.$this->Datos->iddetalle;

        // Propago el referente para que devuelva los valores en los ingresos externos
        // me acordare de todo esto?
        $refe   = (isset($this->Datos->xmlReferente))? '&amp;_xmlreferente='.$this->Datos->xmlReferente: '&amp;_xmlreferente='.$this->Datos->xml;
        $subdir = (isset($this->Datos->subdir) && $this->Datos->subdir != '')       ? '&dir='.$this->Datos->subdir : '';

        if (isset($this->Datos->detailDir)){
            $subdir =  '&dir='.$this->Datos->detailDir;
        }

        $detailXml = $this->Datos->detalle;
        if (isset($row[$this->Datos->detalle])) {
            $detailXml = $row[$this->Datos->detalle];
        }

        $vinDetalle = 'xmlpadre='.$this->Datos->xml.$subdir.'&amp;xmlsub=true&amp;xml='.$detailXml.$this->det.$refe;
        $vinDetalle .= '&parentInstance='.$this->Datos->getInstance();
        $rowParam['detailpar'][]= $vinDetalle;
        $rowParam['detaildiv'][]= $div;

        return $rowParam;
    }

    /**
     * Generate cell link
     * @param  integer $orden
     * @param  boolean $showDetail
     * @return string
     */
    public function detailLink($orden , $showDetail)
    {
        if (isset($this->Datos->detalle) && isset($this->Datos->inline) && $this->Datos->inline == 'true' ) {

            if ($showDetail) {
            $singleclass='';

            if ($this->Datos->inlineSingle=="true")
                $singleclass= 'single';

                $detailButton = '<td style="margin:0px;padding:0px; width:18px;"  noprint="true" campo="__inline__" valor="'.$orden.'"  class="ui-state-default  detailCell '.$singleclass.'" ><span detailCell="true" class="ui-icon ui-icon-triangle-1-e"/></td>';
            } else {
                $detailButton = '<td style="margin:0px;padding:0px; width:18px;" noprint="true" campo="__inline__" valor="'.$orden.'"  ></td>';
            }

            return $detailButton;

        }
    }

    protected function rowButtons($orden)
    {
    }

    private function showTotales($modif=null, $valorcorte='')
    {
        $clase = 'class="sortbottom"';

        $suma = 0;
        if ($modif == 'sub') $clase = 'class="subtotaltabla"';
        $salida = '<tr '.$clase.'>';
        $campos = $this->Datos->camposaMostrar();
        if (isset($this->Datos->detalle) && $this->Datos->inline == 'true')
            $salida .= '<th '.$clase.' />';

        foreach ($campos as $nom => $valor) {
            $atributos = '';
            $cellValue = '';
            $objCampo = $this->Datos->getCampo($valor);
            //$this->Datos->getCampo($valor)->Suma = 0;

            $style ='';
            if (isset($objCampo->Parametro['noshow']) && $objCampo->Parametro['noshow'] == 'true')
                $objCampo->style = 'display:none;';

            $fieldStyle     = (isset($objCampo->style) && $objCampo->style != '') ? $objCampo->style:'';
            $fieldFormStyle = (isset($objCampo->Formstyle) && $objCampo->Formstyle != '') ? $objCampo->Formstyle:'';

            if ($fieldStyle.$fieldFormStyle != '')
                $style = 'style="'.$fieldStyle.';'.$fieldFormStyle.'"';

            if (isset($objCampo->Parametro['noshow']) && $objCampo->Parametro['noshow'] == 'true')
                $objCampo->colstyle = 'display:none;';

            $style = (isset($objCampo->colstyle) && $objCampo->colstyle != '')?'style="'.$objCampo->colstyle.'"':'';

            if (isset($objCampo->Oculto)) continue;
            if (isset($objCampo->jsevaltotal)) {
                foreach ($objCampo->jsevaltotal as $campodestino => $jseval) {
                    //	$atributos .= ' jseval="'.$jseval.'" ';
//                    $atributos .= ' jsevaldest="'.$campodestino.'" ';
                    $arrayJsevaldestTotal[] = $campodestino;
                }


                //$atributos .= ' jsevaldest="'."new Array('" . implode("','", $arrayJsevaldestTotal ) . "')\" ";
                $atributos .= ' jsevaldest="'.htmlspecialchars(json_encode($arrayJsevaldestTotal)).'"';




                $atributos .= ' jsparent="'. $this->Datos->parentInstance.'"';
                unset ($arrayJsevaldestTotal);
        }

            if ($this->Datos->seSuma($valor)) {

                // al Objeto campo se modifico su valor de suma
                // Este debera estar en el objeto contenedor (?)
                if ($modif == 'sub') {
                    $cant = $this->Subtotal[$objCampo->NombreCampo];
                    $this->Subtotal[$objCampo->NombreCampo] = 0;
                } else {
                    $cant = (isset($this->Suma[$objCampo->NombreCampo]))? $this->Suma[$objCampo->NombreCampo]:0;
                }

                if (isset($objCampo->noEmpty) && $objCampo->noEmpty =='true' && $cant == 0) {
                    continue;
                }

                $salida .= '<th id="total_'.$objCampo->NombreCampo.'" align="right" '.$style.$atributos.'>';

                $this->Datos->getCampo($valor)->Suma = $cant;
                $objCampo->updateSetters($cant);
                // NO ZERO
                if (isset($objCampo->noZero) && $objCampo->noZero =='true' && $cant == 0) {
                    $cant ='';
                }
                $noshow=  (isset($objCampo->noshow))?$objCampo->noshow:'';
                if ($noshow != 'true') {
                    if ($objCampo->TipoDato == 'time') {
                        $cellValue = $cant;
                        $suma = $cant;
                    } else {
                        $xsdType= Types::getTypeXSD($objCampo->TipoDato);
                        switch ($xsdType) {
                            case "xsd:decimal" :
                                if ($cant=='') $cant = 0;
                                $formatValue = number_format($cant, 2, '.', ',');
                                $cellValue = $formatValue;
                                $suma   += $formatValue;
                                break;
                            default:
                                $cellValue = $cant;
                                $suma   += $cant;
                                break;
                        }
                    }
                }

                $salida .= $objCampo->getFormatedValue($cellValue);

                $salida .= $this->customTotalJavascript($objCampo, $cellValue);

                $salida .= "</th>";

            } else {
                $salida .= '<td class="sintotal" '.$style.'>';
                if (isset($objCampo->corte) && $objCampo->corte)
                    $salida .= $valorcorte;
              //  $salida .= '&nbsp;';
                $salida .= "</td>";
            }

            $arrayValues[$objCampo->Etiqueta]= $cellValue;

        }
        // Agrego una columna para los ingresos

        if (isset($this->fillTextArray) &&  $this->fillTextArray) {
            $this->hasTotals = true;
            $this->textArray[]= $arrayValues;
        }

        $salida .= $this->addRowCell();

        $salida .= '</tr>';

        return $salida;

        if ($suma != '')	return $salida;
        else return ''.$salida;

    }

    protected function customTotalJavascript($field, $value='')
    {
        return '';
    }

    public function updateTotals()
    {
        return '';
    }

    protected function addRowCell()
    {
        return '';
    }

    // Inline Form for Autofields
    private function autofieldsForm($autofields, $form)
    {
        $salida = '<table class="autofields">';
        foreach ($autofields as $fieldName => $objCampo) {
            $labels .= '<th>'.htmlentities(ucfirst($objCampo->Etiqueta), ENT_QUOTES, 'UTF-8').'</th>';
            $fields .= '<td>'.$objCampo->renderInput($this, $form, '', $objCampo->valor, '', $ProxObj, '').'</td>';
        }
        // Generate Column labels
        $salida .= '<tr>'.$labels.'</tr>';

        // Generate Input Fields
        $salida .= '<tr>'.$fields.'</tr>';
        $salida .= '</table>';

        return $salida;
    }

    /**
     * Show nombers of records and time
     */
    private function showCantidad($tiempo)
    {
        $salida = '';
        if (isset($this->Datos->detallado) && $this->Datos->detallado == 'false') $salida .= '<div class="paginar"><span>';

        if (isset($this->Datos->TotalRegistros))  {
    	   // $cant =  $this->Datos->TotalRegistros;
            $cant = $this->registros;

    	} else {
            $cant = $this->registros;
            $this->Datos->TotalRegistros = $cant;
        }
        if ($cant == '') $cant = 0;

        $textoprefijo = $this->i18n['records'].': ';
        if (isset($this->Datos->prefijoResultados )) $textoprefijo = $this->Datos->prefijoResultados;

        $textosufijo  = (isset($this->Datos->sufijoResultados))?$this->Datos->sufijoResultados:'';

        $salida .= '<td colspan="3">'.$textoprefijo.'<span id="COUNT'.$this->Datos->idxml.'">'.$cant.'</span>'.$textosufijo.'</td>';

        if (isset($this->Datos->detallado) && $this->Datos->detallado == 'false') {
            $salida .= '</span></div>';
        }
        if ($this->tipo != 'ing') {
            $salida .= '<td class="tiempo">'.$tiempo.'</td>';
        }
        $salida .= '</td>';

        $prop['class']='sortbottom';
        $salida = Html::tag('tr',$salida, $prop);

        return $salida;
    }

    public function showSlider($id, $retrac ='')
    {
        $iddet = $this->Datos->iddetalle;
        $style = '';
        $strStyle = '';
        if ($this->Datos->col1 != '') $style='left:'.($this->Datos->col1 - 0.1).'%';

        $uidSlide 	 = UID::getUID('slide');

        $uidSlideImg = UID::getUID('imgslide');
        $divDerecho='DIVFORM'.$this->Datos->idxml;

        if ($retrac  &&  $this->tipo != 'abm')
            $divDerecho = 'Det'.$this->Datos->idxml.$iddet;

        if ($retrac )
            $divDerecho2 = 'Det'.$this->Datos->idxml.$iddet;

        if ($style != '')
            $strStyle = ' style="'.$style.'" ';

        $salida = '<div class="slideVert"  class="ui-draggable" id="'.$uidSlide.'" '.$strStyle.' idxml="'.$id.'" rdiv="'.$divDerecho.'" rdiv2="'.$divDerecho2.'">';
        $salida .= '<div id="'.$uidSlideImg.'"><img src="../img/dragSlideVert.png" ></img></div>';
        $salida .= '</div>';

        $salida .= '<script type="text/javascript">';

        $salida .= "Histrix.resizePanel('$uidSlide');";
        $salida.='</script>';

        return $salida;
    }

}
// 2582 - 18-03-2010
                                                                