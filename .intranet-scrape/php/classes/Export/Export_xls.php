<?php
/* 
 * Loger Class 2013-04-28
 * Export class
 * @author Luis M. Melgratti
 */

class Export_xls extends Export{
    




    public function prepareFile(){
        require "../lib/phpxls/class.writeexcel_workbook.inc.php";
        require "../lib/phpxls/class.writeexcel_worksheet.inc.php";
        require "../lib/phpxls/functions.writeexcel_utility.inc.php";        
        // Generacion del Excel
        $this->fname = tempnam("/tmp", $this->Container->idxml.'.xls');
        $this->workbook = new writeexcel_workbook($this->fname);
        $this->workbook->set_tempdir('/tmp');
        $this->worksheet = $this->workbook->addworksheet(substr($this->Container,0, 30));


    }

    public function header(){

        //# Create a format for the column headings
        $header = $this->workbook->addformat();
        $header->set_bold();
        $header->set_size(12);
        $header->set_color('blue');

        /*
        $odd = $this->workbook->addformat();
        $odd->set_bg_color('silver');
        */
        $campos = $this->Container->camposaMostrar();
        $x=0;
        foreach ($campos as $nom => $valor) {
            $Field = $this->Container->getCampo($valor);

            $Suma[$nom] = 0;
            if (isset($Field->export ) && $Field->export == 'false') continue;

            if (($Field->Oculto))
                continue;

            if ( $Field->Parametro['noshow'] == 'true')
                continue;

            $this->worksheet->write(0, $x, utf8_decode($Field->Etiqueta), $header);
            $x++;
        }
    }

    public function footer(){
        if (is_array($this->totals)){
            //# Create a format for the column headings
                    
            $footer = $this->workbook->addformat();
            $footer->set_bold();
            $footer->set_size(12);
            $footer->set_color('green');
            //print_r($this->totals);
            $y = $this->lastY;
            $y++;
            foreach($this->totals as $x => $value){
                        $this->worksheet->write($y, $x, utf8_decode($value), $footer);

            }
        }
    }


    public function processData($row, $Field, $params){

        $x = $params['x'];
        $y = $params['y'];
        $this->lastY= $y;
        $Valcampo = $params['value'];
        $nomcampo = $params['fieldname'];

        if (count($Field->opcion) > 0 && $Field->TipoDato != "check" && $Field->valop != 'true') {
            $valor = $Field->opcion[$Valcampo];
            if (is_array($valor)) $Valcampo = current($valor);
        }
        $valor = $Valcampo;

        if ($this->Container->seSuma($nomcampo) && $norepeat != true) {
            $Suma[$Field->NombreCampo] += $valor;
            $Subtotal[$Field->NombreCampo] += $valor;
        }

        if ($this->Container->seAcumula($nomcampo) && $norepeat != true) {
            $valor = $Suma[$Field->NombreCampo];
            $Field->TipoDato = "numeric";
        }

        $valorxls= $valor;


        $cell_format = array();

	if (isset($Suma[$Field->NombreCampo]))
        $this->totals[$x] +=$Suma[$Field->NombreCampo];

        switch ($Field->TipoDato) {
            case "numeric" :
                $valorxls = $valor;
            break;
            case "date":
                $cell_format = $date_format;
                $cell_format['num_format']='dd/mm/yyyy';
                $cell_format['align']='center';
                $valorxls = xl_parse_date($valor);

            break;
            case "time":
                $cell_format['num_format']='HH:MM:SS';
                $cell_format['align']='center';
                $valorxls = xl_parse_time($valor);
                $columnFormat[$Field->NombreCampo]=$cell_format;
                if ($this->Container->seSuma($nomcampo) || $Field->suma == 'true')
                    $this->totals[$x]='=SUM('.xl_rowcol_to_cell(1,$x).':'.xl_rowcol_to_cell($y, $x).')';

            break;
            default:
                $valorxls = utf8_decode($valor);
            break;
        }                                                                                                                                                                                                                                                                                                                                                                                                                       // Si tiene opciones de un combo


        if ($norepeat!=true){
            if ($y % 2 != 0) 
                $cell_format['bg_color']='silver';
        }

        if (is_object($valorxls)){
            $valorxls = '';
        }   
        // free some memory
        unset($Field);
        $formathash = serialize($cell_format);

        if (!isset( $this->formatos[$formathash] )){
            $format = $this->workbook->addformat($cell_format);

                $this->formatos[$formathash]= $format;
        }
        
        $this->worksheet->write($y, $x, $valorxls, $this->formatos[$formathash]);
        
        // free memory
        unset($cell_format);
        unset($valorxls);
    }

    public function sendHeaders(){
        header("Content-Type: application/x-msexcel; name=\"".$this->filename.".xls\"");
        header("Content-Disposition: inline; filename=\"".$this->filename.".xls\"");
    }

    public function out(){
        $this->workbook->close();
        $fh=fopen($this->fname, "rb");
        fpassthru($fh);
        unlink($fname);
    }


}
?>
