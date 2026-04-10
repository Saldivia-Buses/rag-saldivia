<?php
/* 
 * 2009-09-09
 * tree class - Luis M. Melgratti
 */

class UI_arbol extends UI_crud {

/**
 * User Interfase constructor
 *
 */
    public function __construct($Datacontainer) {
        parent::__construct($Datacontainer);

        $this->order     = 0;
        $this->isTree    = true;
        $this->hasForm = true;
        $this->slider  = true;
        $this->formClass = 'singleForm';
        $this->hasFieldNameReference = true;
        $this->defaultClass = 'consultaing2';
    }

    // pdf printing of data 
    public function pdf($pdf , $fontsize = '', $opImpresion ='',$anchoTabla='', $posx=''){

        $pdf->SetFontSize($fontsize);           
        $arbol = $this->Datos->ARBOL;
        
        
        
        $pdf->SetY($pdf->GetY() + 4);
        $pdf->numNod=0;
        $pdf->showTree($this->Datos, $arbol, 1, 10 , true);

        /*
        $startX         = 5;
        //		$nodeFormat     = '<%k>';
        //		$nodeFormat     = '';
        //		$childFormat    = '<%k> = [%v]';
        $childFormat    = '%v';
        $w              = 100;
        $h              = 5;
        $border         = 1;
        $fill           = 0;
        $align          = 'L';
        $indent         = 3;
        $vspacing       = 1;
        */

    }
        

    // render de complete XML
    public function show($idFormulario = '', $divcont='', $opt='') {

        $id = 'Show'.$this->Datos->idxml;

        // id del contenedor (creo)
        $id2= str_replace('.', '_',($divcont != '')?$divcont:$id );

        $style = $this->Datos->style;

        $clase      = $this->defaultClass;
        $BarraDrag  = false;
        $drag       = false;
        $retrac	    = true;


        if ($this->contFiltro || $this->Datos->filtros || $this->Datos->autofiltro != 'false') {
            if ($this->Datos->autofiltro !='false')
                $filtros = $this->autoFiltros();
            $filtros .= $this->showFiltrosXML();
//            $script[] = "Histrix.calculoAlturas('".$this->Datos->idxml."', null ".$xmlcabecera." );";
        }
        $salidaDatos = $filtros;
        $salidaDatos    .= $this->showArbol();

              // Columns
        $ancho = (isset($this->Datos->ancho))?$this->Datos->ancho : '';
        $width = (isset($this->Datos->width))?$this->Datos->width : $ancho;

        if ($width != '' || ((isset($this->Datos->detalle) && $this->Datos->inline != 'true' ))) {
            $this->Datos->col1=$width;
            $this->Datos->col2=100 - $width;
            $style.='width:'.$this->Datos->col1.'%;';
        }

    
        if ((isset($this->Datos->detalle) && $this->Datos->inline != 'true' ) ||
                (isset($this->Datos->grafico) && $this->Datos->grafico != '')) {
            $clase	= 'consulta';
            $barraSlide = $this->showSlider($id, true);
        }
      
      
      


        $clasedetalle = 'detalle';
        $paramsDrag     = array('tree');

        $clase_der = 'class="'.$this->formClass.'"';

        if ($this->Datos->form != 'false'){
            $salidaAbm    = $this->showAbm(null, $clase_der);

        }
        else {
            $clasedetalle = 'detalle';
            $clase="consulta";
        }



        // Si se define explicitamente una clase en el xml
        if ($this->Datos->clase != '') {
            $clase = $this->Datos->clase;
        }


        // create Utility dragBar
     /*   if ($this->Datos->barraDrag != 'false') {

            $paramsDrag = $this->dragBarParameters();
            $paramsDrag[] = 'tree';
            $salidaDrag = $this->barraDrag2($id2,null, $paramsDrag ,$BarraDrag, null);
        }
       */

        if (isset($this->Datos->__inline) && $this->Datos->__inline == 'true') {
            $salida .= $salidaDatos;
        }
        else {
            $retorno = '';
            if (isset($this->Datos->campoRetorno)) {
                $uidRetorno = $this->Datos->getCampo($this->Datos->campoRetorno)->uid;
                $retorno = ' origen="'.$uidRetorno.'" ';
            }

            $salida .=  '<div instance="'.$this->Datos->getInstance().'" class="'.$clase.'" id="'.$id.'" style="'.$style.'" '.$retorno.'>';
    //        $salida .= '<div class="contewin" >';
            $salida .= $salidaDatos;
   //         $salida .= '</div>';
            $salida .= '</div>';
        }


        // El Abm
        $salida .= $salidaAbm;
        // Incorporo la barra vertical para slide
        $salida .= $barraSlide;


        // Add Detail div
        $salida .= $this->detailDiv($clasedetalle);

        // create Javascript functions
        //  $script[]= $customjs;
        //$script[]= 'Histrix.registerTableEvents(\'Tree_'.$this->Datos->idxml.'\', \'Tree\') ';
        $script[]= 'Histrix.registerTableEvents(\''.$this->Datos->idxml.'\') ';
        $script[]= "Histrix.registroEventos('".$this->Datos->idxml."')";

        $salida .= Html::scriptTag($script);
        return $salida;

    }


    public function showArbol() {
        $detailTag = '';
        if (isset($this->Datos->detalle) && $this->Datos->detalle != '')
            $detailTag = ' detail="true" ';

        $fillTag= ' fillForm="true" ';
        if (isset($this->Datos->form) && $this->Datos->form == 'false')
            $fillTag= '';

        $salida = '<div id="'.$this->Datos->idxml.'" instance="'.$this->Datos->getInstance().'" '.$fillTag.$detailTag.' class="backgroundColor contTablaInt Tree" '
            .'  style="width:100%; position:absolute;  bottom:35px; top:0px; overflow:auto; "'

            .' >';
        $salida .= '<div style="text-align:center" class="titulo">'.$this->Datos->getTitulo().'</div>';
        $salida .= $this->showTablaInt().'</div>';
        $salida .= $this->botonera();
        return $salida;
    }

    public function showTablaInt($opt = '', $idTabla = '', $segundaVez = '', $nocant='', $div=false, $form=null, $pdf=null, &$parentObject=null) {
        $idTabla = $this->Datos->xml;
        $uid = $this->Datos->getInstance();
        $salida  = '<div>';
        $salida .= '<a href="javascript: arbol'.$uid.'.openAll();" class="boton" style="padding:0 3px;">Expandir </a> <a href="javascript: arbol'.$uid.'.closeAll();" class="boton" style="padding:0 3px;">Contraer </a>';
        $salida .= $this->addButton(' ', "../img/add.png");
        $salida .= '</div>';
        $salida .= '<div id="arbol'.$uid.'" _style="position:absolute; width:100%; top:30px; bottom:0px; overflow:auto;"></div>';
         /*  ejecuto el query */
        $js[] = "arbol".$uid." = new dTree('arbol".$uid."');";
        
        $js[] = "arbol".$uid.".add(0,-1,'".$this->Datos->getTitulo()."');";

        $this->num = 0;
        
        // Reset values to load parameters
        $this->Datos->restaurarValores();

        $padre = $this->Datos->getPadre();
        $keyType = $this->Datos->getCampo($padre)->TipoDato;

        $nivel = $this->Datos->getCampo($padre)->valor;

        $this->Datos->ARBOL = new Nodo('');


        $this->generoArbol($nivel, $padre , $this->Datos->ARBOL, $uid , $keyType);

        $js[] = $this->display($this->Datos->ARBOL, $uid);
        //	loger(print_r($this->Datos->ARBOL, true), 'tree');
        //        die();
        
	    $valorcond = Types::getQuotedValue($nivel, $keyType, 'xsd:integer');        
        $this->Datos->addCondicion($padre, "=", $valorcond, ' and ', 'reemplazo');
        // pruebo asignar el valor clave del arbol
        $this->Datos->setCampo($padre, $nivel);
        $this->Datos->setNuevoValorCampo($padre, $nivel);

        $js[] = '$("#arbol'.$uid.'")[0].innerHTML= arbol'.$uid.';';
        
        $salida .= Html::scriptTag($js);
        return $salida;
    }



    public function generoArbol($nivel, $padre, $arbol, $uid, $keyType='varchar') {

	$nivel = Types::getQuotedValue($nivel, $keyType, 'xsd:integer');        
    
        $this->Datos->addCondicion($padre, "=", $nivel, ' and ', 'reemplazo');

        // pruebo asignar el valor clave del arbol
        $this->Datos->setCampo($padre, $nivel);
        $this->Datos->setNuevoValorCampo($padre, $nivel);
//        echo($this->Datos->getSelect());
        $this->Datos->Select();

        $this->cantCampos = _num_fields($this->Datos->resultSet);
        // Cargo tabla temporal con el resultado del select ODBC
        // Tarda un poco mas, SI, pero despues lo trato mas facil en la temporal :D

        $this->Datos->CargoTablaTemporal();
        $Tablatemp = $this->Datos->TablaTemporal->datos();
        $sumRow = null;
        if (($Tablatemp))
            foreach ($Tablatemp as $orden => $row) {
            //       $this->num++;
                $hijo = '';
                $rowArbol='';

                foreach ($row as $nomcampo => $Valcampo) {

                    $objCampo = $this->Datos->getCampo($nomcampo);
                    if (!($objCampo) || isset($objCampo->Oculto)) continue;

                    if (isset($objCampo->arbol) && $objCampo->arbol != 'padre')
                        $rowArbol[$nomcampo]=$Valcampo;

                    if (isset($objCampo->Arbol) && $objCampo->Arbol == 'hijo')
                        $hijo = $Valcampo;



                    if (isset($objCampo->treeSum) && $objCampo->treeSum =='true')
                        $sumRow[$nomcampo] += $Valcampo;
                }

                $nodo = new Nodo($rowArbol);
                $nodo->dataRow = $row;

                if ($hijo != '') {
                    if ($hijo == $nivel){
                        die('<div class="error">Recursion!</div>');
                        return;
                    }
                    $childSumRow = $this->generoArbol($hijo, $padre, $nodo, $uid, $keyType);

                    if (is_array($childSumRow))
                        foreach ($childSumRow as $name => $value) {
                            $sumRow[$name] += $value;
                            $row[$name]    += $value;
                        }
                }

                $nodo->dataRow = $row;
                $arbol->addNodo($nodo);


            }
        return $sumRow;
    }

    function display($tree, $uid) {
        $html = '';
        $numsup = $this->num;
        if (is_array($tree->nodos))
            foreach($tree->nodos as $nodo ) {
                $this->num++;
                $i = 0;
                $hijo = '';
                $Descripcion = '';
                $td = '';
                $rowID = '';
                if (isset($row['ROWID']))
                    $rowID = $row['ROWID'];
                if ($rowID == '')
                    $rowID = 'hoja_arbol';
                $idContenedorForm = "Form".$this->Datos->idxml;
                $onclick = '';

                $this->det = '';
                $renglonArbol ='';
                $arrtd='';
                $rowArbol='';
                $renglonArbol[0] = '';
                $i ++;
                $row = $nodo->dataRow;
                foreach ($row as $nomcampo => $Valcampo) {

                    $modif = '';
                    $objCampo = $this->Datos->getCampo($nomcampo);
                    if (!($objCampo) || isset($objCampo->Oculto)) continue;

                    if (isset($objCampo->arbol) && $objCampo->arbol != 'padre')
                        $rowArbol[$nomcampo]=$Valcampo;
                    $parametros  ='';

                    // External links
                    if (isset($objCampo->paring) && $objCampo->paring != '') {
                        $parametros .= $this->generateLinkParameters($objCampo, $row);
                    }

                    if (isset($objCampo->Arbol) && $objCampo->Arbol == 'hijo')
                        $hijo = $Valcampo;

                    $renglonArbol[0] .= ' - '.$Valcampo;


                        $displayType = 'cell';
                        // Si el Campo recorrido tiene dentro un contenedor le cargo los parametros y lo muestro
                        if (isset($objCampo->contExterno) && isset($objCampo->esTabla) && $objCampo->showObjTabla == 'true') {
                            $displayType = 'innerTable';
                        }


                        if (isset($objCampo->Parametro['bloqueafila']) && $objCampo->Parametro['bloqueafila']=='true') {
                            if ($Valcampo == 1 || $Valcampo == 'true' )
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
                        if (isset($objCampo->importacion) && $objCampo->importacion != '' && $Valcampo != '') {
                            $displayType = 'importButton';
                        }


                        switch($displayType) {
                            case 'importButton':
                                $tdact = $this->importButton($objCampo, $orden, $Valcampo);
                                break;
                            case 'editable':
                                if ($Valcampo == '0000-00-00') {
                                    $Valcampo = '';
                                }
                                else
                                if ($objCampo->TipoDato == 'date' && $Valcampo != '' ) {
                                    $valfecha = $Valcampo;
                                    $Valcampo = date("d/m/Y", strtotime($Valcampo));
                                }
                                $this->_rowId = $orden;
                                $tdact = '<td '.$modif.' campo="'.$nombrelista.'">'.$objCampo->renderInput($this, $nombreForm, $prefijoId, $Valcampo).'</td>';
                                unset($this->_rowId);
                                break;
                                break;
                            case 'innerTable':
                            // Obtengo los Datos de los parametros definidos para el Contenedor Externo Embebido
                                $tdact = $this->displayInnerTable($objCampo, $row, $Valcampo, $orden, $x, $i , $parametros );

                                break;
                            case 'cell':

                              
                                $tdact = $objCampo->renderCell($this  , $nomcampo , $Valcampo, $this->num, $i, $this->num, $parametros, 'td');

                                break;

                            default:

                        }




                    $tdact = addslashes($tdact);
                    $td .= $tdact;
                    unset($tdact);
                    if (isset($objCampo->Parametro['bloqueaFila']) && $objCampo->Parametro['bloqueaFila']=='true') {
                        if ($Valcampo == 1 || $Valcampo == 'true' || $Valcampo == true)
                            $onclick = '';
                    }
                }

                $arrtd=substr($arrtd,1,40);

                $detailButton = '';
                if ($this->Datos->detalle != '') {
                    $showDetail = true;
                    $hasDetail = $this->Datos->hasDetail;
                    if ($hasDetail != '') {
                        if ($row[$hasDetail] != 0)  $showDetail = true;
                        else $showDetail = false;
                    }
                    $detaltag = '';

                    if ($showDetail) {

                        $div = 'Det'.$this->Datos->idxml.$this->Datos->iddetalle;

                        // Propago el referente para que devuelva los valores en los ingresos externos
                        // me acordare de todo esto?

                        if (isset($this->Datos->xmlReferente))
                            $refe = '&amp;_xmlreferente='.$this->Datos->xmlReferente;
                        else $refe = '&amp;_xmlreferente='.$this->Datos->xml;
                        if ($this->Datos->subdir != '') $refe .= '&amp;dir='.$this->Datos->subdir;
                        $vinDetalle = 'xmlpadre='.$this->Datos->xml.'&amp;xmlsub=true&amp;xml='.$this->Datos->detalle.addslashes(addslashes($this->det)).$refe;
                        $vinDetalle .= '&parentInstance='.$this->Datos->getInstance();
                        
                        if ($this->Datos->inline=="true") $inline ='inline:true';

                        $rowParam['detailPar']= 'detailPar="'.$vinDetalle.'"';
                        $rowParam['detailDiv']= 'detailDiv="'.$div.'"';

                        // inline detail
                        if ($this->Datos->inline == 'true' ) {
                            $detailButton = '<td style="margin:0px;padding:0px;width:18px;"  noprint="true" campo="__inline__"  class="ui-state-default ui-corner-all "><span  detailCell="true" class="ui-icon ui-icon-triangle-1-e"/></td>';
                        }

                       //     $addButton = '<td style="margin:0px;padding:0px;width:18px;"  noprint="true" _campo="__inline__"  class="ui-state-default ui-corner-all "><span class="ui-icon ui-icon-plus"/></td>';
                    }
                }
                if (isset($rowParam))
                    $rowParameters = implode(' ', $rowParam);
                unset($rowParam);

                $Descripcion = '<table xml="'.$this->Datos->xml.'" id="Tree_'.$this->Datos->idxml.'"   instance="' . $this->Datos->getInstance() . '"  class="Tree"><tbody><tr id="'.$rowID.'" o="'.$this->order.'" '.$rowParameters.'> ';

                $Descripcion .= $addButton.$detailButton.$td.'</tr></tbody></table>';
                $this->order++;


                $html .= "\n";
                $html .= "arbol".$uid.".add(".$this->num.",".$numsup.", '".$Descripcion."' );";


                if ($hijo != '') {
                    $html .= $this->display($nodo, $uid);
                }

            }


        return $html;
    }
}
?>