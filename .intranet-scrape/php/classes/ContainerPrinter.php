<?php
/* 
 * Container Printer
 * Print a Data container
 * 2009-09-19 - Luis M. Melgratti
 *
*/
define('FPDF_FONTPATH','../lib/fpdf/font/');
require_once ('../lib/fpdf/fpdf.php');

//define('FPDF_FONTPATH','../lib/tcpdf/font/');
//require_once ('../lib/tcpdf/tcpdf.php');

require_once ('./htmlparser.inc.php');

class containerPrinter {

    public function __construct($Container, $parameters= null, $outputFormat = 'pdf' ) {
        $this->localecho = 0;
        $this->Container    = $Container;
        if ($Container->copies != '')
            $this->copies = $Container->copies;
        $this->parameters   = $parameters;
        $this->outputFormat = $outputFormat;

        $this->pageSize    = (isset($Container->PDFpageSize ))    ? $Container->PDFpageSize: 'A4';
        $this->fontsize    = (isset($Container->PDFfontSize ))    ? $Container->PDFfontSize: 8;
        $this->defaultFont   = (isset($Container->PDFfont))  ? $Container->PDFfont: 'helvetica';

        $this->orientation = (isset($Container->PDForientation )) ? $Container->PDForientation : 'P';
        $this->orientation = (isset($Container->PDForientacion )) ? $Container->PDForientacion : $this->orientation;

        if (isset($Container->PDFpageSizeX ) && $Container->PDFpageSizeY !='')
            $this->pageSize = array($Container->PDFpageSizeX, $Container->PDFpageSizeY);

        if (isset($Container->printer))
            $this->printer = $Container->printer;
    }


    public function setParameters($parameters = null) {
        $this->parameters = $parameters;
    }

    public function printContainer() {

        $this->hideFields();

        // Set type of printer text or pdf
        if (isseT($this->printer))
            $this->setPrinterData($this->printer);
        // get class name to use
        $outputFormat = 'Histrix_'.$this->outputFormat;

        $pdf = new $outputFormat($this->orientation, 'mm', $this->pageSize);

        $pdf->Container = $this->Container;
        $pdf->SetCompression(true);

        // Template method
        if (isset($this->Container->template)) {

            if (is_object($this->Container))
                $pdf = $this->printTemplate();
            
        }
        else {

            if (isset($this->Container->sinFecha))
                $pdf->sinFecha  = $this->Container->sinFecha;
            if (isset($this->Container->sinNumero))
                $pdf->sinNumero = $this->Container->sinNumero;

            if (isset($this->Container->pageNumber))
                $pdf->pageNumber = $this->Container->getCampo($this->Container->pageNumber)->ultimo ;


            $titulo = strip_tags(($this->Container->titulo_div != '')? utf8_decode($this->Container->titulo_div) : ( is_object($this->Container)?(utf8_decode($this->Container->getTitulo())):''));

            if (isset($this->Container->PDFsincabecera) && $this->Container->PDFsincabecera == true)
                $pdf->sincab = true;

            if (isset($this->Container->PDFconpie) && $this->Container->PDFconpie == true)
                $pdf->conpie = true;


            if (is_object($this->Container)){
                $UI = 'UI_'.str_replace('-', '', $this->Container->tipo);
                $UIContainer = new $UI($this->Container);

                if (isset($this->Container->CabeceraMov)){
                    $contCabecera = $this->Container->CabeceraMov;
                    if (isset ($contCabecera)) {
                        foreach ($contCabecera as $NCabecera => $ContCab) {
                            $titulo = strip_tags(utf8_decode($ContCab->getTitulo()));
                        }
                    }
                }
            }
            /* Generate footer*/
            $pdf->pie['text'] = $this->generateFooter($this->Container, $pdf);

            // INITIALIZE pdf
            if ( isset($_SESSION['properties']['supportName']) )
                $creator = $_SESSION['properties']['supportName'];
            else
                $creator = 'Histrix';

            $pdf->titulo = $titulo;
            $pdf->setTitle($titulo);
            $pdf->SetCreator( $creator );
            $pdf->SetAuthor($_SESSION['usuario']);

            $pdf->SetAutoPageBreak(true , 2);
            $pdf->bottomMargin=-12;
              $pdf->bottomMargin=50;
            if (isset($this->Container->bottomMargin))
                $pdf->bottomMargin = $this->Container->bottomMargin;


            if ($this->headerFields != ''){
                $pdf->headerFields = $this->headerFields;
                $pdf->Container =  $this->Container ;
            }
            if ($this->footerFields != ''){
                $pdf->footerFields = $this->footerFields;
                $pdf->Container =  $this->Container ;
	    }

            $pdf->AliasNbPages();
            $pdf->AddPage();
            $pdf->fontsize = $this->fontsize;
            $pdf->defaultFont = $this->defaultFont;
            $pdf->SetFont($this->defaultFont, '', $this->fontsize);

            // custom pdf execution
            if (isset($this->Container->pdfCode)) {
                $code = $this->Container->pdfCode;
                if ($code != '') {
                    eval($code);
                }
            }

            // PRINT header
            if (isset ($contCabecera)) {
                foreach ($contCabecera as $NCabecera => $ContCab) {
                    $tablacab[$NCabecera] = $pdf->impAbm($ContCab);
                    $wcab = $pdf->setAnchoCol($tablacab[$NCabecera]);
                    $pdf->SetY($pdf->maxY);
                    $pdf->SetY($pdf->GetY() + 4);
                    $pdf->WriteTable($tablacab[$NCabecera], $wcab);
                }
            }
            // Print filters
            if (isset ($this->Container->filtros) || isset ($this->Container->filtroPrincipal)) {
                $pdf->SetY($pdf->maxY );
                $pdf->impFiltros($this->Container);
            }

            // reposition
            $pdf->SetY($pdf->GetY() + 2);

            // Print Main content            
            if (isset($UIContainer))
                $UIContainer->pdf($pdf, $this->fontsize);
            
          
        //    $this->printContent($pdf);


            $pdf->showGraficos($this->Container);
        }

        $this->output($pdf, $this->target, $titulo);

    }

    public function hideFields($parameters = '') {
        $parameters = ($parameters != '')? $parameters: $this->parameters;

        if (is_object($this->Container))
            $campos = $this->Container->camposaMostrar();
        else return;

        foreach ($campos as $num => $valor) {
            if (isset($parameters['__'.$valor])) {
                $fieldName = $parameters['__'.$valor];
                if ($fieldName != '') {

                    $field = $this->Container->getCampo($valor);
                    if ($field != '') {
                        $field->print  = 'true';
                        if ($fieldName == 0)
                            $field->print  = 'false';
                    }
                }
            }
            else {
                $field = $this->Container->getCampo($valor);
                if (isset($field->PDFheader) && $field->PDFheader =="true") {
                    $field->print='false';
                    $this->headerFields[] = $field;
                }
                if (isset($field->PDFfooter) && $field->PDFfooter =="true") {
                    $field->print='false';
                    $this->footerFields[] = $field;
                }
                
            }
        }
    }

    /*
     * Output to apropiate Medium
    */

    public function output($pdf, $dest, $titulo) {

        $MisDatos = $this->Container;
        switch($dest) {
            case 'mail':
                $dir = '../database/'.$_SESSION['datapath'].'tmp/';

//                $fname = tempnam($dir, $titulo.'_');
	          $fname = $dir.utf8_encode($titulo);
                $fileName = $fname.'.'.$this->outputFormat;
                
                

                if (isset($pdf->attachedFiles)){
                    $_GET['filesArray']= urlencode(serialize($pdf->attachedFiles));
    //                print_r($_GET['filesArray']);
                }
                $pdf->Output($fileName, 'F');

                $fileName = basename($fileName);
                if (is_object($MisDatos)) {
                    $toField = $MisDatos->getFieldByAttribute(array('email'=> 'to'));
                    if ($toField != false){
                        $_GET['para']= $toField->getValor();
                    }


                    $toField = $MisDatos->getFieldByAttribute(array('email'=> 'subject'));
                    if ($toField != false){
                        $_GET['subject']= $toField->getValor();
                    }

                    $toField = $MisDatos->getFieldByAttribute(array('email'=> 'body'));
                    if ($toField != false){
                        $_GET['mensaje']= $toField->getValor();
                    }


                    $_GET['adjunto']=$fileName;
                }
                $_GET['dir']= 'histrix/mensajeria';
                $_GET['xml']= 'mails_send.xml';

                include "histrixLoader.php";

                break;
            case 'pdf':
                                            
                $pdf->Output(urlencode($titulo . '.pdf'), 'I');
                break;
            case 'printer':

                $dataPath = $_SESSION['datapath'];
                if ($dataPath != '') {
                    $tmpbase= '../database/'.$dataPath;
                }

                $titulo = utf8_encode($titulo);
                $fileName = $tmpbase."tmp/".uniqid($titulo).'.'.$this->outputFormat;

                $salida =  $pdf->Output($fileName, 'F');

                $copies=($this->copies != '')? $this->copies : 1;

                if ($this->outputFormat == 'txt' ||
                        $this->outputFormat == 'dotPrinter') {

                    // RAW TXT print
                    // Linux Only Direct Print
                    // window$ version?? don't think so...
                    //$exec = 'lpr -# '.$copies.' -P'.$this->printerName.' "'.$fileName.'"';
                    $exec = 'lp -n'.$copies.' -d'.$this->printerName.' "'.$fileName.'"  2>/tmp/error.print';
                    if ($this->localecho != 1) {
                        
                        $outputText = file_get_contents($fileName);
                        echo '<pre>'.$fileName.$outputText.'</pre>';
                        
                    }
                    else {
                        $header = 'Imresion - '.date('H:i:s', time());
                        $text   = 'Se envio '.$titulo.' a '.$this->printerName;
                        $code[] = "Histrix.notification('Print', {icon:'printer1.png',title:'$header',text:'$text',fade:4000 })";
                        echo Html::scriptTag($code);
                    }


                }
                else {
                    // PDF direct Print
                    if($MisDatos->PDFpageSizeX != '' && $MisDatos->PDFpageSizeY != '') {
                        $lprMedia = ' -o media=Custom.'.$MisDatos->PDFpageSizeX.'x'.$MisDatos->PDFpageSizeY.'mm';
                        $inchesX = $MisDatos->PDFpageSizeX * 0.0393700787;
                        $inchesY = $MisDatos->PDFpageSizeY * 0.0393700787;
                        $gsDEVICEHEIGHTPOINTS = ' -dDEVICEHEIGHTPOINTS='. round($inchesY * 72, 0);
                        $gsDEVICEWIDTHPOINTS  = ' -dDEVICEWIDTHPOINTS=' . round($inchesX * 72, 0);
                        ;

                    }

                    if($MisDatos->gsDEVICEHEIGHTPOINTS != '') {
                        $gsDEVICEHEIGHTPOINTS=' -dDEVICEHEIGHTPOINTS='.$MisDatos->gsDEVICEHEIGHTPOINTS;
                        $lprMedia = ' -o media=Custom.'.$MisDatos->gsDEVICEWIDTHPOINTS.'x'.$MisDatos->gsDEVICEHEIGHTPOINTS;
                    }

                    if($MisDatos->gsDEVICEWIDTHPOINTS != '') {
                        $gsDEVICEWIDTHPOINTS='-dDEVICEWIDTHPOINTS='.$MisDatos->gsDEVICEWIDTHPOINTS;
                        $lprMedia = ' -o media=Custom.'.$MisDatos->gsDEVICEWIDTHPOINTS.'x'.$MisDatos->gsDEVICEHEIGHTPOINTS;
                    }

                    if($MisDatos->gsPAPERSIZE != '') {
                        $gsPAPERSIZE=' -sPAPERSIZE='.$MisDatos->gsPAPERSIZE;
                        $lprMedia = ' -o media='.$MisDatos->gsPAPERSIZE;
                    }

                    if ($this->printerClass != '') {

                        /*                        $exec = 'gs  -q -dBATCH -dNOPAUSE -dQUIET -r60x72 '.$gsPAPERSIZE.' '.$gsDEVICEHEIGHTPOINTS.' '.$gsDEVICEWIDTHPOINTS.
                            ' -sDEVICE='.$this->printerClass.' -dNOPLATFONTS -sOutputFile=%pipe%"lpr -l -# '.
                            $copies.' '.$lprMedia.' -P'.$this->printerName.'" "'.$fileName.'"'; */
                        $exec = 'gs  -q -dBATCH -dNOPAUSE -dQUIET -r60x72 '.$gsPAPERSIZE.' '.$gsDEVICEHEIGHTPOINTS.' '.$gsDEVICEWIDTHPOINTS.
                                ' -sDEVICE='.$this->printerClass.' -dNOPLATFONTS -sOutputFile=%pipe%"lp -l -n'.
                                $copies.'  -d'.$this->printerName.'" -o raw "'.$fileName.'"';
                    }
                    else {
                        //$exec = 'lpr -# '.$copies.' -P'.$this->printerName.' "'.$fileName.'"';
                        $exec = 'lp -n'.$copies.' -d'.$this->printerName.' "'.$fileName.'" 2>/tmp/error.print';

                    }
                }

                loger($exec, 'print.log');
                $sal = shell_exec($exec);

              //  unlink($fileName); // remove Temp ??

                break;

        }
    }


    /*
     * generate String Footer
    */

    public function generateFooter($MisDatos, $pdf) {
        if (isset($MisDatos->pie) && $MisDatos->pie != '') {
            $pdf->pie = $MisDatos->pie;
            if (strpos($pdf->pie['text'], '[__') !== false && strpos($pdf->pie['text'], '__]') !== false ) {
                foreach ($MisDatos->tablas[$MisDatos->TablaBase]->campos as $MiNro => $Items) {
                    $matrizNombres[] = '[__' . $Items->NombreCampo . '__]';

                    if (isset ($Items->suma))
                        $valorAcumulado = $Items->Suma;
                    else
                        $valorAcumulado = $Items->ultimo;

                    if (trim($valorAcumulado) =='') {
                        $valorAcumulado = $Items->valor;
                    }

                    $matrizValores[] = $valorAcumulado;
                }
                $str = str_replace($matrizNombres, $matrizValores, $pdf->pie['text']);
            }
            else $str = $pdf->pie['text'];

            
            return $str;
        }
        
    }

    /**
     * Get especifications for printer
     * @param string $printer
     */
    public function setPrinterData($printer) {
        //GET PRINTER Data
        if ($printer != '') {
            $sql = 'select * from HTXPRINTERS where idPrinter = "'.$printer.'" limit 1';
            $rs = @consulta($sql, null, 'nolog');
            if ($rs)
                while ($row = _fetch_array($rs)) {
                    $printerClass = $row['printerClass'];
                    if ($row['outputFormat'] != '')
                        $outputFormat = $row['outputFormat'];
                    $systemName = $row['systemName'];
                    if ($systemName == '')  $systemName = $printer;

                }

            $this->printerClass = $printerClass;
            $this->printerName  = $systemName;
            if ($outputFormat != '')
                $this->outputFormat = $outputFormat;

        }
    }

    public function printTemplate() {

        $MisDatos = $this->Container;

        // se this... mmmm
        $dataPath = $_SESSION['datapath'];
        $tplfile = '../database/'.$dataPath.'/tpl/'.$MisDatos->template;

        $Template=new Histrix_Template($tplfile);
        foreach ($MisDatos->tablas[$MisDatos->TablaBase]->campos as $MiNro => $Items) {
            $valor = $Items->valor;

            if ($Items->TipoDato == 'date' && $valor != '')
                $valor = date("d/m/Y", strtotime($valor));
            if ($Items->opcion != '') {
                if (is_array($Items->opcion[$valor]))
                    $valor = current($Items->opcion[$valor]);
                else $valor = $Items->opcion[$valor];
            }

            if ($Items->aletras == true) {
                if (is_numeric($valor))
                    $valor = NumeroALetras($valor);
            }

            //$valor = utf8_decode($valor);
            //if ($Items->TipoDato == 'decimal')
            //	$valor = number_format($valor, 2, '.', ',');

            if (isset ($Items->contExterno) && $Items->obj != '') {

                $valor = $Items->contExterno->TablaTemporal->datos();
            }

            $campos_a_Mostrar[$Items->NombreCampo] = $valor;

        }

        $Template->asignData($campos_a_Mostrar);

        $Template->parse();

        $this->target='printer';
        $this->outputFormat = 'txt';

        return $Template;
    }


}

?>
