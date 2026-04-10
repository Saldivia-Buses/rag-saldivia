<?php
/* 
 * Dot Matrix Printer 
*/

class Histrix_dotPrinter {



    function __construct($orientation, $units, $pageSize) {

        $this->orientation = $orientation;
        $this->units = $units;
        $this->pageNumber = 1;

        $this->pageWidth = 120;
        $this->pageLength = 66;
        $this->pageSize = $pageSize;


        $this->setCodes();

        switch($this->pageSize) {
            case 'A4':
                $this->pageLength = 62;
                break;
        }


        $db = $_SESSION["db"];
        $datosbase = Cache::getCache('datosbase'.$db);
        if ($datosbase === false) {
            $config = new config('config.xml', '../database/', $db);
            $datosbase = $config->bases[$db];
            Cache::setCache('datosbase'.$db, $datosbase);
        }

        $this->datosbase = $datosbase;
        $this->datapath    = '../database/'.$datosbase->xmlPath;
        $this->tmpDir       = $this->datapath.'/tmp/' ;
        $this->html= '';
    }

    private function setCodes() {
        // Epson Escape Codes
        $this->code['paperFeed'] = chr(12);
        $this->code['clearAll'] = chr(27).chr(64);
        $this->code['charsetNormal'] =chr(27).chr(37).chr(0);
        $this->code['bold']['on']  = chr(27).chr(69);
        $this->code['bold']['off'] = chr(27).chr(70);

        $this->code['italic']['on']  = chr(27).chr(52);
        $this->code['italic']['off'] = chr(27).chr(53);

        $this->code['enlarge']['on']  = chr(27).chr(87).'1';
        $this->code['enlarge']['off'] = chr(27).chr(87).'0';

        $this->code['condensed']['on']  = chr(15);
        $this->code['condensed']['off'] = chr(18);

    }

// dummy function (yet)?
    public function SetCompression($bool) {

    }

    public function setTitle($title) {
        $this->title = $title;
    }
    public function setCreator($creator) {
        $this->creator = $author;
    }
    public function setAuthor($author) {
        $this->author = $author;
    }

    public function SetAutoPageBreak($auto , $margin) {
        $this->autoPageBreak = $auto;
        $this->bottomMargin = $margin;
    }

    public function AliasNbPages() {

    }
    public function AddPage() {
// todo ad page
        $this->html .= $this->Header($this->pageNumber);
    }



    private function Header($pageNumber) {

        $header['client'] = '       '.$this->datosbase->nombre.'         ';
        $header['title']  = '           '.strtoupper($this->title).'          ';

        if (!$this->sinFecha)
            $header['date'] = '   Fecha Imp:' . date('d/m/Y').'   ';
        if (!$this->sinNumero){
            //$header['Page'] = '                  '.$pageNumber.' ';
            $header['Page'] = '                      '.str_pad($pageNumber, 8, '0', STR_PAD_LEFT);
        }

        $data =  array($header);
        $renderer = new ArrayToTextTable($data);


        $renderer->showHeaders(false);
        $renderer->setOuterBorder('', '-');
        $renderer->setCenterCharacter('');

        $this->line = 1;
//        $header  = $this->code['clearAll'];

        $header = $renderer->render(true)."\n";
        return $header;

    }

    private function textTable($data, $hasTotals) {
        $renderer = new ArrayToTextTable($data);
        $renderer->setTotals(true);
        $renderer->setCenterCharacter( '');
        $renderer->setOuterBorder('', '=');
        $renderer->setInnerBorder(' ', '=');
        $renderer->showHeaders(true);
        return $renderer->render(true);

    }


    public function SetFont($font,$style,$size) {
        $this->fontType   = $font;
        $this->fontStyle  = $style;
        $this->fontSize   = $size;
    }

    public function SetX($X) {

    }
    public function SetY($Y) {

    }

    public function GetX() {

    }
    public function GetY() {

    }

// histrix part :)
    public function impAbm($Container) {
        $this->footer= $Container->txtFooter;
        $this->header= $Container->txtHeader;

        if ($Container->nosql != 'true' /*&& $accion != 'insert'*/)
            $Container->Select();

        $UI = 'UI_'.str_replace('-', '', $Container->tipo);
        $abmDatos = new $UI($Container);

        $abmDatos->muestraCant = 'false';
        $this->html .= $this->html2text($abmDatos->showAbmInt('readonly', 'INT'.$Container->xml));
    }

    /**
     * get html simple representation to print
     * @param <type> $Container
     * @param <type> $n1
     * @param <type> $n2
     * @param <type> $n3
     * @param <type> $fontsize
     */
    public function impTabla($Container, $n1, $n2, $n3, $fontsize) {
        $UI = 'UI_'.str_replace('-', '', $Container->tipo);
        $abmDatos = new $UI($Container);

        $this->footer= $Container->txtFooter;
        $this->header= $Container->txtHeader;
        $this->cols  = $Container->txtCols;

        $abmDatos->fillTextArray = true;
        $abmDatos->showTablaInt('txt', '', $act, 'true', null, null, null, $Container);

        $data = $abmDatos->textArray;
        $this->html .= $this->textTable($data, $ambDatos->hasTotals);



    }


    public function impFiltros($Container) {
        $this->footer= $Container->txtFooter;
        $this->header= $Container->txtHeader;
        $this->cols  = $Container->txtCols;

        if (isset ($Container->filtros)) {
//  $this->html[]='<table>';

            foreach ($Container->filtros as $nomfiltro => $objFiltro) {

                if ($objFiltro->print == 'false') continue;

                $CampoFiltro = $Container->getCampo($objFiltro->campo);
                if ($CampoFiltro == null)
                    continue;
                $valor = $objFiltro->valor;
                if ($CampoFiltro->opcion != '') {
                    $valor = $CampoFiltro->opcion[$valor];
                    if ($valor == '') {
                        $valor = current($CampoFiltro->opcion);
                    }
                    if (is_array($valor))
                        $valor = current($valor);
                }

                $label = utf8_decode($objFiltro->label);

                if( $valor == '')
                    $valor = $CampoFiltro->valor;

                if ($objFiltro->modpos != 'nobr') {
                    $i++;
                }

                $lines[$i] .= $label.'  '.$valor.' ';
            }

            $this->html .="\n";
            foreach($lines as $line) {

                $this->html .= $line."\n";
            }


            $this->html .="\n";

        }

    }



    public function setAnchoCol($Array) {

    }

    public function WriteTable($Table, $width) {

    }

    public function showTree($Container, $tree, $n1, $n2 , $bool) {

    }

    public function WriteHTML($htmlTable) {

    }

    public function showGraficos($Container) {

    }


    /**
     * PAginate Content
     * @param string $text
     * @return string
     */
    private function paginateText($text) {
        $arrayIN = explode("\n",$text);
        foreach($arrayIN as $lineNum => $row) {
            $currentLine++;
            $array[]= $row;

            if ($currentLine == $this->pageLength) {
                $this->pageNumber++;
                $array[]= "\n".$this->code['paperFeed'];
                $headerArray = explode("\n",$this->Header($this->pageNumber));
                array_splice($array, count($array), 0, $headerArray);
//array_push($array, $headerArray);
                $currentLine = 0;
            }
        }

        $text = implode("\n", $array);
        
        return $text;
    }

    public function Output($fileName, $type) {
        $text = $this->paginateText($this->html);

        // Add custom Codes
        $text .= $this->code['pageFeed'];
        //$text .= $this->code['charsetNormal'];
    //    $text .= $this->code['bold']['on'];

/*
        echo '<pre>';
        print_r($text);
        echo '</pre>';
        die();
        */
        $this->txt = $text;
        $fh=fopen($fileName, "w+");
        fwrite($fh, $this->txt);
        $this->addTemplates($fileName);

//   unlink($fname);

    }

    public function Line($x, $y, $w, $h) {

    }

    public function html2text($content) {
        $outputFile = $this->tmpDir.uniqid('html');

        $fh=fopen($outputFile, "w+");
        fwrite($fh, $content);

        $out = ' > "'. $outputFile.'_2"';

        $cols = ($this->cols != '')? $this->cols: 77;

//  exec("html2text $file ".$out, $salida);
        exec("cat '$outputFile' | w3m -cols $cols -dump -T text/html ".$out, $salida);


        $salida = file_get_contents($outputFile.'_2');
        unlink($outputFile.'_2');
        unlink($outputFile);

        return $salida;
    }

    private function addTemplates($outputFile) {
        if ($this->header != '') {
            $header = '\''.$this->datapath.'/tpl/'.$this->header.'\'';
        }
        if ($this->footer != '') {
            $footer = '\''.$this->datapath.'/tpl/'.$this->footer.'\'';
        }
        exec("mv '$outputFile'  '".$outputFile."_2' ");
        exec("cat $header '".$outputFile."_2' $footer > '$outputFile' ");
        unlink($outputFile.'_2');
    }



}
?>