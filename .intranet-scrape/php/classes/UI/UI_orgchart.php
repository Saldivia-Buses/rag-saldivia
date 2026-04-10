<?php
/* 
 * 2009-09-09
 * tree class - Luis M. Melgratti
 */

class UI_orgchart extends UI_arbol {

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
	$this->Datos->clase = ' ';
    }

    public function showTablaInt($opt = '', $idTabla = '', $segundaVez = '', $nocant='', $div=false, $form=null, $pdf=null, &$parentObject=null) {
        $idTabla = $this->Datos->xml;
        $uid = $this->Datos->getInstance();
        $salida .= '<div id="orgchart'.$uid.'" _style="position:absolute; width:100%; top:30px; bottom:0px; overflow:auto;">';

	/*
        $js[] = "arbol".$uid." = new dTree('arbol".$uid."');";
        
        $js[] = "arbol".$uid.".add(0,-1,'".$this->Datos->getTitulo()."');";
    */
        $this->num = 0;
        
        // Reset values to load parameters
        $this->Datos->restaurarValores();

        $padre = $this->Datos->getPadre();
        $keyType = $this->Datos->getCampo($padre)->TipoDato;

        $nivel = $this->Datos->getCampo($padre)->valor;

        $this->Datos->ARBOL = new Nodo('');


        $this->generoArbol($nivel, $padre , $this->Datos->ARBOL, $uid , $keyType);
	$salida .= '</div>';

	$salida .= $this->display($this->Datos->ARBOL, $uid);
	/*
        
	    $valorcond = Types::getQuotedValue($nivel, $keyType, 'xsd:integer');        
        $this->Datos->addCondicion($padre, "=", $valorcond, ' and ', 'reemplazo');
        // pruebo asignar el valor clave del arbol
        $this->Datos->setCampo($padre, $nivel);
        $this->Datos->setNuevoValorCampo($padre, $nivel);
	*/

    //    $js[] = '$("#orgchartul'. $uid . '").orgChart({container: $("#orgchart'. $uid .'")});';
	$js[] = '$("#orgchartul'. $uid . '").jOrgChart({chartElement: $("#orgchart'. $uid .'")});';



      $salida .= Html::scriptTag($js);
        return $salida;
    }


    function display($tree, $uid) {

        $html = '<ul  id="orgchartul'.$uid.'" style="display:none;">';
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

                $rowAllowChild = 'true';

                foreach ($row as $nomcampo => $Valcampo) {

                    $modif = '';
                    $objCampo = $this->Datos->getCampo($nomcampo);

                    // get relations
                    if (isset($objCampo->arbol) && $objCampo->arbol == 'padre')
                        $parentField['name']= $nomcampo;

                    if (isset($objCampo->arbol) && $objCampo->arbol == 'hijo')
                        $parentField['value']= $Valcampo;


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

                              
                                $tdact = $objCampo->renderCell($this  , $nomcampo , $Valcampo, $this->num, $i, $this->num, $parametros, 'div');

                                break;

                            default:

                        }


                    if (isset($objCampo->rowAllowChild) && $objCampo->rowAllowChild == 'true') {

                        $rowAllowChild = $Valcampo;
                        
                    }
                    


                    //$tdact = addslashes($tdact);
                        $td .= $tdact ;

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
		
		/*
                $Descripcion = '<table class="Tree" xml="'.$this->Datos->xml.'" id="Tree_'.$this->Datos->idxml.'"  instance="' . $this->Datos->getInstance() . '" ><tbody><tr id="'.$rowID.'" o="'.$this->order.'" '.$rowParameters.'> ';
                $Descripcion .= $addButton.$detailButton.$td.'</tr></tbody></table>';
		*/
		$Descripcion = $td;
                $this->order++;

                $html .= "\n";

                $html .= '<li>';
		$html .= $Descripcion;
		/*
                $html .= '<table width="98%"><tr><td>'.$Descripcion.'</td>';
                if ($rowAllowChild == 'true')
                    $html .= '<td width="5px">'.$this->addChildButton($parentField).'</td>';

                $html .= '</tr></table>';
		*/
                if ($hijo != '') {
                    $html .= $this->display($nodo, $uid);
                }
                $html .= '</li>';

            }

            $html .= '</ul>';
        return $html;
    }


}
?>