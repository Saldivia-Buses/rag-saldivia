<?php
/**
 * Generic Text Printer
 * @author Luis M. Melgratti
 */

class Histrix_txt {

    function __construct($orientation, $units, $pageSize){
        $this->orientation = $orientation;
        $this->units = $units;
        $this->pageSize = $pageSize;
//        $this->datosbase = $_SESSION['datosbase'];

        $db = $_SESSION["db"];
        if ($db != '') {
            $datosbase = Cache::getCache('datosbase'.$db);
            if ($datosbase === false){
                $config = new config('config.xml', '../database/', $db);
                $datosbase = $config->bases[$db];
                Cache::setCache('datosbase'.$db, $datosbase);
            }

            $this->datosbase = $datosbase;
            $this->datapath    = '../database/'.$datosbase->xmlPath;
        }
        $this->html= '';
    }


    // dummy function (yet)?
    public function SetCompression($bool){}

    public function setTitle($title){
        $this->title = $title;
    }
    public function setCreator($creator){
        $this->creator = $author;
    }
    public function setAuthor($author){
        $this->author = $author;
    }

    public function SetAutoPageBreak($auto , $margin){
        $this->autoPageBreak = $auto;
        $this->bottomMargin = $margin;
    }

    public function AliasNbPages(){}
    public function AddPage(){
        // todo ad page
        $this->html[] = $this->Header();
    }

	function Header() {
        
        $salida .= '<table border="1">';
        $salida .= '<tr>';
        $salida .= '<th width="20%">'. $this->datosbase->nombre.'</th>';
        $salida .= '<th>'. $this->title.'</th>';
        $salida .= '<th width="40%">'.'Fecha Imp:' . date('d/m/Y').'</th>';
        $salida .= '</tr>';
        $salida .= '</table>';
        $salida .= '<br/>';
        return $salida;
    }


    public function SetFont($font,$style,$size){
        $this->fontType   = $font;
        $this->fontStyle  = $style;
        $this->fontSize   = $size;
    }

    public function SetX($X){}
    public function SetY($Y){}

    public function GetX(){}
    public function GetY(){}

    // histrix part :)
    public function impAbm($Container){
        $this->footer= $Container->txtFooter;
        $this->header= $Container->txtHeader;

        if ($Container->nosql != 'true' /*&& $accion != 'insert'*/)
            $Container->Select();
            
    	$UI = 'UI_'.str_replace('-', '', $Container->tipo);
    	$abmDatos = new $UI($Container);
            
        $abmDatos->muestraCant = 'false';
        $this->html[] = $abmDatos->showAbmInt('readonly', 'INT'.$Container->xml);
    }

    public function impTabla($Container, $n1, $n2, $n3, $fontsize){
        $UI = 'UI_'.str_replace('-', '', $Container->tipo);
	$abmDatos = new $UI($Container);
    
        $this->footer= $Container->txtFooter;
        $this->header= $Container->txtHeader;
        $this->cols  = $Container->txtCols;
        
        $this->html[] = $abmDatos->showTablaInt($opt, '', $act, 'true', null, null, null, $Container);
    }

    public function impFiltros($Container){
        $this->footer= $Container->txtFooter;
        $this->header= $Container->txtHeader;
        $this->cols  = $Container->txtCols;

        if (isset ($Container->filtros)){
            $this->html[]='<table>';

            foreach ($Container->filtros as $nomfiltro => $objFiltro) {
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


                $this->html[]='<tr><td>'.$label.'</td><td>'.$valor.'</td></tr>';

            }
            $this->html[]='</table>';
        }

    }



    /*
    public function impFiltros($Container){
        $this->footer= $Container->txtFooter;
        $this->header= $Container->txtHeader;
        $this->cols  = $Container->txtCols;

        $UI = 'UI_'.str_replace('-', '', $Container->tipo);
	$abmDatos = new $UI($Container);
        
        $this->html[] = $abmDatos->showFiltrosXML('', false);
    }
*/


    

    public function setAnchoCol($Array){}

    public function WriteTable($Table, $width){}

    public function showTree($Container, $tree, $n1, $n2 , $bool){}

    public function WriteHTML($htmlTable){}

    public function showGraficos($Container){}

    public function htmlHeader($title){
        $html = '<head>';
        $html .= '<title>'.$title.'</title>';
        $html .= '</head>';
       	$cssFile = '../../../css/histrix.css';

        $html .= '<link rel="stylesheet" type="text/css" href="'.$cssFile.'" />';
        return $html;
    }

    public function Output($fileName, $type){
        $html = $this->html;
        if ($html != ''){
    	    $tempname = uniqid('txt');
            $fname = $this->datapath.'/tmp/'.$tempname.'.html';
            $fh=fopen($fname, "w+");
            fwrite($fh, '<html>');
            $header = $this->htmlHeader($fileName);
            fwrite($fh, $header);

            foreach($html as $order => $htcode){
                fwrite($fh, $htcode);
            }
            fwrite($fh, '</html>');
            fclose($fh);
        }

        $this->txt =  $this->html2text($fname, $fileName);

        echo '<pre>'.$this->txt.'</pre>';

        unlink($fname);

    }

    public function Line($x, $y, $w, $h){

    }


    public function html2text($file, $outputFile=''){


        if ($outputFile != ''){
            $out = ' > "'. $outputFile.'_2"';

            $cols = ($this->cols != '')? $this->cols: 77;

     //    exec("html2text $file ".$out, $salida);
          exec("cat '$file' | w3m -cols $cols -dump -T text/html ".$out, $salida);

          if ($this->header != ''){
              $header = '\''.$this->datapath.'/tpl/'.$this->header.'\'';
          }
          if ($this->footer != ''){
              $footer = '\''.$this->datapath.'/tpl/'.$this->footer.'\'';
          }

          exec("cat $header '$outputFile'_2 $footer > '$outputFile'");
          unlink($outputFile.'_2');

          $salida = file_get_contents($outputFile);
        }
        return $salida;
    }

}
?>