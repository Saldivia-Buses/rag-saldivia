<?php
/* 
 * 2009-09-09
 * tree class - Luis M. Melgratti
 */

class UI_TreeView extends UI_arbol {

/**
 * User Interfase constructor
 *
 */
    public function __construct($Datacontainer) {
/*        $this->Datos     = &$Datacontainer;
        $this->tituloAbm = $Datacontainer->tituloAbm;
        $this->tipo      = $Datacontainer->tipoAbm;
*/
        parent::__construct($Datacontainer);

        $this->order     = 0;
        $this->isTree    = true;
        $registry =& Registry::getInstance();
        $this->i18n = $registry->get('i18n');

        $this->uidAutoFiltros = UID::getUID('AF',true);
    }


    // render de complete XML
    public function show($idFormulario = '', $divcont='', $opt='') {

        $id = 'Show'.$this->Datos->idxml;

        // id del contenedor (creo)
        $id2=($divcont != '')?$divcont:$id;

        $id2  = str_replace('.', '_', $id2);

        $style = $this->Datos->style;

        $clase      = 'consultaancha';

        if ($this->Datos->detalle != '' &&  $this->Datos->inLine =='false')
            $clase      = 'consulta';

        $BarraDrag  = false;
        $drag       = false;


        if ($this->contFiltro || $this->Datos->filtros || $this->Datos->autofiltro != 'false') {
            if ($this->Datos->autofiltro !='false')
                $filtros = $this->autoFiltros();
            $filtros .= $this->showFiltrosXML();
//            $script[] = "Histrix.calculoAlturas('".$this->Datos->idxml."', null ".$xmlcabecera." );";
        }
        $salidaDatos = $filtros;
        $salidaDatos    .= $this->showArbol();


        $clasedetalle = 'detalle2';
        $paramsDrag     = array('tree');

        //    $clase_der = 'class="consulta_der"';
        $clasedetalle = 'detalle';


        // Si se define explicitamente una clase en el xml
        if ($this->Datos->clase != '') {
            $clase = $this->Datos->clase;
        }

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


        if( $this->Datos->detalle != '' && $this->Datos->inLine != '') {
            $barraSlide     = $this->showSlider($id, $retrac);
            // Incorporo la barra vertical para slide
            $salida .= $barraSlide;


            // Add Detail div
            $salida .= $this->detailDiv($clasedetalle);

        }
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

        $salida = '<div id="'.$this->Datos->idxml.'" '.$fillTag.$detailTag.' class="backgroundColor contTablaInt Tree" '
            .'  style="width:100%; position:absolute; top:10px; bottom:35px;"'

            .' >'
            .$this->showTablaInt().'</div>';
        $salida .= $this->botonera();
        return $salida;
    }


    public function showTablaInt($opt = '', $idTabla = '', $segundaVez = '', $nocant='', $div=false, $form=null, $pdf=null, &$parentObject=null) {

        $idTabla = $this->Datos->xml;
        $uid=UID::getUID();
        //  $salida = '<p><a href="javascript: arbol'.$uid.'.openAll();">Expandir Todo</a> | <a href="javascript: arbol'.$uid.'.closeAll();">Contraer Todo</a></p>';
        $salida .= '<div id="arbol'.$uid.'" style="position:absolute; width:100%; top:30px; bottom:0px; overflow:auto;"></div>';
         /*  ejecuto el query */
        //    $js[] = "arbol".$uid." = new dTree('arbol".$uid."');";
        //    $js[] = "arbol".$uid.".add(0,-1,'".$this->Datos->getTitulo()."');";

        $this->num = 0;
        $padre = $this->Datos->getPadre();
        $nivel = $this->Datos->getCampo($padre)->valor;
        $this->Datos->ARBOL = new Nodo('');

        $this->generoArbol($nivel, $padre , $this->Datos->ARBOL, $uid);
	$this->level = 0;

	$Tree = $this->display($this->Datos->ARBOL, $uid);
        $Table = $this->tableDisplay();


	$salida .= '<table>';
	$salida .= '<tr>';
	$salida .= '<td>';
	$salida .= $Tree;

	$salida .= '</td>';
	$salida .= '<td>';
	$salida .= $Table;
	$salida .= '</td>';
	$salida .= '</tr>';
	$salida .= '</table>';
        //	loger(print_r($this->Datos->ARBOL, true), 'tree');
        //        die();
        $this->Datos->addCondicion($padre, "=", "'".$nivel."'", ' and ', 'reemplazo');
        // pruebo asignar el valor clave del arbol
        $this->Datos->setCampo($padre, $nivel);
        $this->Datos->setNuevoValorCampo($padre, $nivel);

        //   $js[] = '$("#arbol'.$uid.'")[0].innerHTML= arbol'.$uid.';';

//        $salida .= $Tree;
        $salida .= '</div>';
        return $salida;
    }

    function display($tree, $uid) {
        $numsup = $this->num;
        if (is_array($tree->nodos)) {
            $html .= '<ul>';

            foreach($tree->nodos as $nodo ) {
		$this->level++;
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
                    if (!($objCampo) || ($objCampo->Oculto)) continue;

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


                    $tdact = $objCampo->renderCell($this  , $nomcampo , $Valcampo,
                        $this->num, $i, $this->num, $parametros);

                    $tdact = addslashes($tdact);


		    if (isset($objCampo->subItem ) && $objCampo->subItem == 'true'){
			$this->subItems[$this->level][$nomcampo] = $tdact;
		    }
		    else 
	                    $td .= $tdact;


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

                        if ($this->Datos->inline=="true") $inline ='inline:true';

                        $rowParam['detailPar']= 'detailPar="'.$vinDetalle.'"';
                        $rowParam['detailDiv']= 'detailDiv="'.$div.'"';

                        // inline detail
                        if ($this->Datos->inline == 'true') {
                            $detailButton = '<td style="margin:0px;padding:0px;"  noprint="true" campo="__inline__" valor="'.$orden.'" class="ui-state-default ui-corner-all "><span  detailCell="true" class="ui-icon ui-icon-triangle-1-e"/></td>';
                        }

                    }
                }

                if (isset($rowParam))
                    $rowParameters = implode(' ', $rowParam);
                unset($rowParam);

                $Descripcion = '<table xml="'.$this->Datos->xml.'" id="Tree_'.$this->Datos->idxml.'"  class="Tree"><tbody><tr id="'.$rowID.'" o="'.$this->order.'" '.$rowParameters.'> ';

                $Descripcion .= $detailButton.$td.'</tr></tbody></table>';
                $this->order++;


                $html .= "\n";
                $html .= '<li>'.$Descripcion.'</li>';


                if ($hijo != '') {
                    $html .= $this->display($nodo, $uid);
                }

            }
            $html .= '</ul>';
        }
        return $html;
    }
    
	function tableDisplay(){
		$html = '<table border="1">';
		for($i = 0; $i < $this->level;$i++){
			$tds = '';
			if (isset($this->subItems[$i]))
			foreach($this->subItems[$i] as $Nom => $td){
				$tds .= $td;
			}
			$html .= '<tr>'.$tds.'</tr>';
		}
		$html .='</table>';
		return $html;
	}
}
?>
