<?php
/*
 * 2009-09-09
 * Base class for UI construction
 */

class UI extends Histrix
{
/**
 * User Interfase constructor
 *
 */
    public function __construct(&$dataContainer)
    {
        parent::__construct();

        $this->Datos     = $dataContainer;
        $this->tituloAbm = (isset($dataContainer->tituloAbm))?$dataContainer->tituloAbm:'';
        $this->tipo      = (isset($dataContainer->tipoAbm))?$dataContainer->tipoAbm:'';

        $this->disabledCheckDefault = true;
    }

    public function initialJavascript()
    {
        $js = '';
        if (isset($this->Datos->autoupdate)) {

            $js="Histrix.periodicalUpdate('#tbody_".$this->Datos->idxml."', {
                    url : 'process.php?accion=autoupdate&instance=".$this->Datos->getInstance()."&xmldatos=".$this->Datos->xml."',
                    minTimeout:".$this->Datos->autoupdate."});";
        }

        return $js;
    }

    public function getTitulo()
    {
        return $this->tituloAbm;
    }

    public function setTitulo($tit)
    {
            $this->tituloAbm = $tit;
    }


    // pdf printing of data 
    public function pdf($pdf , $fontsize = '', $opImpresion ='',$anchoTabla='', $posx=''){
	//	    $abmDatosDet->pdf($this, $fontsize, $pdfwidth, $objCampo->PDFsize, $objCampo->posx);
//        $this->impTabla($objCampo->contExterno, null, $opImpresion, $pdfwidth, $objCampo->PDFsize, $objCampo->posx);
    
        $pdf->impTabla($this->Datos, null, $opImpresion, $anchoTabla, $fontsize, $posx);        
    }

    // render de complete XML

   public function show($idFormulario = '', $divcont='', $opt='')
   {
      /* si es un detalle de otra consulta no pongo el div principal */

        $id = 'Show'.$this->Datos->idxml;


        // id del contenedor (creo)
        $id2=($divcont != '')?$divcont:$id;

        $id2  = str_replace('.', '_', $id2);

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

        $clase		 = 'consultaing2';
        $barraDrag	 = false;
        if (($this->Datos->detalle !='' && $this->Datos->inline != 'true' ) ||
            $this->Datos->grafico != '') {
            $clase	= 'consulta';
            $barraSlide = $this->showSlider($id, $retrac);
        }

        // Si se define explicitamente una clase en el xml
        if ($this->Datos->clase != '') {
            $clase = $this->Datos->clase;
        }

        // display Table
        $salidaDatos = $this->showTabla();

        $customjs    = 'Histrix.registerTableEvents(\'TablaInterna'.$this->Datos->idxml.'\');';

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
            $salida .= '<div class="contewin ui-widget ui-widget-content" >';
            $salida .= $salidaDatos;
            $salida .= '</div>';
            $salida .= '</div>';
        }

        // Incorporo la barra vertical para slide
        $salida .= $barraSlide;

        if ($this->Datos->detalle != '' && $this->Datos->inline != 'true') {
        // solo muestro la cabecera de lo consultado si se pide explicitamente
            if ($this->Datos->showCab == 'true')
                $salidaDetalle .= $this->showAbm('readonly');

            if ($this->Datos->col1 != '') $style_2='left:'.($this->Datos->col1 + 0.5).'%';
            $salidaDetalle .= '<div  class="'.$clasedetalle.'" id="Det'.$this->Datos->idxml.$this->Datos->iddetalle.'"   style="'.$style_2.'">';
            $salidaDetalle .= '<div class="esperareloj ui-widget ui-corner-all ui-widget-header" >'.$this->i18n['selectRowToDetail'].'</div>';
            $salidaDetalle .= '</div>';

            $salida .= $salidaDetalle;
        }

        // Graficos
        if ($this->Datos->grafico != '') {
            $salida .= $this->showGraficos();
        }

        // create Javascript functions
        $script[]= $customjs;
        $script[]= "Histrix.registroEventos('".$this->Datos->idxml."')";

//        if ($this->Datos->resize != "false")
  //          $script[]= "Histrix.calculoAlturas('".$this->Datos->idxml."', null ".$xmlcabecera." ); ";

        $salida .= Html::scriptTag($script);

        return $salida;

    }

    public function showGraficos()
    {
        $tmpbase= '../database/'.$_SESSION['datapath'].'/tmp/';
    
	//style="background-image:url(\''.$tmpbase.$id_grafico.'.png\');"
        foreach ($this->Datos->grafico as $id_grafico => $grafico) {
            $paramsDrag= array('idgraf');
            $img_window_open  = '<div class="grafico" id="'.$id_grafico.'"   >';
            $img_window_open  .= $this->barraDrag2($id_grafico, $grafico['titulo'], $paramsDrag, true,$id_grafico);
//-            $img_window_open  .= '<div class="contewin ui-widget">';
            $img_window_open  .= '<div class="contewin ui-widget" >';

            $width= (isset($grafico['ancho']))?'width="'.$grafico['ancho'].'px" ':'';
            $height= (isset($grafico['alto']))?'height="'.$grafico['alto'].'px" ':'';
            $image = '<img class="chart" id="IMG'.$id_grafico.
                '" src="grafico.php?grafico='.$id_grafico.
                '&amp;uid='.UID::getUID().'" alt="'.
                $this->i18n['loading'].
                '..." tittle="'.$grafico['titulo'].'" '.$width.' '.$height.
                //           'OnMouseMove="getMousePosition(event);" OnMouseOut="nd();"'.
                ' >';

            $img_window_close  = '</div></div>';

            if ($this->tipo != 'chart') {
                $salida  = $img_window_open.$image.$img_window_close;
            } else {
        	$salida  = '<div style="_position:absolute; height:100%;width:100%;background-image:url(\''.$tmpbase.$id_grafico.'.png\');background-repeat:no-repeat;">';
        	$salida .= $image;
        	$salida .= '</div>';
    	    }
    	    


            $script[]="$('#$id_grafico').draggable({ handle: '#dragbar$id_grafico'});";
        //  $script[]='LoadImageMap("IMG'.$id_grafico.'","grafico.php?action=GetImageMap&grafico='.$id_grafico.'");';

        }

        $salida .= Html::scriptTag($script);

        return $salida;
    }

    /*
     * Muestra los filtros de las consultas
     * (Rehacer para no pasar dos veces para implementar los grupos)
     */

    public function tabNumber()
    {
        $this->Datos->tabindex++;

        return $this->Datos->tabindex;
    }

    public function tabindex()
    {
        $tabindex = ' tabindex="'.$this->tabNumber().'" ';

        return $tabindex;
    }

    public function dragBarParameters()
    {
        //$parameters= array('autofiltros', 'graf');

        // RSS
        if ($this->Datos->rss != '')  $parameters[]='rss';

        // Add PrintOptions in top bar
//        if ($this->Datos->imprime != 'false') $parameters[]='printOptions';

        // Barra  superior para arrastrar y para filtros
        //if ($this->Datos->log != '') $parameters[]='log';
        return $parameters;
    }

    public  function barraDrag2($idContenedor, $titop='', $opt='', $drag=true, $idgraf=null)
    {
      //  $minimize = '';
      //  $maximize = '';
        $imgmove = '';
        $rss = '';
        $close = '';
        $estilo = '';

        if ($drag) {

            $close = new Html_button(null, "../img/icon_close_u.png" ,"Cerrar" );
            $close->addEvent('onclick', 'cerrarVent(\''.$idContenedor.'\')');
            //$close->addParameter('title', $this->i18n['close']);

            $close = $close->show();

            if ($titop == '') {
                if (is_object($this->Datos))
                    $titop = htmlentities(ucfirst($this->Datos->getTitulo()), ENT_QUOTES, 'UTF-8');
            }

        }
        if (!is_array($opt))
            $opt = array($opt);
        $graf = '';
        $rss = '';
        if (is_object($this) && $this->disableToolbar !== true) {
            if (in_array('idgraf', $opt))
                $graf  = '<img style="cursor:pointer;" src="../img/x-office-spreadsheet.png"  alt="gr&aacute;ficos" title="Modificar Gr&aacute;fico"  onClick="addGraficos(\''.$this->Datos->xml.'\',  \''.$idgraf.'\');" />';
            if (in_array('rss', $opt)) {
                $rsslink = 'exportrss.php?db=saldivia&user='.$_SESSION['usuario'].'&dir='.$this->Datos->subdir.'&xmldatos='.$this->Datos->xml.'&p='.md5($_SESSION['pass']);
                $rss = '<a style="border:0px;" target="blank" href="'.$rsslink.'" title="Puede suscribirse a este enlace mediante un lector de noticias" >'.'<img  style="border:0px;" src="../img/rss.png" />'.'</a>';
            }
        }

        if ($drag) $estilo = 'style="cursor:move;"';
        $salida = '<div id="dragbar'.$idContenedor.'" class="barrasup"  '.$estilo.'>';
        if (in_array('filter', $opt)) {
         //    $salida .= $this->i18n['filter'].'<input name="filter" type="text" style="height:15px; float:left;margin-left:20px;"/>';
        }
        $salida .= $titop;
        $salida .= '<span class="mover">'.$imgmove.$graf.$rss.'</span>';
        $salida .= '<span class="buttons">'.$close.'</span>';
        $salida .= '</div>';

        return $salida;
    }

    public function xmlEditor()
    {
        $file = $this->Datos->xml;

        $helpId  = $this->Datos->getHelpLink();

        $imgEditor ='<a target="_blank" class="internalLink" href="'.$helpId.'" >';
        $imgEditor .= '<img src="../img/Manual.png"   title="Manual"  _onClick="'.$execSql.'" />';

//        $imgEditor ='<a target="_blank" class="boton" icon="ui-icon-help" text="false" href="'.$helpId.'" title="Manual">';
        //$imgEditor .= '<div class="boton " icon="ui-icon-help" text="false" src="../img/Manual.png"   title="Manual"  _onClick="'.$execSql.'" />';
        $imgEditor .= '</a>';

        if ($_SESSION['EDITOR'] == 'editor') {
            if ($this != '') {
                $dir = $this->Datos->dirxml;
                $titulo = 'EDITOR '.$file;
                $titulo2 = 'Last SQL '.$file;
                $execEditor = "xmlLoader('Edit_$file', '&url=codeEditor.php&file=$file&dir=$dir', {title:'$titulo', loader:'cargahttp.php'});";

                $execSql    = "Histrix.loadInnerXML('Debug_$file', 'debug.php?xml=$file&instance=".$this->Datos->getInstance()."', '$titulo2', '$titulo2',null,null, 'cargahttp.php');";

                                $imgEditor  .= '<img style="cursor:pointer;" src="../img/edit.png"   title="Editar: '.$dir.'/'.$file.'"  onClick="'.$execEditor.'" />';
                                $imgEditor .= '<img style="cursor:pointer;" src="../img/sql_ico.jpg"   title="SQL: '.$dir.'/'.$file.'"  onClick="'.$execSql.'" />';

                //$imgEditor .= '<div class="boton " icon="ui-icon-pencil" text="false" title="Editar: '.$dir.'/'.$file.'"  onClick="'.$execEditor.'" />';
                //$imgEditor .= '<div class="boton " icon="ui-icon-script" text="false" title="SQL: '.$dir.'/'.$file.'"  onClick="'.$execSql.'" />';

            }
        }

        if ($_SESSION['administrator'] == 1) {
            $history    = "Histrix.loadInnerXML('log_cons_xml', 'histrixLoader.php?xml=log_cons.xml&amp;xmlprog=".$this->Datos->xml."&amp;dir=histrix/menu&amp;menuId=".$this->Datos->_menuId."', '','History', 'DIVlog_cons.xml','History' , {width:'580px', height:'260px'});";
            $imgEditor .= '<img style="cursor:pointer;" src="../img/appointment-new.png"  title="History '.$file.'"  onClick="'.$history.'" />';
//            $imgEditor .= '<div class="boton " icon="ui-icon-clock" text="false" title="History '.$file.'"  onclick="'.$history.'" />';
                                
            $permisions = "Histrix.loadInnerXML('profileAuthItem_xml', 'histrixLoader.php?xml=profileAuthItem.xml&amp;dir=histrix/menu&amp;menuId=".$this->Datos->_menuId."', '','Permisos', 'DIVprofileAuthItem.xml','Permisos' , {width:'580px', height:'260px'});";
            $imgEditor .= '<img style="cursor:pointer;" src="../img/emblem-readonly.png"  title="Permissions '.$file.'"  onClick="'.$permisions.'" />';
//                $imgEditor .= '<div  class="boton " icon="ui-icon-locked" text="false" title="Permissions '.$file.'"  onclick="'.$permisions.'" />';

            $userpermisions = "Histrix.loadInnerXML('userAuthItem_xml', 'histrixLoader.php?xml=userAuthItem.xml&amp;dir=histrix/menu&amp;menu_id=".$this->Datos->_menuId."', '','Usuarios', 'DIVuserAuthItem.xml','Permisos' , {width:'580px', height:'260px'});";
            $imgEditor .= '<img style="cursor:pointer;" src="../img/system-users.png"  title="Permissions '.$file.'"  onClick="'.$userpermisions.'" />';

                // $imgEditor .= '<div class="boton " icon="ui-icon-person" text="false"src="../img/system-users.png"  title="Permissions '.$file.'"  onclick="'.$userpermisions.'" />';

            $notification = "Histrix.loadInnerXML('htxnotifuser.xml' , 'histrixLoader.php?xml=htxnotifuser.xml&amp;dir=histrix/menu&amp;menuId=".$this->Datos->_menuId."', '','Notificaciones', 'DIVhtxnotifuser.xml','Notification' , {width:'580px', height:'260px'});";
            $imgEditor .= '<img style="cursor:pointer;" src="../img/mail_generic.png"  title="Notifications '.$file.'"  onClick="'.$notification.'" />';

//                $imgEditor .= '<div class="boton " icon="ui-icon-mail-closed"  text="false" src="../img/mail_generic.png"  title="Notifications '.$file.'"  onclick="'.$notification.'" />';

        }

        /*
        if (isset($this->Datos->helplink)) {
            $filename = '../files/'.trim($this->Datos->helplink);
            $path = dirname($filename);
            $file = basename($filename);
            $title= $file;
            $file = urlencode  ( $file  );
            $sessionvar = uniqid('prev');
            $_SESSION[$sessionvar]= $path;
            $downlink =  'pdfviewer.php?dir='.$sessionvar.'&f='.$file.'&ancho=650&alto=500&tipo=doc';
            $onclick = "Histrix.loadInnerXML ('$sessionvar', '$downlink', null, '$title',  null, null,  {width:'80%', height:'90%', modal:true})";

            $viewer = new Html_button('ISO 9000', '../img/question.gif');
            $viewer->addParameter('title', 'Preview');
            $viewer->addEvent('onclick', $onclick);

            $imgEditor .= $viewer->show();
        }
        */

        $output = '<div class="helper_buttons">'.$imgEditor.'</div>';
        return $output;
    }

    public function generateLinkParameters($objCampo, $row)
    {
        $parametros = '';
        if ($objCampo->paring != '') {
            foreach ($objCampo->paring as $destino => $ncampo) {
                $extras='';
                $fieldValorDestino = $ncampo['valor'];
                $operador	  = $ncampo['operador'];
                $reemplazo	  = $ncampo['reemplazo'];

                $valorDestino = (isset($row[$fieldValorDestino]))?$row[$fieldValorDestino]:$fieldValorDestino;
                if ($operador  != '') $extras .='&amp;__OP__'.$destino.'='.htmlentities(urlencode($operador));
                if ($reemplazo != '') $extras .='&amp;__RE__'.$destino.'='.htmlentities(urlencode($reemplazo));
		
		if ($operador == 'in' && ($valorDestino == $fieldValorDestino  || $valorDestino == ''  )) {
		    continue;
		}
                $parametros .= '&amp;'.$destino.'='.urlencode($valorDestino).$extras;
            }

            return $parametros;
        }
    }

    public function printButtons()
    {
            $formPrint = 'FormPrint'.$this->Datos->idxml;
            $salida .= '<span style="float:left;clear;right; top:0px; padding:0px;" >';
            $salida .= $this->btnExportar();
            $salida .= '</span> ';
            $salida .= '<span style="float:left;clear;right; top:0px; padding:0px;" >';
            $salida .= $this->btnOptions();
            $salida .= '</span> ';

            if (isset($this->Datos->notifica) && $this->Datos->notifica =='true' )
                $salida .= $this->btnNotificar();

            if (isset($this->Datos->imprimetanda) && $this->Datos->imprimetanda == 'true') {
                $btnImpTanda = new Html_button($this->i18n['printBatch'], "../img/print_class.png" ,$this->i18n['print'] );
                $btnImpTanda->addEvent('onclick', 'Histrix.imprimirTandapdf(\''.$this->Datos->xml.'\' , \''.$formPrint.'\' , \''.$this->Datos->subdir.'\', \''.$this->Datos->getInstance().'\' )');
                $btnImpTanda->tabindex = $this->tabindex();
                $salida .= $btnImpTanda->show();
            }


            $btnImprimir = new Html_button($this->i18n['print'], "../img/printer1.png" ,$this->i18n['print'] );
            
            //$salida .= '<div>';
            //$btnImprimir = new Html_button($this->i18n['print'], "" ,$this->i18n['print'] );
            $btnImprimir->addEvent('onclick', 'Histrix.imprimirpdf(\''.$this->Datos->xml.'\', \''.$this->Datos->getTitulo().'\', \''.$formPrint.'\' , null, null, \''.$this->Datos->getInstance().'\'  )');
            $btnImprimir->addParameter('class', 'printButton');
            $btnImprimir->addParameter('icon', 'ui-icon-print');        
            $btnImprimir->tabindex = $this->tabindex();

            $salida .= $btnImprimir->show();

            $btnImpOpc = new Html_button(null, "../img/down.png" , $this->i18n['print']);
            //$btnImpOpc = new Html_button('options', null, $this->i18n['print']);

            $btnImpOpc->addEvent('onclick', 'Histrix.showopprint(this, \''.$formPrint.'\')');
            $btnImpOpc->addParameter('class', 'printOptionButton');
            //$btnImpOpc->addParameter('class', 'printOpButton');
            $btnImpOpc->addParameter('icon', 'ui-icon-triangle-1-s');      
            $btnImpOpc->addParameter('text', 'false');                      
            $salida .= $btnImpOpc->show();
            //$salida .= '</div>';
            $salida .= $this->printOptions($formPrint);



            return $salida;
    }

    public function botonera($buttons='')
    {
        $salida = '';
        $doPrint = (isset($this->Datos->imprime))?$this->Datos->imprime:'';
        if ($doPrint != 'false' || $buttons != '') {

            $salida = '<div  class="botoneraImp ui-accordion-header ui-state-default"  id="botonera'.$this->Datos->idxml.'">';

            // custom buttons
            $salida .= $buttons;

            if ($doPrint != 'false')
                $salida .= $this->printButtons();

            $salida .= '</div>';
        }

        return $salida;

    }

    public function printersSelect($printForm=null)
    {
        $printers['']='PDF';
        $sessionPrinters = $_SESSION['PRINTERS'];
        $count= 0;
        if (is_array($sessionPrinters))
        foreach ($sessionPrinters as $printerName) {
            if ($printerName != '') {
            $printers[$printerName]=$printerName;
            $count++;
            }
        }
        if ($count == 0 ) return;

        $printerSelect = new Html_select($printers, $this->Datos->defaultPrinter);
        $printerSelect->addParameter('name', 'printername');
        $printerSelect->addParameter('class', 'selectmenu');        

        if ($printForm != '') {
            $printerSelect->addEvent('onchange', 'Histrix.togglePrint(this, \''.$printForm.'\', false);');
        }
        $txt  =  $printerSelect->show();
        $txt .=  '<hr>';

        return $txt;

    }


    public function printOptions($printForm=null) {

        $salida = '<div class="printForm optionPanel ui-widget ui-widget-content ui-corner-all ui-state-default" id="ORI'.$printForm.'" style="display:none;z-index:100;">';

        if ($printForm!= null)
            $salida .= '<form id="'.$printForm.'" >';
        $salida .= '<fieldset>';
        $salida .= $this->printersSelect($printForm);

 //       $salida .= '<br>';

        $valor= (isset($this->Datos->PDForientacion) && $this->Datos->PDForientacion != '')? $this->Datos->PDForientacion : 'P';
        $valor= (isset($this->Datos->PDForientation) && $this->Datos->PDForientation != '')? $this->Datos->PDForientation : $valor;
        $opciones['P']='P';
        $opciones['L']='L';

        $extra['P'] = '<img src="../img/portrait.png"  alt="'.$this->i18n['portrait'].'"  title="'.$this->i18n['portrait'] .'">';
        $extra['L'] = '<img src="../img/landscape.png" alt="'.$this->i18n['landscape'].'" title="'.$this->i18n['landscape'].'">';
        $radio = new Html_radio($opciones, $valor, $extra);

       // $radio->Parameters = $arrayAtributos;
        $radio->addParameter('type', 'radio');
        $radio->addParameter('name', '__orientacion');
        $radio->addEvent('onchange', '$(\'#ORI'.$printForm.'\').slideToggle();');
        $salida .= '<div>'.$radio->show().'</div>';
        $printOp = 'Histrix.headerOptions(\''.$this->Datos->xml.'\');';
 //       $imgPrintOptions = '<img style="cursor:pointer;" src="../img/printer2.png"  title="ColumnOptions" onClick="'.$printOp.'" />';
        $salida .=  '<hr>';

        $salida .=  $this->PDFpageSize();
        $salida .=  '<hr>';

        $btnOpt = new Html_button($this->i18n['chooseColumns'], '' );
        $btnOpt->addEvent('onclick', $printOp);
        $salida .=  $btnOpt->show();
        $salida .=  '<hr>';

	if (Histrix_Functions::hasInternetConnection()){
        //$btn = new Html_button($this->i18n['send'], "../img/mail_generic.png" ,$this->i18n['send']);
        $btn = new Html_button($this->i18n['send'], '',$this->i18n['send']);
        $btn->addEvent('onclick', 'Histrix.imprimirpdf(\''.$this->Datos->xml.'\', \''.$this->Datos->getTitulo().'\', \''.$printForm.'\',\'send\',  null, \''.$this->Datos->getInstance().'\'  )');
        $btn->addParameter('icon', 'ui-icon-mail-closed');      

        if ($_SESSION['email'] == ''){
            $btn->addParameter('disabled', 'disabled');
            $btn->addParameter('class', 'ui-state-disabled');
            $btn->addParameter('title', $this->i18n['emaildisabled']);
        }


        $salida .= $btn->show();
        
	}
        $salida .= '</fieldset>';
        if ($printForm!= null)
            $salida .= '</form>';

        $salida .= '</div>';

        return $salida;
    }

    /**
     *
     * @return string page size options radio buttons
     */
    private function PDFpageSize()
    {
        $value= (isset($this->Datos->PDFpageSize) && $this->Datos->PDFpageSize != '')? $this->Datos->PDFpageSize : 'A4';

        $opciones['A3']='A3';
        $opciones['A4']='A4';
        $opciones['A5']='A5';
        $opciones['Legal']='Legal';
        $opciones['Letter']='Letter';

        $input = new Html_select($opciones, $value);
        //$radio->addParameter('type', 'radio');
        $input->addParameter('name', '__pagesize');
    //  $input->addEvent('onchange', '$(\'#ORI'.$printForm.'\').slideToggle();');
        $output = $this->i18n['size'].': ';
        $output .= $input->show();

        return $output;
    }

    public function cantCampos()
    {
        return count($this->Datos->camposaMostrar());
    }

    public function linkButton($objCampo, $valor, $valfecha, $params, $formatoCampo)
    {
        $title 	  = $objCampo->linkdes;

//        if (isset($objCampo->tooltip) ) $title = $objCampo->tooltip;

        $dirfile = dirfile($objCampo->linkint, $objCampo->linkdir );
        $objCampo->linkdir = $dirfile['dir'];
        $objCampo->linkint = $dirfile['file'];

        $linkdir = (isset($this->Datos->dirxml) && $this->Datos->dirxml != '')?'&dir='.$this->Datos->dirxml:'';
        $linkdir = (isset($objCampo->linkdir) && $objCampo->linkdir != '')?'&dir='.$objCampo->linkdir : $linkdir;

        $linkint = $objCampo->linkint;
        if ($objCampo->linkint == '' && $objCampo->contExterno->xml != '') {
            $dirfile = dirfile($objCampo->contExterno->xml, $objCampo->contExterno->subdir);
            $linkint = $dirfile['file'];
           // $linkdir = '&dir='.$objCampo->contExterno->subdir;
            $linkdir = '&dir='.$dirfile['dir'];

        }

        $xmlOrig = '&_xmlreferente='.$this->Datos->xml.'&parentInstance='.$this->Datos->getInstance();
        $contetiq = 'xml='.$linkint.$params.$linkdir.$xmlOrig;

        $idobj = $objCampo->NombreCampo;
        //					$opcetiq = '\'&amp;'. $ObjNombre->NombreCampo.'=\' + if ($(\''.$ObjNombre->NombreCampo.'\')) $(\''.$ObjNombre->NombreCampo.'\').value';
        //$opcetiq = '\'\'';
        $padre = (isset($this->Datos->xmlOrig))?'DIV'.$this->Datos->xmlOrig:'DIV'.$this->Datos->xml;

        //$valorid = $valor;
        if ($valor != '') {
            //$class_link ='';
            // formateo el valor
            if ($formatoCampo != '') {
            // FORMATEO EXPLICITO

                if ($objCampo->TipoDato =='date') {
                    $valor = date($formatoCampo, strtotime($valfecha));
                } else {
                    $valor = sprintf($formatoCampo, $valor);
                }
            }
            $btnval = new Html_button($valor, $objCampo->linkIcon ,$valor );

            if ($objCampo->linkTag != '')
                $btnval->tag = $objCampo->linkTag;

    //        $btnval->addStyle('width','100%');
            $tempStyle  = (isset($objCampo->style))?$objCampo->style:'';
            $tempStyle .= (isset($objCampo->colstyle))?$objCampo->colstyle:'';


    	    if ($tempStyle != ''){
                 $btnval->addParameter('style', $tempStyle, true);
    	    }

            $btnval->addParameter('title', $title, true);

            // Link Parameters
            $btnval->addParameter('linkloader', $contetiq);

            if (isset($objCampo->linkmodal))
                    $btnval->addParameter('linkmodal', "true");

            // propagate menuid for notifications;
            $btnval->addParameter('menuid', $this->Datos->_menuId);

            if (isset($objCampo->linkPrint) && $objCampo->linkPrint=="true") {

                $btnval->addParameter('linktarget', 'print');
            } else {
                $btnval->addParameter('linkfather', $padre);

                if (isset($objCampo->linktab) && $objCampo->linktab=="true") {
                    $btnval->addParameter('linktarget', 'tab')->addParameter('linkint', 'DIV'.$linkint);

                } else {
                    $btnval->addParameter('linktarget', 'win')->addParameter('linkint', $linkint)->addParameter('linkFname', $idobj);

                    if (isset($objCampo->linkWidth))
                        $btnval->addParameter('linkwidth', $objCampo->linkWidth);

                    if (isset($objCampo->linkHeight))
                        $btnval->addParameter('linkheight', $objCampo->linkHeight);

                    if (isset($objCampo->linkReposition) && $objCampo->linkReposition == "true")
                        $btnval->addParameter('linkreposition', 'true');

                }

            }

            return $btnval->show();
        }
    }

}
