<?php
/* 
 * Loger Class 2013-04-28
 * Export class
 * @author Luis M. Melgratti
 */

class Export_ods extends Export{
    




    public function prepareFile(){

        require ("../lib/OpenOffice/classes/OpenOfficeSpreadsheet.class.php");       
        
        $this->workbook  = new OpenOfficeSpreadsheet($this->filename.'.ods', "/tmp");
        $this->worksheet = $this->workbook->addSheet(substr($this->filename,0, 30));


    }

    public function header(){
        $this->cellini = $this->worksheet->getCell(0, 0);

        $campos = $this->Container->camposaMostrar();
        $x=0;
        $cellini = $this->worksheet->getCell(0, 0);        

        foreach ($campos as $nom => $valor) {
            $Field = $this->Container->getCampo($valor);

            $Suma[$nom] = 0;
            if (isset($Field->export ) && $Field->export == 'false') continue;

            if (($Field->Oculto))
                continue;

            if ( $Field->Parametro['noshow'] == 'true')
                continue;

            $this->worksheet->setCellContent(utf8_decode($Field->Etiqueta), $x, 0);
            $cell = $this->worksheet->getCell($x, 0);
            $cell->setTextAlign('center');
            $cell->setFontWeight('bold');
            $this->worksheet->setCellColor('#999933', $cell, $cell);
            $x++;
        }



    }

    public function footer(){
        if (is_array($this->totals)){

            //print_r($this->totals);
            $y = $this->lastY;
            $y++;
            foreach($this->totals as $x => $value){
                        $this->worksheet->setCellContent(utf8_decode($value), $x, $y);

            }
        }

        $this->worksheet->setCellBorder('0.02cm solid #000000', $this->cellini, $this->cellfinal);
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


        $this->totals[$x]=$Suma[$Field->NombreCampo];

        switch ($Field->TipoDato) {
            case "numeric" :
                $valorxls = $valor;
            break;
            /*
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
            */
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
        /*
        $formathash = serialize($cell_format);
        
        if (!isset( $this->formatos[$formathash] )){
            $format = $this->workbook->addformat($cell_format);

                $this->formatos[$formathash]= $format;
        }
        */
        $this->worksheet->setCellContent(utf8_decode($valorxls), $x, $y);
        $this->cellfinal = $this->worksheet->getCell($x, $y);
    }

    public function sendHeaders(){
        header("Content-Type:  application/x-vnd.oasis.opendocument.spreadsheet; name=\"".$this->filename.".ods\"");
        header("Content-Disposition: inline; filename=\"".$this->filename.".ods\"");
    }

    public function out(){
        $this->workbook->output();
    }


}
?>
