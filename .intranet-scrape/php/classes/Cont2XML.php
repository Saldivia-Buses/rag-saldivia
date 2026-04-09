<?php
/*
 * Created on 16/08/2006
 *
 * To change the template for this generated file go to
 * Window - Preferences - PHPeclipse - PHP - Code Templates
 *
 *
 * Clase que convierte un Contenedor de Datos a XML
 */



class Cont2XML {
    var $Datos;
    var $xml;
    var $xmlOrigen;
    var $todos;
    var $modificar;
    /**
     * Constructor del generador de XML del Contenedor Actual
     * $Datos: Objeto ContDatos
     * $xmlori: xml que genera la consulta
     * $todos: si devuelve todos los campos del registro o solamente los detallados
     * $modo:  responde que se necesita una consulta para determinar un registro único
     */
    public function Cont2XML($Datos ='', $xmlori='', $todos=false, $modo=false, $modificar=true, $noActBoton=false, $campoOrigen='') {
        $this->Datos = $Datos;
        $this->xmlOrigen = $xmlori;
        $this->campoOrigen = $campoOrigen;
        $this->todos = $todos;
        $this->modificar = $modificar;
        $this->noActBoton = $noActBoton;
        if ($Datos != '')
            $this->parseCont($modo);
    }


    public function generateFeed() {
        $rss = new UniversalFeedCreator();
        $rss->title = $this->Datos->tituloAbm;
        $rss->description = $this->Datos->tituloAbm;
        $https =  ($_SERVER['HTTPS']=='on')?'https':'http';
        $server = $https.'://'.$_SERVER['SERVER_NAME'];
        $rss->link = $server."/index.php";
        $rss->syndicationURL = $server."/index.php";

        $image = new FeedImage();
        $image->title = "histrix";
        $image->url = $server."/img/h05_ico.jpg";
        $image->link = $server."/index.php";
        $image->description = "Histrix";
        $rss->image = $image;

        $Tablatemp = $this->Datos->TablaTemporal->datos();
        if ($Tablatemp)
            foreach ($Tablatemp as $orden => $row) {
                $item = new FeedItem();
                foreach($row as $Nnombre => $valor) {
                    $ObjCampo = $this->Datos->getCampo($Nnombre);

                    if (count($ObjCampo->opcion) > 0) {
                        $valor = $ObjCampo->opcion[$valor];
                        if (is_array($valor)) $valor = current($valor);
                    }

                    if (isset($ObjCampo->contExterno) && $ObjCampo->esTabla && $ObjCampo->showObjTabla == 'true') {

                    // Obtengo los Datos de los parametros definidos para el Contenedor Externo Embebido
                        if ($ObjCampo->paring !='') {
                            foreach ($ObjCampo->paring as $destinodelValor => $origendelValor) {

                                $valorDelCampo = $row[$origendelValor['valor']];
                                if ($valorDelCampo == '') $valorDelCampo = '0';
                                if ($this->Datos->getCampo($origendelValor['valor'])->TipoDato == 'varchar') $comillas = '"';
                                $operador ='';
                                $operador = $origendelValor['operador'];
                                if ($operador == '')
                                    $operador ='=';
                                $reemplazo ='';
                                $reemplazo = $origendelValor['reemplazo'];
                                if ($reemplazo != 'false')
                                    $reemplazo ='reemplazo';
                                else $reemplazo = '';


                                $ObjCampo->contExterno->addCondicion($destinodelValor, $operador, $comillas.$valorDelCampo.$comillas, 'and', $reemplazo, true);

                                $ObjCampo->contExterno->setCampo($destinodelValor, $valorDelCampo);
                                $ObjCampo->contExterno->setNuevoValorCampo($destinodelValor, $valorDelCampo);

                            }
                        }

                        $ObjCampo->contExterno->esInterno = true;

                        $UI = 'UI_'.str_replace('-', '', $ObjCampo->contExterno->tipo);
                        $abmDatosDet = new $UI($ObjCampo->contExterno);

                        $opt='micro';

                        $valor = $abmDatosDet->showTablaInt($opt, null, null, null, null, 'Form'.$this->Datos->xml);

                    }


                    if ($ObjCampo->rss=='title') 		$item->title = $valor;
                    if ($ObjCampo->rss=='description') 	$item->description .= $valor;
                    if ($ObjCampo->rss=='pubDate')		$item->date =   date("r",$valor);
                    if ($ObjCampo->rss=='author')		$item->author = $valor;

                //$item->link = '';
                //$item->source = "";

                }
                $rss->addItem($item);
            }
        $this->xml = $rss->createFeed($format = "RSS2.0");
    //$rss->saveFeed("RSS1.0", "histrix.xml");
    }
    public function generateRss2() {

        $strxml  =  '<?xml version="1.0" encoding="utf-8" ?>';
        $strxml  .=  '<rss version="2.0">';
        $strxml  .=  '<channel>';
        $strxml  .=  '<title>'.$this->Datos->tituloAbm.'</title>';
        $strxml  .=  '<link>'.'</link>';
        $strxml  .=  '<description>'.'</description>';
        $strxml  .=  '<language>es-ar</language>';
        $strxml  .=  '<pubDate>'.date("D M j G:i:s T Y").'</pubDate>';
        $strxml  .=  '<lastBuildDate>'.date("D M j G:i:s T Y").'</lastBuildDate>';
        // $strxml  .=  '<docs>http://blogs.law.harvard.edu/tech/rss</docs>';
        $strxml  .=  '<generator>Histrix</generator>';
        //$strxml  .=  '<managingEditor>editor@example.com</managingEditor>';
        $strxml  .=  '<webMaster>info@espacioabierto.com.ar</webMaster>';

        $Tablatemp = $this->Datos->TablaTemporal->datos();
        if ($Tablatemp)
            foreach ($Tablatemp as $orden => $row) {
                $strxml .= '<item>';

                foreach($row as $Nnombre => $valor) {
                    $ObjCampo = $this->Datos->getCampo($Nnombre);
                    if ($ObjCampo->Oculto) continue;
                    if ($ObjCampo->noshow == 'true') continue;

                    if ($ObjCampo->tipo == 'date') {
                        $fecha = '<pubDate>'.$valor.'</pubDate>';
                    }
                    else {
                        $desc .= $valor.'<br/>';
                        $description = '<description>'.$desc.'</description>';
                    }
                }
                unset($desc);
                $strxml .= $fecha;
                $strxml .= $description;

                $strxml .= '</item>';
            }
        $strxml  .=  '</channel>';
        $strxml  .=  '</rss>';
        $this->xml = $strxml;
    }

    /**
     * Genero el xml
     */
    public function parseCont($modo) {
        $strxml  =  '<?xml version="1.0" encoding="UTF-8" ?>';
        $instance = $this->Datos->getInstance();
        if ($this->modificar == false)
            $mod=' modificar="false" ';
        else $mod=' modificar="true" ';

        $campoOrigen = ($this->campoOrigen !='')?' campoOrigen="'.$this->campoOrigen.'" ':'';
        $mod .= ($this->noActBoton)? ' noactboton="true" ':' noactboton="false" ';

        if ($modo) {
            $strxml .= '<xmlaux id="_aux_'.$this->xmlOrigen.'" instance="'.$instance.'" parentInstance="'.$this->Datos->parentInstance.'"/>';
        }
        else {
            if ($this->xmlOrigen != '') $xmlOri = ' xmlorigen="'.$this->xmlOrigen.'" ';
            $xmlpadre = ($this->Datos->xmlpadre  != '')?' xmlpadre="'.$this->Datos->xmlpadre.'" ' : '';

            $strxml .= '<resultado '.$xmlOri.' '.$mod.$campoOrigen.$xmlpadre.'  instance="'.$instance.'" parentInstance="'.$this->Datos->parentInstance.'"  >';

            if (is_array($this->Datos->tablas[$this->Datos->TablaBase]->campos)){
        	foreach ($this->Datos->tablas[$this->Datos->TablaBase]->campos as $clavecampo => $CampoT) {

            	    $Campo =&$this->Datos->getCampoRef($clavecampo);
            	    if( isset($Campo->contExterno) && isset($Campo->contExterno->isInner)) $refCont='obj="'.$Campo->contExterno->xml.'"';
            	    else $refCont='';

            	    $valor = (isset($Campo->valor))?$Campo->valor:'';

            	    // USE CDATA for special cases
            	    if (strpos($valor, '<') !== false || strpos($valor, '&') !== false){
                	$fieldValue  = '<valor><![CDATA['.$valor.']]></valor>';
            	    }
            	    else
                	$fieldValue  = '<valor>'.$valor.'</valor>';

            	    if ($this->todos === true) {
            	        $destino = $Campo->NombreCampo;
            	        $strxml .= '<campo id="'.$Campo->NombreCampo.'" destino="'.$destino.'" name="'.$destino.'" '.$refCont.'>';
            	        $strxml .= $fieldValue;
            	        $strxml .= '</campo>';
            	    }
            	    else
                	if (isset($Campo->Detalle) && $Campo->Detalle != '' ) {
                    	    if ($Campo->Detalle)
                        	foreach($Campo->Detalle as $ndest => $destino) {
                            	    $strxml .= '<campo id="'.$Campo->NombreCampo.'" destino="'.$destino.'" name="'.$destino.'" '.$refCont.'>';
                            	    $strxml .= $fieldValue;
                            	    $strxml .= '</campo>';
                        	}
                	}
        	}
	    } else {
	         $strxml .= '<TEST>NO HAY</TEST>';

	    }
            $strxml .= '</resultado>';
        }

        $this->xml = $strxml;
    }

    public function importData($xml) {
        $table = (string) $xml['table'];
        $destinationTable = (string) $xml['destinationTable'];
        $table= ($destinationTable != '')?$destinationTable:$table;
        if ($table != '') {
            $Datos = new ContDatos($table, 'IMPORTACION '.$table, 'insert');
            $Datos->tipoInsert     = (string) $xml['tipoInsert'];
            $Datos->onDuplicateKey = (string) $xml['onDuplicateKey'];

            foreach ($xml->row as $row) {

                foreach ($row->field as $fieldNumber => $field) {

                    $value = (string) $field;
                    $id    = (string) $field['id'];
                    $dataType    = (string) $field['dataType'];

                    $Datos->addCampo($id, '', '', '', $table, '');
                    $Datos->setTipo($id, $dataType ,'' ,'');
                    $Datos->setNuevoValorCampo($id, $value);
                }
                $sql[] = $Datos->getInsert();
            }
        }

        if ($sql != '') {
            foreach ($sql as $n => $sqlstring) {
                $response = updateSQL($sqlstring);
                if ($response === -1) return -1;
                $this->inserts++;
            }
        }
    }

    public function exportData() {
        $Tablatemp = $this->Datos->TablaTemporal->datos();

        $table            = $this->Datos->TablaBase;
        $destinationTable = $this->Datos->destinationTable;
        $tipoInsert       = $this->Datos->tipoInsert;
        $onDuplicateKey   = $this->Datos->onDuplicateKey;

        $salida = '<?xml version="1.0" encoding="UTF-8"?>';

        $salida .="\n";
        if ($destinationTable != '')
            $dest = ' destinationTable="'.$destinationTable.'" ';
        if ($tipoInsert != '')
            $tipo = ' tipoInsert="'.$tipoInsert.'" ';
        if ($onDuplicateKey != '')
            $onDup = ' onDuplicateKey="'.$onDuplicateKey.'" ';

        $root = (isset($this->Datos->XMLroot)?$this->Datos->XMLroot:'data');
        $salida .= '<'.$root.' table="'.$table.'" '.$dest.' '.$tipo.' '.$onDup.' >';

        $y=0;
        if ($Tablatemp == '') {
            $this->Datos->cargoTablaTemporalDesdeCampos();
            $Tablatemp = $this->Datos->TablaTemporal->datos();
        }

        $rowname = (isset($this->Datos->XMLrow)?$this->Datos->XMLrow:'row');

        if ($Tablatemp != '')
        {
            foreach ($Tablatemp as $orden => $row) {
                $y++;
                $x=0;
                $salida .='<'.$rowname.' order="'.$y.'">';

                foreach ($row as $nomcampo => $Valcampo) {

                    if ($nomcampo =='') continue;

                    $ObjCampo = $this->Datos->getCampo($nomcampo);
                    if ($ObjCampo->export == 'false') continue;
                    if ($ObjCampo->Oculto) continue;

                    if (isset ($ObjCampo->contExterno) && $ObjCampo->esTabla) {
                        $mixml = new Cont2XML($ObjCampo->contExterno);
                        $mixml->exportData();
                        $mixml->out('F', $ObjCampo->contExterno->xml, $ObjCampo->contExterno->titulo);

                        foreach ($mixml->xmlfiles as $n => $file ) {
                            $this->xmlfiles[$n] = $file;
                        }
                        continue;
                    }



                    if ($ObjCampo->norepeat == 'true')
                        $norepeat = true;
                    $valor = $Valcampo;


                    if ($this->Datos->seAcumula($nomcampo) && $norepeat != true) {
                        $valor = $Suma[$ObjCampo->NombreCampo];
                    }
                    if ($this->Datos->seSuma($nomcampo) && $norepeat != true) {
                        $Suma[$ObjCampo->NombreCampo] += $valor;
                        $Subtotal[$ObjCampo->NombreCampo] += $valor;
                    }

                    if ($this->Datos->seAcumula($nomcampo) && $norepeat != true) {
                        $valor = $Suma[$ObjCampo->NombreCampo];
                    }


                    // ultimo valor del campo
                    $ObjCampo->ultimo=$valor;

                    $valorxml= '##';

                    switch ($ObjCampo->TipoDato) {
                        case "numeric" :
                            $valorxml = $valor;
                            break;
                    }																																																																																																					    // Si tiene opciones de un combo

                    if (count($ObjCampo->opcion) > 0 && $ObjCampo->TipoDato != "check" && $ObjCampo->valop != 'true') {
                        $valor = $ObjCampo->opcion[$Valcampo];
                        if (is_array($valor)) $valor = current($valor);
                    }

                    if ($valorxml == '##') $valorxml = $valor;

                    $fieldName = (isset($ObjCampo->xmlField)?$ObjCampo->xmlField:'field');
                    $cdataOpen  = ($ObjCampo->xmlCdata=='false')?'':'<![CDATA[';
                    $cdataClose = ($ObjCampo->xmlCdata=='false')?'':']]>';

                    $valorxml = html_entity_decode($valor);

                    if ($norepeat!=true)
                        $salida .='<'.$fieldName.' id="'.$nomcampo.'" label="'.$ObjCampo->Etiqueta.'" dataType="'.$ObjCampo->TipoDato.'" >'.$cdataOpen.$valorxml.$cdataClose.'</'.$fieldName.'>';


                    $x++;
                }
                $salida .="</$rowname>\n";
            }
            $salida .="</$root>";
        }
        $this->xml = $salida;
    }


    public function buildDataArray() 
    {
        $Tablatemp = $this->Datos->TablaTemporal->datos();

        $table            = $this->Datos->TablaBase;
        $destinationTable = $this->Datos->destinationTable;
        $tipoInsert       = $this->Datos->tipoInsert;
        $onDuplicateKey   = $this->Datos->onDuplicateKey;

        if ($table != '')
            $tbl = ' table="'.$table.'"';

        if ($destinationTable != '')
            $dest = ' destinationTable="'.$destinationTable.'"';

        if ($tipoInsert != '')
            $tipo = ' tipoInsert="'.$tipoInsert.'"';

        if ($onDuplicateKey != '')
            $onDup = ' onDuplicateKey="'.$onDuplicateKey.'"';

        $root = (isset($this->Datos->XMLroot) ? $this->Datos->XMLroot : null);

        $y=0;

        if ($Tablatemp == '') 
        {
            $this->Datos->cargoTablaTemporalDesdeCampos();
            $Tablatemp = $this->Datos->TablaTemporal->datos();
        }

        $rowname = (isset($this->Datos->XMLrow) ? $this->Datos->XMLrow : null);

        if ($Tablatemp != '')
        {
            $var = [];

            foreach ($Tablatemp as $orden => $row) 
            {
                $y++;
                $x=0;
                
                foreach ($row as $nomcampo => $Valcampo) 
                {

                    if ($nomcampo == '') continue;

                    $ObjCampo = $this->Datos->getCampo($nomcampo);

                    if ($ObjCampo->export == 'false') continue;

                    if ($ObjCampo->Oculto) continue;

                    if (isset ($ObjCampo->contExterno) && $ObjCampo->esTabla) 
                    {
                        $mixml = new Cont2XML($ObjCampo->contExterno);
                        $mixml->exportData();
                        $mixml->out('F', $ObjCampo->contExterno->xml, $ObjCampo->contExterno->titulo);

                        foreach ($mixml->xmlfiles as $n => $file ) 
                        {
                            $this->xmlfiles[$n] = $file;
                        }
                        continue;
                    }

                    if ($ObjCampo->norepeat == 'true')
                        $norepeat = true;

                    $valor = $Valcampo;


                    if ($this->Datos->seAcumula($nomcampo) && $norepeat != true) {
                        $valor = $Suma[$ObjCampo->NombreCampo];
                    }
                    if ($this->Datos->seSuma($nomcampo) && $norepeat != true) {
                        $Suma[$ObjCampo->NombreCampo] += $valor;
                        $Subtotal[$ObjCampo->NombreCampo] += $valor;
                    }

                    if ($this->Datos->seAcumula($nomcampo) && $norepeat != true) {
                        $valor = $Suma[$ObjCampo->NombreCampo];
                    }


                    // ultimo valor del campo
                    $ObjCampo->ultimo=$valor;

                    $valorxml= '##';

                    switch ($ObjCampo->TipoDato) 
                    {
                        case "numeric" :
                            $valorxml = $valor;
                            break;
                    }

                    // Si tiene opciones de un combo
                    if (count($ObjCampo->opcion) > 0 && $ObjCampo->TipoDato != "check" && $ObjCampo->valop != 'true') 
                    {
                        $valor = $ObjCampo->opcion[$Valcampo];
                        if (is_array($valor)) $valor = current($valor);
                    }

                    if ($valorxml == '##') $valorxml = $valor;

                    $fieldName  = (isset($ObjCampo->xmlField)?$ObjCampo->xmlField:'field');
                    
                    $cdataOpen  = ($ObjCampo->xmlCdata=='false')?'':'<![CDATA[';
                    
                    $cdataClose = ($ObjCampo->xmlCdata=='false')?'':']]>';

                    $valorxml = html_entity_decode($valor);

                    if ($norepeat != true) $var[ $fieldName ] = $valorxml;
                    
                    $x++;
                }

                $salida[] = ( is_null($rowname) ? $var : array($rowname => $var) );

            }
        }

        $this->xml = (is_null($root) ? $salida : array($root => $salida) );

    }

    public function show() {
        return $this->xml;
    }

    public function out($dest, $fileName='', $titulo='') {
        switch($dest) {
            case 'I':
                header ("Content-type: text/xml; charset=UTF-8");
                header ("Content-Disposition:attachment; filename=\"".$titulo.".xml\"");
                echo $this->xml;
                break;

            case 'F':
                $datapath    = $_SESSION["datapath"];
                $dir = '../database/'.$datapath.'/tmp/';

                $filename = ($fileName != '')?$fileName: str_replace(' ', '_', $titulo.'.xml');
                $tempxmlfiles = $this->xmlfiles;
                
                unset($this->xmlfiles);

                $this->xmlfiles[$fileName] = $dir .$filename;
                
                if (is_array($tempxmlfiles)){
                    $this->xmlfiles = array_merge($this->xmlfiles, $tempxmlfiles);
                }
                $resource = fopen($dir .$filename, 'w+');

                fwrite($resource, $this->xml);
                fclose($resource);
                break;
        }

    }
}
